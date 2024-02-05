package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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

type DeviceRegistration struct {
	DeviceLibraryIdentifier string    `json:"deviceLibraryIdentifier"`
	PassTypeIdentifier      string    `json:"passTypeIdentifier"`
	SerialNumber            string    `json:"serialNumber"`
	PushToken               string    `json:"pushToken"`
	CreatedAt               time.Time // Automatically managed by GORM for creation time
	UpdatedAt               time.Time // Automatically managed by GORM for update time
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
		" dbname=postgres" +
		" port=" + os.Getenv("POSTGRES_PORT") +
		" sslmode=disable TimeZone=Etc/UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	databaseName := os.Getenv("POSTGRES_DB")

	// Check if the database exists
	var dbName string
	err = db.Raw("SELECT datname FROM pg_database WHERE datname = ?", databaseName).Scan(&dbName).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to query database existence")
	}

	if dbName == "" {
		// Database does not exist, attempt to create it
		err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", databaseName)).Error
		if err != nil {
			return nil, err
		} else {
			log.Info().Msgf("Database %s created successfully", databaseName)
		}
	} else {
		log.Info().Msgf("Database %s already exists, skipping creation", databaseName)
	}

	// Close the initial connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.Close()

	dsn = fmt.Sprintf("%s dbname=%s", dsn, os.Getenv("POSTGRES_DB"))
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Create the extension within the new database
	err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		return nil, err
	} else {
		log.Info().Msg("uuid-ossp extension created successfully")
	}

	// Migrate the schema
	db.AutoMigrate(&Pass{}, &DeviceRegistration{})

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

	log.Debug().
		Interface("Pass", pass).
		Msg("New pass created/updated")

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

	log.Debug().
		Interface("Pass", pass).
		Msg("Cashback updated")

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

func RegisterDevice(db *gorm.DB, deviceLibraryIdentifier, serialNumber, pushToken string) (DeviceRegistration, error, bool) {
	deviceReg := DeviceRegistration{
		DeviceLibraryIdentifier: deviceLibraryIdentifier,
		PassTypeIdentifier:      "pass.com.finom.bank2wallet",
		SerialNumber:            serialNumber,
		PushToken:               pushToken,
	}

	rec := db.Where(DeviceRegistration{SerialNumber: serialNumber}).Find(&deviceReg)
	exists := rec.RowsAffected > 0
	if exists {
		if deviceReg.PushToken != pushToken {
			if err := db.Model(&deviceReg).Update("push_token", pushToken).Error; err != nil {
				return DeviceRegistration{}, err, false
			}
		}
	} else {
		if err := db.Create(&deviceReg).Error; err != nil {
			return DeviceRegistration{}, err, false
		}
	}

	return deviceReg, nil, exists
}

func GetPassesByDeviceID(db *gorm.DB, deviceLibraryIdentifier string) ([]string, error) {
	var deviceRegs []DeviceRegistration
	if err := db.Where("device_library_identifier = ?", deviceLibraryIdentifier).Find(&deviceRegs).Error; err != nil {
		return nil, err
	}

	serialNumbers := make([]string, 0, len(deviceRegs))
	for _, deviceReg := range deviceRegs {
		serialNumbers = append(serialNumbers, deviceReg.SerialNumber)
	}

	return serialNumbers, nil
}

func GetUpdatedPasses(db *gorm.DB, deviceLibraryIdentifier, passesUpdatedSince string) ([]string, error) {
	var passes []Pass
	if err := db.Where("updated_at > ?", passesUpdatedSince).Find(&passes).Error; err != nil {
		return nil, err
	}

	serialNumbers := make([]string, 0, len(passes))
	for _, pass := range passes {
		serialNumbers = append(serialNumbers, pass.ID.String())
	}

	return serialNumbers, nil
}

func DeletePassOnDevice(db *gorm.DB, serialNumber string) error {
	if err := db.Where("serial_number = ?", serialNumber).Delete(&DeviceRegistration{}).Error; err != nil {
		return err
	}

	return nil
}
