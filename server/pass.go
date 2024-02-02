package main

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Pass represents the pass model
type Pass struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"` // ID is the UUID of the pass
	CompanyID   string    // CompanyID is the ID of the company
	CompanyName string    // CompanyName is the name of the company
	IBAN        string    // IBAN is the International Bank Account Number
	BIC         string    // BIC is the Bank Identifier Code
	Address     string    // Address is the address of the company
	Cashback    string    // Cashback is the cashback balance in euros
	CreatedAt   time.Time // Automatically managed by GORM for creation time
	UpdatedAt   time.Time // Automatically managed by GORM for update time
}

// BeforeCreate is a GORM hook that is called before creating a new pass. It sets the ID of the pass to a new UUID.
func (pass *Pass) BeforeCreate(tx *gorm.DB) (err error) {
	pass.ID = uuid.New()
	return
}

// getDBConnection returns a new database connection
func getDBConnection() (*gorm.DB, error) {
	dsn := "host=" + os.Getenv("POSTGRES_HOST") +
		" user=" + os.Getenv("POSTGRES_USER") +
		" password=" + os.Getenv("POSTGRES_PASSWORD") +
		" dbname=" + os.Getenv("POSTGRES_DB") +
		" port=" + os.Getenv("POSTGRES_PORT") +
		" sslmode=disable TimeZone=Etc/UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	db.AutoMigrate(&Pass{})

	return db, nil
}

// AddNewPass creates a new pass with the given data and saves it in the database. It returns the pass data
func AddNewPass(db *gorm.DB, companyID, cashback, companyName, iban, bic, address string) (Pass, error) {
	// Create a new pass
	pass := Pass{
		CompanyID:   companyID,
		CompanyName: companyName,
		IBAN:        iban,
		BIC:         bic,
		Address:     address,
		Cashback:    cashback,
	}

	// Check if a pass with the given companyID already exists, if not create a new one
	if err := db.Where(Pass{CompanyID: companyID}).Assign(pass).FirstOrCreate(&pass).Error; err != nil {
		return Pass{}, err
	}

	log.Printf("Pass updated or created: %v\n", pass)

	return pass, nil
}

// UpdatePassByCompanyID updates the cashback of the pass with the given companyID
func UpdatePassByCompanyID(db *gorm.DB, companyID, cashback string) (Pass, error) {
	// Update cashback
	var pass Pass
	if err := db.Where("company_id = ?", companyID).First(&pass).Error; err != nil {
		return Pass{}, err
	}

	if err := db.Model(&pass).Update("cashback", cashback).Error; err != nil {
		return Pass{}, err
	}

	log.Printf("Cashback updated: %v\n", pass)

	return pass, nil
}

// GetPassByCompanyID returns the pass with the given companyID
func GetPassByCompanyID(db *gorm.DB, companyID string) (Pass, error) {
	var pass Pass
	if err := db.Where("company_id = ?", companyID).First(&pass).Error; err != nil {
		return Pass{}, err
	}

	return pass, nil
}
