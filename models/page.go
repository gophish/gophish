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
	log.Debug("Starting parseHTML for page:", p.Name)
	d, err := goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	if err != nil {
		log.Error("Error parsing HTML:", err)
		return err
	}

	forms := d.Find("form")
	log.Debug("Found", forms.Length(), "forms in the page")
	
	forms.Each(func(i int, f *goquery.Selection) {
		log.Debug("Processing form", i+1)
		
		// We always want the submitted events to be sent to our server
		f.SetAttr("action", "")
		f.SetAttr("method", "POST")
		log.Debug("Set form action to empty string and method to POST")
		
		if p.CaptureCredentials {
			log.Debug("Credential capture is enabled")
			inputs := f.Find("input")
			log.Debug("Found", inputs.Length(), "input fields")
			
			// Add the rid parameter as a hidden input
			f.PrependHtml(`<input type="hidden" name="rid" value="{{.RId}}"/>`)
			log.Debug("Added hidden RId field")
			
			inputs.Each(func(j int, input *goquery.Selection) {
				inputType, _ := input.Attr("type")
				inputName, hasName := input.Attr("name")
				placeholder, _ := input.Attr("placeholder")
				id, _ := input.Attr("id")
				
				log.Debug("Processing input field:", j+1, "Type:", inputType, "Name:", inputName, "HasName:", hasName, "Placeholder:", placeholder, "ID:", id)
				
				// Always try to assign a name if we don't have one
				if !hasName {
					switch strings.ToLower(inputType) {
					case "text", "email", "tel":
						// First try to identify by placeholder or id
						fieldText := strings.ToLower(placeholder + " " + id)
						if strings.Contains(fieldText, "email") || 
						   strings.Contains(fieldText, "username") || 
						   strings.Contains(fieldText, "user") || 
						   strings.Contains(fieldText, "phone") || 
						   strings.Contains(fieldText, "mobile") || 
						   strings.Contains(fieldText, "skype") || 
						   strings.Contains(fieldText, "login") {
							input.SetAttr("name", "username")
							log.Debug("Assigned name 'username' to input based on field text:", fieldText)
						} else {
							// Fallback - if it's the first text/email input, assume it's username
							input.SetAttr("name", "username")
							log.Debug("Assigned name 'username' to first text/email input as fallback")
						}
					case "password":
						if p.CapturePasswords {
							input.SetAttr("name", "password")
							log.Debug("Assigned name 'password' to password field")
						}
					}
				}
			})

			// Add a hidden input to track the original field names
			f.AppendHtml(`<input type="hidden" name="__original_url" value="{{.URL}}"/>`)
			log.Debug("Added hidden URL tracking field")

			// Remove any JavaScript that might prevent form submission
			f.Find("script").Each(func(i int, s *goquery.Selection) {
				scriptText := s.Text()
				if strings.Contains(scriptText, "preventDefault") {
					s.Remove()
					log.Debug("Removed script containing preventDefault")
				}
			})
		} else {
			log.Debug("Credential capture is disabled - removing all input names")
			inputFields := f.Find("input")
			inputFields.Each(func(j int, input *goquery.Selection) {
				input.RemoveAttr("name")
			})
		}

		// Remove any existing submit handlers
		f.RemoveAttr("onsubmit")
		
		// Add our submit handler
		f.SetAttr("onsubmit", "return true;")
		log.Debug("Added form submit handler")
	})
	
	// Also try to remove any external scripts that might interfere
	d.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptText := s.Text()
		if strings.Contains(scriptText, "preventDefault") {
			s.Remove()
			log.Debug("Removed external script containing preventDefault")
		}
	})
	
	p.HTML, err = d.Html()
	if err != nil {
		log.Error("Error getting final HTML:", err)
	} else {
		log.Debug("Successfully processed page HTML")
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
