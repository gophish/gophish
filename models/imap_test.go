package models

import (
	"encoding/json"
	"strings"

	check "gopkg.in/check.v1"
)

// TestIMAPMarshalHidesPassword verifies the CVE-2024-55196 fix: marshaling an
// IMAP object never leaks the stored password (neither the value nor the
// "password" key).
func (s *ModelsSuite) TestIMAPMarshalHidesPassword(ch *check.C) {
	im := IMAP{
		Host:     "imap.example.com",
		Username: "user@example.com",
		Password: "SUPER_SECRET",
	}
	b, err := json.Marshal(im)
	ch.Assert(err, check.Equals, nil)
	out := string(b)
	ch.Assert(strings.Contains(out, "SUPER_SECRET"), check.Equals, false)
	ch.Assert(strings.Contains(out, "\"password\""), check.Equals, false)
	// Non-secret fields must still be present.
	ch.Assert(strings.Contains(out, "\"username\":\"user@example.com\""), check.Equals, true)
}

// TestIMAPUnmarshalAcceptsPassword verifies input is unaffected: a "password"
// in the request body still populates the struct.
func (s *ModelsSuite) TestIMAPUnmarshalAcceptsPassword(ch *check.C) {
	body := `{"host":"imap.example.com","username":"user@example.com","password":"SUPER_SECRET"}`
	im := IMAP{}
	err := json.Unmarshal([]byte(body), &im)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(im.Password, check.Equals, "SUPER_SECRET")
}
