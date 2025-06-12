package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultSystemTenantName = "0sysuser"
)

// Tenant represents a top-level organization in the multi-tenant model
type Tenant struct {
	ID        string    `gorm:"type:text;primary_key"`
	Name      string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"type:timestamp"`
	UpdatedAt time.Time `gorm:"type:timestamp"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (t *Tenant) BeforeCreate() error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// BeforeUpdate will update the UpdatedAt timestamp
func (t *Tenant) BeforeUpdate() error {
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Validate checks if the tenant has valid data
func (t *Tenant) Validate() error {
    if t.Name == "" {
        return errors.New("tenant name cannot be empty")
    }
	if t.ID == "" {
        return errors.New("tenant ID cannot be empty")
    }
    return nil
}

// Create inserts a new tenant into the database
func (t *Tenant) Create() error {
    if err := t.Validate(); err != nil {
        return err
    }
	if t.ID == "" {
		t.ID = uuid.New().String()
    }
    t.CreatedAt = time.Now().UTC()
    t.UpdatedAt = time.Now().UTC()
    return db.Create(t).Error
}

// Update modifies an existing tenant in the database
func (t *Tenant) Update() error {
    if err := t.Validate(); err != nil {
        return err
    }
    return db.Save(t).Error
}

// Delete removes a tenant from the database
func (t *Tenant) Delete() error {
    // Check if the record exists first
    var count int64
    if err := db.Model(&Tenant{}).Where("id = ?", t.ID).Count(&count).Error; err != nil {
        return fmt.Errorf("tenant not found: %v", err)
    }
    if count == 0 {
        return fmt.Errorf("tenant not found: record does not exist")
    }

    err := db.Delete(t).Error
    if err != nil {
        return fmt.Errorf("tenant not found: %v", err)
    }
    return nil
}

// GetTenant retrieves a tenant by ID
func GetTenant(id string) (*Tenant, error) {
	if id == "" {
        return nil, errors.New("invalid tenant ID")
    }
    
    tenant := &Tenant{}
    err := db.Where("id = ?", id).First(tenant).Error
    if err != nil {
        return nil, fmt.Errorf("tenant not found: %v", err)
    }
    return tenant, nil
}

// GetTenants retrieves all tenants
func GetTenants() ([]*Tenant, error) {
    var tenants []*Tenant
    err := db.Find(&tenants).Error
    return tenants, err
}

// GetTenantByName returns a tenant by its name
func GetTenantByName(name string) (Tenant, error) {
	t := Tenant{}
	err := db.Where("name = ?", name).First(&t).Error
	return t, err
} 