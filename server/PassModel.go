package main

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Pass struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CompanyID   string
	CompanyName string
	IBAN        string
	BIC         string
	Address     string
	Cashback    string
	CreatedAt   time.Time // Automatically managed by GORM for creation time
	UpdatedAt   time.Time
}

func (pass *Pass) BeforeCreate(tx *gorm.DB) (err error) {
	pass.ID = uuid.New()
	return
}

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

func UpdatePassByCompanyID(db *gorm.DB, companyID, cashback string) error {
	// Update cashback
	var pass Pass
	if err := db.Where("company_id = ?", companyID).First(&pass).Error; err != nil {
		return err
	}

	if err := db.Model(&pass).Update("cashback", cashback).Error; err != nil {
		return err
	}

	log.Printf("Cashback updated: %v\n", pass)

	return nil
}
