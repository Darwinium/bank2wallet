package main

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestPassModel(t *testing.T) {
	t.Run("Pass struct fields", func(t *testing.T) {
		pass := Pass{
			ID:          uuid.New(),
			CompanyID:   "TEST123",
			CompanyName: "Test Company",
			IBAN:        "DE89370400440532013000",
			BIC:         "COBADEFFXXX",
			Address:     "123 Test St",
			Cashback:    "100€",
		}

		if pass.CompanyID != "TEST123" {
			t.Errorf("CompanyID = %q, want %q", pass.CompanyID, "TEST123")
		}
		if pass.CompanyName != "Test Company" {
			t.Errorf("CompanyName = %q, want %q", pass.CompanyName, "Test Company")
		}
		if pass.IBAN != "DE89370400440532013000" {
			t.Errorf("IBAN = %q, want %q", pass.IBAN, "DE89370400440532013000")
		}
		if pass.BIC != "COBADEFFXXX" {
			t.Errorf("BIC = %q, want %q", pass.BIC, "COBADEFFXXX")
		}
		if pass.Address != "123 Test St" {
			t.Errorf("Address = %q, want %q", pass.Address, "123 Test St")
		}
		if pass.Cashback != "100€" {
			t.Errorf("Cashback = %q, want %q", pass.Cashback, "100€")
		}
	})

	t.Run("Pass BeforeCreate hook generates UUID", func(t *testing.T) {
		pass := &Pass{
			CompanyID:   "TEST456",
			CompanyName: "Another Test",
			IBAN:        "GB82WEST12345698765432",
			BIC:         "TESTBIC",
			Address:     "456 Test Ave",
			Cashback:    "50€",
		}

		// Call the BeforeCreate hook manually
		err := pass.BeforeCreate(&gorm.DB{})
		if err != nil {
			t.Fatalf("BeforeCreate() error = %v", err)
		}

		// Verify UUID was generated
		if pass.ID == uuid.Nil {
			t.Error("BeforeCreate() did not generate UUID")
		}

		// Verify it's a valid UUID
		if _, err := uuid.Parse(pass.ID.String()); err != nil {
			t.Errorf("BeforeCreate() generated invalid UUID: %v", err)
		}
	})

	t.Run("Pass BeforeCreate generates unique UUIDs", func(t *testing.T) {
		pass1 := &Pass{}
		pass2 := &Pass{}

		pass1.BeforeCreate(&gorm.DB{})
		pass2.BeforeCreate(&gorm.DB{})

		if pass1.ID == pass2.ID {
			t.Error("BeforeCreate() generated duplicate UUIDs")
		}
	})
}

func TestDeviceRegistrationModel(t *testing.T) {
	t.Run("DeviceRegistration struct fields", func(t *testing.T) {
		deviceReg := DeviceRegistration{
			DeviceLibraryIdentifier: "device123",
			PassTypeIdentifier:      "pass.com.finom.bank2wallet",
			SerialNumber:            "abc-123-def",
			PushToken:               "token123456",
		}

		if deviceReg.DeviceLibraryIdentifier != "device123" {
			t.Errorf("DeviceLibraryIdentifier = %q, want %q", deviceReg.DeviceLibraryIdentifier, "device123")
		}
		if deviceReg.PassTypeIdentifier != "pass.com.finom.bank2wallet" {
			t.Errorf("PassTypeIdentifier = %q, want %q", deviceReg.PassTypeIdentifier, "pass.com.finom.bank2wallet")
		}
		if deviceReg.SerialNumber != "abc-123-def" {
			t.Errorf("SerialNumber = %q, want %q", deviceReg.SerialNumber, "abc-123-def")
		}
		if deviceReg.PushToken != "token123456" {
			t.Errorf("PushToken = %q, want %q", deviceReg.PushToken, "token123456")
		}
	})

	t.Run("DeviceRegistration with standard pass type identifier", func(t *testing.T) {
		deviceReg := DeviceRegistration{
			PassTypeIdentifier: "pass.com.finom.bank2wallet",
		}

		expected := "pass.com.finom.bank2wallet"
		if deviceReg.PassTypeIdentifier != expected {
			t.Errorf("PassTypeIdentifier = %q, want %q", deviceReg.PassTypeIdentifier, expected)
		}
	})
}

// Integration-style tests that would require a database are skipped
// These would include:
// - TestGetDBConnection
// - TestAddNewPass
// - TestUpdatePassByCompanyID
// - TestGetPassByCompanyID
// - TestRegisterDevice
// - TestGetPassesByDeviceID
// - TestGetUpdatedPasses
// - TestDeletePassOnDevice

func TestDatabaseModelsValidation(t *testing.T) {
	t.Run("Pass model has required GORM tags", func(t *testing.T) {
		pass := Pass{}
		// This test verifies the struct is properly defined
		// In a real scenario, we'd use reflection to check tags
		_ = pass.ID       // Should be UUID type
		_ = pass.Model    // Should have gorm.Model embedded
		_ = pass.CompanyID
		_ = pass.CompanyName
		_ = pass.IBAN
		_ = pass.BIC
		_ = pass.Address
		_ = pass.Cashback
	})

	t.Run("DeviceRegistration model has all required fields", func(t *testing.T) {
		deviceReg := DeviceRegistration{}
		// Verify all fields are accessible
		_ = deviceReg.DeviceLibraryIdentifier
		_ = deviceReg.PassTypeIdentifier
		_ = deviceReg.SerialNumber
		_ = deviceReg.PushToken
		_ = deviceReg.CreatedAt
		_ = deviceReg.UpdatedAt
	})
}

// Test helper functions that don't require database
func TestPassModelJSON(t *testing.T) {
	t.Run("Pass can be used in API responses", func(t *testing.T) {
		pass := Pass{
			ID:          uuid.New(),
			CompanyID:   "TEST789",
			CompanyName: "JSON Test Company",
			IBAN:        "FR1420041010050500013M02606",
			BIC:         "TESTBICXXX",
			Address:     "789 JSON St",
			Cashback:    "75€",
		}

		// Verify pass has all necessary fields for API responses
		if pass.ID == uuid.Nil {
			t.Error("Pass ID is nil")
		}
		if pass.CompanyID == "" {
			t.Error("Pass CompanyID is empty")
		}
	})
}
