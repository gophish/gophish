package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProviderType represents the cloud provider type
type ProviderType string

const (
	ProviderTypeAzure ProviderType = "azure"
	ProviderTypeAWS   ProviderType = "aws"
)

// IsValid checks if the provider type is valid
func (pt ProviderType) IsValid() bool {
	switch pt {
	case ProviderTypeAzure, ProviderTypeAWS:
		return true
	}
	return false
}

// ProviderTenant represents a cloud provider tenant/account linked to a SaaS tenant
type ProviderTenant struct {
	ID               string       `gorm:"type:text;primary_key"`
	TenantID         string       `gorm:"type:text;index"`
	ProviderType     ProviderType `gorm:"type:text"`
	ProviderTenantID string       `gorm:"type:text"`
	DisplayName      string       `gorm:"type:text"`
	Region           string       `json:"region"`
	CreatedAt        time.Time    `gorm:"type:timestamp"`
	UpdatedAt        time.Time    `gorm:"type:timestamp"`
	Tenant           *Tenant      `json:"tenant" gorm:"foreignkey:TenantID"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (pt *ProviderTenant) BeforeCreate() error {
	if pt.ID == "" {
		pt.ID = uuid.New().String()
	}
	pt.CreatedAt = time.Now().UTC()
	pt.UpdatedAt = time.Now().UTC()
	return nil
}

// BeforeUpdate will update the UpdatedAt timestamp
func (pt *ProviderTenant) BeforeUpdate() error {
	pt.UpdatedAt = time.Now().UTC()
	return nil
}

// Validate checks if the provider tenant has valid data
func (pt *ProviderTenant) Validate() error {
	if pt.TenantID == "" {
		return errors.New("tenant ID cannot be empty")
	}
	if pt.ID == "" {
		return errors.New("provider tenant ID cannot be empty")
	}
	if !pt.ProviderType.IsValid() {
		return errors.New("invalid provider type")
	}
	if pt.ProviderTenantID == "" {
		return errors.New("provider tenant ID cannot be empty")
	}
	return nil
}

// Create inserts a new provider tenant into the database
func (pt *ProviderTenant) Create() error {
	if err := pt.Validate(); err != nil {
		return err
	}
	// Verify parent tenant exists
	if _, err := GetTenant(pt.TenantID); err != nil {
		return errors.New("parent tenant not found")
	}
	return db.Create(pt).Error
}

// Update modifies an existing provider tenant in the database
func (pt *ProviderTenant) Update() error {
	if err := pt.Validate(); err != nil {
		return err
	}
	return db.Save(pt).Error
}

// Delete removes a provider tenant from the database
func (pt *ProviderTenant) Delete() error {
	// Check if the record exists first
	var count int64
	if err := db.Model(&ProviderTenant{}).Where("id = ?", pt.ID).Count(&count).Error; err != nil {
		return fmt.Errorf("provider tenant not found: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("provider tenant not found: record does not exist")
	}

	err := db.Delete(pt).Error
	if err != nil {
		return fmt.Errorf("failed to delete provider tenant: %v", err)
	}
	return nil
}

// GetProviderTenant retrieves a provider tenant by ID
func GetProviderTenant(id string) (*ProviderTenant, error) {
	if id == "" {
		return nil, errors.New("invalid provider tenant ID")
	}
	
	providerTenant := &ProviderTenant{}
	err := db.Where("id = ?", id).First(providerTenant).Error
	if err != nil {
		return nil, fmt.Errorf("provider tenant not found: %v", err)
	}
	return providerTenant, nil
}

// GetProviderTenantsByTenant retrieves all provider tenants for a given tenant
func GetProviderTenantsByTenant(tenantID string) ([]*ProviderTenant, error) {
	if tenantID == "" {
		return nil, errors.New("invalid tenant ID")
	}
	
	var providerTenants []*ProviderTenant
	err := db.Where("tenant_id = ?", tenantID).Find(&providerTenants).Error
	if err != nil {
		return nil, err
	}
	return providerTenants, nil
}

// GetProviderTenants retrieves all provider tenants
func GetProviderTenants() ([]*ProviderTenant, error) {
	var providerTenants []*ProviderTenant
	err := db.Find(&providerTenants).Error
	return providerTenants, err
}

// GetProviderTenantsByType retrieves all provider tenants of a specific type
func GetProviderTenantsByType(providerType ProviderType) ([]*ProviderTenant, error) {
	if !providerType.IsValid() {
		return nil, errors.New("invalid provider type")
	}
	
	var providerTenants []*ProviderTenant
	err := db.Where("provider_type = ?", providerType).Find(&providerTenants).Error
	return providerTenants, err
}

// GetProviderTenantByProviderTenantID returns a provider tenant by its provider_tenant_id
func GetProviderTenantByProviderTenantID(providerTenantID string) (ProviderTenant, error) {
	pt := ProviderTenant{}
	err := db.Where("provider_tenant_id = ?", providerTenantID).First(&pt).Error
	return pt, err
}

const DisplayNameMicrosoft = "microsoft" 