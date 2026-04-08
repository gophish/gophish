package models

import (
	"encoding/base64"
	"fmt"
	"strings"

	check "gopkg.in/check.v1"
)

type mockTemplateContext struct {
	URL         string
	FromAddress string
}

func (m mockTemplateContext) getFromAddress() string {
	return m.FromAddress
}

func (m mockTemplateContext) getBaseURL() string {
	return m.URL
}

func decodeQRPayload(c *check.C, qr string) []byte {
	const prefix = "<img src=\"data:image/png;base64,"
	const suffix = "\" alt=\"QR Code\">"

	c.Assert(strings.HasPrefix(qr, prefix), check.Equals, true)
	c.Assert(strings.HasSuffix(qr, suffix), check.Equals, true)

	encoded := strings.TrimSuffix(strings.TrimPrefix(qr, prefix), suffix)
	png, err := base64.StdEncoding.DecodeString(encoded)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(png) >= 8, check.Equals, true)
	c.Assert(string(png[:8]), check.Equals, "\x89PNG\r\n\x1a\n")
	return png
}

func (s *ModelsSuite) TestNewTemplateContext(c *check.C) {
	r := Result{
		BaseRecipient: BaseRecipient{
			FirstName: "Foo",
			LastName:  "Bar",
			Email:     "foo@bar.com",
		},
		RId: "1234567",
	}
	ctx := mockTemplateContext{
		URL:         "http://example.com",
		FromAddress: "From Address <from@example.com>",
	}
	expected := PhishingTemplateContext{
		URL:           fmt.Sprintf("%s?rid=%s", ctx.URL, r.RId),
		BaseURL:       ctx.URL,
		BaseRecipient: r.BaseRecipient,
		TrackingURL:   fmt.Sprintf("%s/track?rid=%s", ctx.URL, r.RId),
		From:          "From Address",
		RId:           r.RId,
	}
	expected.Tracker = "<img alt='' style='display: none' src='" + expected.TrackingURL + "'/>"
	got, err := NewPhishingTemplateContext(ctx, r.BaseRecipient, r.RId)
	c.Assert(err, check.Equals, nil)
	c.Assert(got.QR, check.Not(check.Equals), "")
	decodeQRPayload(c, got.QR)
	expected.QR = got.QR
	c.Assert(got, check.DeepEquals, expected)
}

func (s *ModelsSuite) TestValidateTemplateWithQR(c *check.C) {
	err := ValidateTemplate("<div>{{.QR}}</div>")
	c.Assert(err, check.Equals, nil)
}
