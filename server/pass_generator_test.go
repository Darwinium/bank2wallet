package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCreatePassStructure(t *testing.T) {
	// Set required environment variables
	os.Setenv("WEB_SERVICE_URL", "https://test.example.com")
	os.Setenv("AUTH_TOKEN", "test-auth-token")
	defer func() {
		os.Unsetenv("WEB_SERVICE_URL")
		os.Unsetenv("AUTH_TOKEN")
	}()

	// Create a test pass
	testPass := Pass{
		Model:       gorm.Model{},
		ID:          uuid.New(),
		CompanyID:   "TEST123",
		CompanyName: "Test Company Ltd.",
		IBAN:        "DE89370400440532013000",
		BIC:         "COBADEFFXXX",
		Address:     "123 Test Street, Berlin, Germany",
		Cashback:    "50€",
	}

	passData := CreatePassStructure(testPass)

	// Test basic structure
	t.Run("Format version is correct", func(t *testing.T) {
		if passData.FormatVersion != 1 {
			t.Errorf("FormatVersion = %d, want 1", passData.FormatVersion)
		}
	})

	t.Run("Pass type identifier is set", func(t *testing.T) {
		expected := "pass.com.finom.bank2wallet"
		if passData.PassTypeIdentifier != expected {
			t.Errorf("PassTypeIdentifier = %q, want %q", passData.PassTypeIdentifier, expected)
		}
	})

	t.Run("Serial number matches pass ID", func(t *testing.T) {
		expected := testPass.ID.String()
		if passData.SerialNumber != expected {
			t.Errorf("SerialNumber = %q, want %q", passData.SerialNumber, expected)
		}
	})

	t.Run("Web service URL is correct", func(t *testing.T) {
		expected := "https://test.example.com/pass/v1/registerDevice"
		if passData.WebServiceURL != expected {
			t.Errorf("WebServiceURL = %q, want %q", passData.WebServiceURL, expected)
		}
	})

	t.Run("Authentication token is set", func(t *testing.T) {
		expected := "test-auth-token"
		if passData.AuthenticationToken != expected {
			t.Errorf("AuthenticationToken = %q, want %q", passData.AuthenticationToken, expected)
		}
	})

	t.Run("Team identifier is correct", func(t *testing.T) {
		expected := "35XPTK6L36"
		if passData.TeamIdentifier != expected {
			t.Errorf("TeamIdentifier = %q, want %q", passData.TeamIdentifier, expected)
		}
	})

	t.Run("Organization name is set", func(t *testing.T) {
		expected := "Finom"
		if passData.OrganizationName != expected {
			t.Errorf("OrganizationName = %q, want %q", passData.OrganizationName, expected)
		}
	})

	t.Run("Colors are set correctly", func(t *testing.T) {
		if passData.BackgroundColor != "rgb(255, 76, 92)" {
			t.Errorf("BackgroundColor = %q, want %q", passData.BackgroundColor, "rgb(255, 76, 92)")
		}
		if passData.ForegroundColor != "rgb(255, 255, 255)" {
			t.Errorf("ForegroundColor = %q, want %q", passData.ForegroundColor, "rgb(255, 255, 255)")
		}
		if passData.LabelColor != "rgb(11, 0, 46)" {
			t.Errorf("LabelColor = %q, want %q", passData.LabelColor, "rgb(11, 0, 46)")
		}
	})

	t.Run("Header fields contain cashback", func(t *testing.T) {
		if len(passData.Generic.HeaderFields) != 1 {
			t.Fatalf("HeaderFields length = %d, want 1", len(passData.Generic.HeaderFields))
		}
		field := passData.Generic.HeaderFields[0]
		if field.Key != "cashback" {
			t.Errorf("HeaderField key = %q, want %q", field.Key, "cashback")
		}
		if field.Label != "CASHBACK" {
			t.Errorf("HeaderField label = %q, want %q", field.Label, "CASHBACK")
		}
		if field.Value != testPass.Cashback {
			t.Errorf("HeaderField value = %q, want %q", field.Value, testPass.Cashback)
		}
	})

	t.Run("Primary fields contain company name", func(t *testing.T) {
		if len(passData.Generic.PrimaryFields) != 1 {
			t.Fatalf("PrimaryFields length = %d, want 1", len(passData.Generic.PrimaryFields))
		}
		field := passData.Generic.PrimaryFields[0]
		if field.Key != "company-name" {
			t.Errorf("PrimaryField key = %q, want %q", field.Key, "company-name")
		}
		if field.Value != testPass.CompanyName {
			t.Errorf("PrimaryField value = %q, want %q", field.Value, testPass.CompanyName)
		}
	})

	t.Run("Secondary fields contain IBAN and BIC", func(t *testing.T) {
		if len(passData.Generic.SecondaryFields) != 2 {
			t.Fatalf("SecondaryFields length = %d, want 2", len(passData.Generic.SecondaryFields))
		}

		// Check IBAN field
		ibanField := passData.Generic.SecondaryFields[0]
		if ibanField.Key != "iban" {
			t.Errorf("IBAN field key = %q, want %q", ibanField.Key, "iban")
		}
		if ibanField.Value != testPass.IBAN {
			t.Errorf("IBAN field value = %q, want %q", ibanField.Value, testPass.IBAN)
		}

		// Check BIC field
		bicField := passData.Generic.SecondaryFields[1]
		if bicField.Key != "bic" {
			t.Errorf("BIC field key = %q, want %q", bicField.Key, "bic")
		}
		if bicField.Value != testPass.BIC {
			t.Errorf("BIC field value = %q, want %q", bicField.Value, testPass.BIC)
		}
	})

	t.Run("Auxiliary fields contain address", func(t *testing.T) {
		if len(passData.Generic.AuxiliaryFields) != 1 {
			t.Fatalf("AuxiliaryFields length = %d, want 1", len(passData.Generic.AuxiliaryFields))
		}
		field := passData.Generic.AuxiliaryFields[0]
		if field.Key != "address" {
			t.Errorf("AuxiliaryField key = %q, want %q", field.Key, "address")
		}
		if field.Value != testPass.Address {
			t.Errorf("AuxiliaryField value = %q, want %q", field.Value, testPass.Address)
		}
	})

	t.Run("Back fields contain metadata", func(t *testing.T) {
		if len(passData.Generic.BackFields) != 3 {
			t.Fatalf("BackFields length = %d, want 3", len(passData.Generic.BackFields))
		}

		// Check for serial number field
		found := false
		for _, field := range passData.Generic.BackFields {
			if field.Key == "serialNumber" {
				found = true
				if field.Value != testPass.ID.String() {
					t.Errorf("SerialNumber back field value = %q, want %q", field.Value, testPass.ID.String())
				}
			}
		}
		if !found {
			t.Error("SerialNumber field not found in back fields")
		}

		// Check for company ID field
		found = false
		for _, field := range passData.Generic.BackFields {
			if field.Key == "companyID" {
				found = true
				if field.Value != testPass.CompanyID {
					t.Errorf("CompanyID back field value = %q, want %q", field.Value, testPass.CompanyID)
				}
			}
		}
		if !found {
			t.Error("CompanyID field not found in back fields")
		}
	})

	t.Run("Barcode contains QR code with correct format", func(t *testing.T) {
		if passData.Barcode.Format != "PKBarcodeFormatQR" {
			t.Errorf("Barcode format = %q, want %q", passData.Barcode.Format, "PKBarcodeFormatQR")
		}
		if passData.Barcode.MessageEncoding != "iso-8859-1" {
			t.Errorf("Barcode encoding = %q, want %q", passData.Barcode.MessageEncoding, "iso-8859-1")
		}

		// Check EPC QR code format
		expectedMessage := "BCD\n001\n1\nSCT\n" + testPass.BIC + "\n" + testPass.CompanyName + "\n" + testPass.IBAN
		if passData.Barcode.Message != expectedMessage {
			t.Errorf("Barcode message = %q, want %q", passData.Barcode.Message, expectedMessage)
		}
	})

	t.Run("Pass structure can be marshaled to JSON", func(t *testing.T) {
		jsonData, err := json.Marshal(passData)
		if err != nil {
			t.Fatalf("Failed to marshal pass to JSON: %v", err)
		}

		// Try to unmarshal it back
		var unmarshaled PassData
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal pass JSON: %v", err)
		}

		// Verify key fields
		if unmarshaled.SerialNumber != passData.SerialNumber {
			t.Errorf("Unmarshaled SerialNumber = %q, want %q", unmarshaled.SerialNumber, passData.SerialNumber)
		}
	})
}

