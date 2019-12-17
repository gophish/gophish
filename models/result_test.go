package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/mail"
	"net/url"
	"regexp"
	"time"

	"crypto/sha256"

	"gopkg.in/check.v1"
)

func (s *ModelsSuite) TestGenerateResultId(c *check.C) {
	r := Result{}
	r.GenerateId(db)
	match, err := regexp.Match("[a-zA-Z0-9]{7}", []byte(r.RId))
	c.Assert(err, check.Equals, nil)
	c.Assert(match, check.Equals, true)
}

func (s *ModelsSuite) TestFormatAddress(c *check.C) {
	r := Result{
		BaseRecipient: BaseRecipient{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "johndoe@example.com",
		},
	}
	expected := &mail.Address{
		Name:    "John Doe",
		Address: "johndoe@example.com",
	}
	c.Assert(r.FormatAddress(), check.Equals, expected.String())

	r = Result{
		BaseRecipient: BaseRecipient{Email: "johndoe@example.com"},
	}
	c.Assert(r.FormatAddress(), check.Equals, r.Email)
}

func (s *ModelsSuite) TestResultSendingStatus(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	// This campaign wasn't scheduled, so we expect the status to
	// be sending
	for _, r := range c.Results {
		ch.Assert(r.Status, check.Equals, StatusSending)
		ch.Assert(r.ModifiedDate, check.Equals, c.CreatedDate)
	}
}
func (s *ModelsSuite) TestResultScheduledStatus(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	c.LaunchDate = time.Now().UTC().Add(time.Hour * time.Duration(1))
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	// This campaign wasn't scheduled, so we expect the status to
	// be sending
	for _, r := range c.Results {
		ch.Assert(r.Status, check.Equals, StatusScheduled)
		ch.Assert(r.ModifiedDate, check.Equals, c.CreatedDate)
	}
}

func (s *ModelsSuite) TestResultVariableStatus(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	c.LaunchDate = time.Now().UTC()
	c.SendByDate = c.LaunchDate.Add(2 * time.Minute)
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)

	// The campaign has a window smaller than our group size, so we expect some
	// emails to be sent immediately, while others will be scheduled
	for _, r := range c.Results {
		if r.SendDate.Before(c.CreatedDate) || r.SendDate.Equal(c.CreatedDate) {
			ch.Assert(r.Status, check.Equals, StatusSending)
		} else {
			ch.Assert(r.Status, check.Equals, StatusScheduled)
		}
	}
}

func (s *ModelsSuite) TestDuplicateResults(ch *check.C) {
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{BaseRecipient: BaseRecipient{Email: "test1@example.com", FirstName: "First", LastName: "Example"}},
		Target{BaseRecipient: BaseRecipient{Email: "test1@example.com", FirstName: "Duplicate", LastName: "Duplicate"}},
		Target{BaseRecipient: BaseRecipient{Email: "test2@example.com", FirstName: "Second", LastName: "Example"}},
	}
	group.UserId = 1
	ch.Assert(PostGroup(&group), check.Equals, nil)

	// Add a template
	t := Template{Name: "Test Template"}
	t.Subject = "{{.RId}} - Subject"
	t.Text = "{{.RId}} - Text"
	t.HTML = "{{.RId}} - HTML"
	t.UserId = 1
	ch.Assert(PostTemplate(&t), check.Equals, nil)

	// Add a landing page
	p := Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	ch.Assert(PostPage(&p), check.Equals, nil)

	// Add a sending profile
	smtp := SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	ch.Assert(PostSMTP(&smtp), check.Equals, nil)

	c := Campaign{Name: "Test campaign"}
	c.UserId = 1
	c.Template = t
	c.Page = p
	c.SMTP = smtp
	c.Groups = []Group{group}

	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)
	ch.Assert(len(c.Results), check.Equals, 2)
	ch.Assert(c.Results[0].Email, check.Equals, group.Targets[0].Email)
	ch.Assert(c.Results[1].Email, check.Equals, group.Targets[2].Email)
}

