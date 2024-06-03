package models

import (
	"fmt"

	"github.com/jinzhu/gorm"

	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) TestPostSMTP(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "foo@example.com",
		UserId:      1,
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, nil)
	ss, err := GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 1)
}

func (s *ModelsSuite) TestPostSMTPNoHost(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		FromAddress: "foo@example.com",
		UserId:      1,
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrHostNotSpecified)
}

func (s *ModelsSuite) TestPostSMTPNoFrom(c *check.C) {
	smtp := SMTP{
		Name:   "Test SMTP",
		UserId: 1,
		Host:   "1.1.1.1:25",
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrFromAddressNotSpecified)
}

func (s *ModelsSuite) TestPostInvalidFrom(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "Foo Bar <foo@example.com>",
		UserId:      1,
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrInvalidFromAddress)
}

func (s *ModelsSuite) TestPostInvalidFromEmail(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "example.com",
		UserId:      1,
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, ErrInvalidFromAddress)
}

func (s *ModelsSuite) TestPostSMTPValidHeader(c *check.C) {
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "foo@example.com",
		UserId:      1,
		Headers: []Header{
			Header{Key: "Reply-To", Value: "test@example.com"},
			Header{Key: "X-Mailer", Value: "gophish"},
		},
	}
	err := PostSMTP(&smtp)
	c.Assert(err, check.Equals, nil)
	ss, err := GetSMTPs(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(ss), check.Equals, 1)
}

func (s *ModelsSuite) TestSMTPGetDialer(ch *check.C) {
	host := "localhost"
	port := 25
	smtp := SMTP{
		Host:             fmt.Sprintf("%s:%d", host, port),
		IgnoreCertErrors: false,
	}
	d, err := smtp.GetDialer()
	ch.Assert(err, check.Equals, nil)

	dialer := d.(*Dialer).Dialer
	ch.Assert(dialer.Host, check.Equals, host)
	ch.Assert(dialer.Port, check.Equals, port)
	ch.Assert(dialer.TLSConfig.ServerName, check.Equals, host)
	ch.Assert(dialer.TLSConfig.InsecureSkipVerify, check.Equals, smtp.IgnoreCertErrors)
}

func (s *ModelsSuite) TestGetInvalidSMTP(ch *check.C) {
	_, err := GetSMTP(-1, 1)
	ch.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestDefaultDeniedDial(ch *check.C) {
	host := "169.254.169.254"
	port := 25
	smtp := SMTP{
		Host: fmt.Sprintf("%s:%d", host, port),
	}
	d, err := smtp.GetDialer()
	ch.Assert(err, check.Equals, nil)
	_, err = d.Dial()
	ch.Assert(err, check.ErrorMatches, ".*upstream connection denied.*")
}
