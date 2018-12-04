package models

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	log "github.com/gophish/gophish/logger"
)

// If the user did not enter a friendly name
var ErrFriendlyNameNotSpecified = errors.New("No friendly name specified")

//The user did not enter a public key
var ErrPublicKeyNotSpecified = errors.New("No public key specified")

// Public key contains the fields used for a Public key model
type PublicKey struct {
	Id           int64  `json:"id"`
	FriendlyName string `json:"name"`
	UserId       int64  `json:"-"`
	PubKey       string `json:"pub_key"`
}

func (p *PublicKey) Validate() error {
	switch {
	case p.FriendlyName == "":
		return ErrFriendlyNameNotSpecified
	case p.PubKey == "":
		return ErrPublicKeyNotSpecified
	}

	_, err := DecodePEMBlock(p.PubKey)

	return err
}

// PutPublicKey adds public key
func PutPubKey(p *PublicKey) error {
	err := db.Save(p).Error
	return err
}

// GetPublicKeys returns all public keys for given user (id)
func GetPublicKeys(uid int64) ([]PublicKey, error) {
	ps := []PublicKey{}
	err := db.Where("user_id=?", uid).Find(&ps).Error
	if err != nil {
		log.Error(err)
		return ps, err
	}
	return ps, err
}

// GetPublicKey returns the public key, if it exists, specified by the given id and user_id.
func GetPublicKey(id int64, uid int64) (PublicKey, error) {
	p := PublicKey{}
	err := db.Where("id = ?", id).Where("user_id = ?", uid).Find(&p).Error
	if err != nil {
		log.Errorf("%s: public key not found (id: %d,uid: %d)", err, id, uid)
		return p, err
	}
	return p, err
}

// GetPublicKeyByName returns the public key, if it exists, specified by the given name and user_id.
func GetPublicKeyByName(n string, uid int64) (PublicKey, error) {
	p := PublicKey{}
	err := db.Where("user_id=? and friendly_name=?", uid, n).Find(&p).Error
	if err != nil {
		log.Error(err)
		return p, err
	}

	return p, err
}

// PutPublicKey edits an existing public key in the database.
// Per the PUT Method RFC, it presumes all data for a public key is provided.
func PutPublicKey(p *PublicKey) error {
	err := p.Validate()
	if err != nil {
		log.Error(err)
		return err
	}

	record, err := GetPublicKeyByName(p.FriendlyName, p.UserId)
	if err == nil && record.Id != p.Id {
		return ErrRecordAlreadyExists
	}

	err = db.Where("id=?", p.Id).Save(p).Error
	if err != nil {
		log.Error(err)
	}

	return err
}

// PostPublicKey creates a new publc key in the database.
func PostPublicKey(p *PublicKey) error {
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

// DeletePublicKey deletes an existing Public key in the database.
// An error is returned if a Public key with the given user id and public key id is not found.
func DeletePublicKey(id int64, uid int64) error {
	err = db.Where("user_id=?", uid).Delete(PublicKey{Id: id}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

func DecodePEMBlock(pemBlock string) (pubkey *rsa.PublicKey, err error) {
	block, _ := pem.Decode([]byte(pemBlock))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("Block was not public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("Not RSA public key")
	}

}