func (s *ModelsSuite) TestCreateEvent(ch *check.C) {
	c := s.createCampaignDependencies(ch)
	c.LaunchDate = time.Now().UTC().Add(time.Hour * time.Duration(1))
	ch.Assert(PostCampaign(&c, c.UserId), check.Equals, nil)

	mockResult := Result{
		CampaignId: c.Id,
		UserId:     c.UserId,
	}
	mockResult.Email = "dummy@dummy.com"

	payload, err := url.ParseQuery("hackerman=password&password=pass")
	ch.Assert(err, check.Equals, nil)

	mockDetails := EventDetails{
		Payload: payload,
	}

	returnEvent, err := mockResult.createEvent(EventDataSubmit, mockDetails)
	ch.Assert(err, check.Equals, nil)

	privKey := BytesToPrivateKey([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAz8qUODbqjWxcL8eNngjCfwO6bstHOt6p8EvHahem6JQ/VeId
J4h7Hy0eTxm68sXKvliWrs6J3uJUAAZlZqX5E9uMaSjiaF+aLVjQOj+fqmh/+Unp
ZUa/p2WtWy1YyZuAZ9o4EcbeUFokkS8oIXfWvyPrE5ggAUNiW4p/4iHltjqKyt8z
2cids36j09OLz0hnGxzAq4PQvdYnW0OyLkbqFwvpiR8/9JY4O7pM4dUaQBQhvj+a
hbuYhdO+tsnE7cRMOLNXfc8vDtdTY08BfL5ZvFsuNexQlF1DnL5VIETx9WHmbT77
A00VJp3VeUxADpYoSyrKQ5settc+dSFvp7kPqwIDAQABAoIBAQChT3XjchaCdbXX
TcaOAeTj78QCkQKaHBO3PUzw+q2TbssAZEoXz6ctW7xk0efE4nHzdRh9Wk/D3NYz
MsPzfTOxC8akOJ4TQuyQ+ZqlLZFJHFkE8TEkc1kDnjaHStvbn0F+3fSbMFST8nbi
0sYHvV3UAxFSo81szaAEyq9eoMwQu1m8Oi1CA+A8Sc0fuCkRTpIplWiaz2gzh+QO
kBFtcdcNVSR4qGLFDMMsdgZL/Dg9GE9TpTbskeOApZjYkp0+Yh+bl838gXDdiyJw
vELdbTFouex5PSkh136xNzVi0YRlCqorPy/TB2H2oXagwRop9IyZZAzZdel7dm7C
TSJSNeqxAoGBAOazOVcCGvCjxUOHzFnwxmgrsNV3uJSIoOBdYd9KLvI2miMM+gxd
IJdsdUIHMwUcYmUhqZfO0FkLGhtGSPkXQikQdZqOXXA6zQd1QasapmNQE8Wfj8CZ
05WmjuzVDcsped7+/zdo5FDsch6zdyRxiRFWUfVJFpsqy5wNMVRKSIg5AoGBAOaU
NEDnHC/bERZiEhwW5TRQ9DkDGLvp7sopdh/ZUwiNEo+GMw/UL1avmAg7gbZVs7eX
F/e0xr9RdZYvXB7sDWvKtYQI6bC0SjhOhvd8bzqbyC4OqtgrSD0xEnoUtQEi2Xx5
d9PQeWqlpEDK2wFnIJwsES5lJsr7F2cMQUvExS8DAoGADZks3RMTsXGF1CgyBG8r
0sIYh0yqRZ8UFIWMmlPOFprfVQeTyZzHqgVLmBvChx+YMSvdykP3hfggjtECxiP3
02HT/Ms9eLsOkMz5lPNaMWpr7+8q0wh+L0kFDbK1QG9ubpWLR6HYK2j0hRjBAhXr
JWl4JUQsn/LS05z3dmd2hQkCgYAeokQK92l8RiuQALmNN9F90N+Rj4LCvIK4IygJ
dTMd6Lg1j0vLZ5JefvfA6D8EfYBh/NX3V/IryuPHb0Va6luiHY1eHF0H1/wgXPZ7
fPG+JKJE1DgIfj+buaBNzeB6ZSnl6rFr17+51oXrAch0+EGR3hzuQAwWXaOvUiZ+
rYbRBwKBgQDfZ4FDt/8f7sSZQffiZp09D5+OTqp7lzfjgTxG86ulBsZRKQRvfvS2
r8KnQTykvDjHQG4VSjENjuTcQaqYE19FlpPkrOZx30/16SGUrbxH6hMgKQVL/ATb
RHmhRpPQiO9WJPNaZBtJvgRh886u0Vss6nNCIRb7L2mCCGZwjeFmdw==
-----END RSA PRIVATE KEY-----`), ch)

	key, err := base64.StdEncoding.DecodeString(returnEvent.Key)
	ch.Assert(err, check.Equals, nil)

	plainTextKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, key, []byte("key"))
	ch.Assert(err, check.Equals, nil)

	IVAndcipherText, err := base64.StdEncoding.DecodeString(returnEvent.Details)
	ch.Assert(err, check.Equals, nil)

	iv := IVAndcipherText[:aes.BlockSize]
	cipherText := IVAndcipherText[aes.BlockSize:]

	blockCipher, err := aes.NewCipher(plainTextKey)
	ch.Assert(err, check.Equals, nil)

	streamCipher := cipher.NewCFBDecrypter(blockCipher, iv)
	streamCipher.XORKeyStream(cipherText, cipherText)

	ch.Assert(string(cipherText), check.Equals, `{"payload":{"hackerman":["password"],"password":["pass"]},"browser":null}`)

}

func BytesToPrivateKey(priv []byte, ch *check.C) *rsa.PrivateKey {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {

		b, err = x509.DecryptPEMBlock(block, nil)
		ch.Assert(err, check.Equals, nil)
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	ch.Assert(err, check.Equals, nil)
	return key
}