func TestFieldStructure(t *testing.T) {
	t.Run("Field JSON marshaling", func(t *testing.T) {
		field := Field{
			Key:   "test-key",
			Label: "Test Label",
			Value: "Test Value",
		}

		jsonData, err := json.Marshal(field)
		if err != nil {
			t.Fatalf("Failed to marshal field: %v", err)
		}

		var unmarshaled Field
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal field: %v", err)
		}

		if unmarshaled.Key != field.Key {
			t.Errorf("Field key = %q, want %q", unmarshaled.Key, field.Key)
		}
		if unmarshaled.Label != field.Label {
			t.Errorf("Field label = %q, want %q", unmarshaled.Label, field.Label)
		}
		if unmarshaled.Value != field.Value {
			t.Errorf("Field value = %q, want %q", unmarshaled.Value, field.Value)
		}
	})
}

func TestBarcodeStructure(t *testing.T) {
	t.Run("Barcode JSON marshaling", func(t *testing.T) {
		barcode := Barcode{
			Format:          "PKBarcodeFormatQR",
			Message:         "Test message",
			MessageEncoding: "iso-8859-1",
		}

		jsonData, err := json.Marshal(barcode)
		if err != nil {
			t.Fatalf("Failed to marshal barcode: %v", err)
		}

		var unmarshaled Barcode
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal barcode: %v", err)
		}

		if unmarshaled.Format != barcode.Format {
			t.Errorf("Barcode format = %q, want %q", unmarshaled.Format, barcode.Format)
		}
		if unmarshaled.Message != barcode.Message {
			t.Errorf("Barcode message = %q, want %q", unmarshaled.Message, barcode.Message)
		}
		if unmarshaled.MessageEncoding != barcode.MessageEncoding {
			t.Errorf("Barcode encoding = %q, want %q", unmarshaled.MessageEncoding, barcode.MessageEncoding)
		}
	})
}
