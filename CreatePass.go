package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// PassData represents the data structure for pass.json
type Field struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type Generic struct {
	HeaderFields    []Field `json:"headerFields"`
	PrimaryFields   []Field `json:"primaryFields"`
	SecondaryFields []Field `json:"secondaryFields"`
	BackFields      []Field `json:"backFields"`
	AuxiliaryFields []Field `json:"auxiliaryFields"`
}

type Barcode struct {
	Format          string `json:"format"`
	Message         string `json:"message"`
	MessageEncoding string `json:"messageEncoding"`
}

type PassData struct {
	FormatVersion       int     `json:"formatVersion"`
	PassTypeIdentifier  string  `json:"passTypeIdentifier"`
	SerialNumber        string  `json:"serialNumber"`
	WebServiceURL       string  `json:"webServiceURL"`
	AuthenticationToken string  `json:"authenticationToken"`
	TeamIdentifier      string  `json:"teamIdentifier"`
	OrganizationName    string  `json:"organizationName"`
	Description         string  `json:"description"`
	LogoText            string  `json:"logoText"`
	BackgroundColor     string  `json:"backgroundColor"`
	ForegroundColor     string  `json:"foregroundColor"`
	LabelColor          string  `json:"labelColor"`
	Generic             Generic `json:"generic"`
	Barcode             Barcode `json:"barcode"`
}

// CreatePass generates the necessary directories and files for a pass
func CreatePass() error {
	pass := PassData{
		FormatVersion:       1,
		PassTypeIdentifier:  "pass.com.finom.bank2wallet",
		SerialNumber:        "8j23fm3",
		WebServiceURL:       "https://finom.co/passes/",
		AuthenticationToken: "vxwxd7J8AlNNFPS8k0a0FfUFtq0ewzFdc",
		TeamIdentifier:      "35XPTK6L36",
		OrganizationName:    "Finom",
		Description:         "Your bank details in Finom",
		LogoText:            "Your Bank Details",
		BackgroundColor:     "rgb(255, 76, 92)",
		ForegroundColor:     "rgb(255, 255, 255)",
		LabelColor:          "rgb(11, 0, 46)",
		Generic: Generic{
			HeaderFields: []Field{
				{
					Key:   "plan",
					Label: "PLAN",
					Value: "Premium",
				},
			},
			PrimaryFields: []Field{
				{
					Key:   "company-name",
					Value: "Ivo Dimitrov",
				},
			},
			SecondaryFields: []Field{
				{
					Key:   "iban",
					Label: "IBAN",
					Value: "NL24 FNOM 0698 9885 95",
				},
				{
					Key:   "bic",
					Label: "BIC",
					Value: "FNOMNL22",
				},
			},
			AuxiliaryFields: []Field{
				{
					Key:   "address",
					Label: "ADDRESS",
					Value: "Kastanienallee 102, 10435, Berlin, Germany",
				},
			},
			BackFields: []Field{
				{
					Key:   "info",
					Label: "Additional Information",
					Value: "This pass contains your bank credentials in Finom and is valid for SEPA payments only. \nGo to https://finom.co/passes/ for more information.",
				},
			},
		},
		Barcode: Barcode{
			Format:          "PKBarcodeFormatQR",
			Message:         "BCD\n001\n1\nSCT\nFNOMNL22\nIvo Dimitrov\nNL24FNOM0698988595",
			MessageEncoding: "iso-8859-1",
		},
	}

	passName := "Ivo_Dimitrov" + ".pass"

	// Directories and files to create
	dirs := []string{passName}
	// files := []string{"icon.png", "logo.png", "strip.png"}

	// Create directories
	for _, dir := range dirs {
		if err := os.MkdirAll("./tmp/"+dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %v", dir, err)
		}
	}

	// Create pass.json
	passJSON, err := json.MarshalIndent(pass, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling pass.json: %v", err)
	}
	passFilePath := "./tmp/" + passName + "/pass.json"
	if err := os.WriteFile(passFilePath, passJSON, 0644); err != nil {
		return fmt.Errorf("error writing pass.json: %v", err)
	}

	// Move images from template directory to pass directory
	err = copyImages("./template", "./tmp/"+passName)
	if err != nil {
		log.Printf("error copying images: %v", err)
	}

	return nil
}

func copyImages(srcDir, dstDir string) error {
	// Ensure destination directory exists.
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// Get list of files in source directory.
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	// Iterate over each file.
	for _, file := range files {
		// Skip directories.
		if file.IsDir() {
			continue
		}

		// Open source file.
		srcFile, err := os.Open(filepath.Join(srcDir, file.Name()))
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Create destination file.
		dstFile, err := os.Create(filepath.Join(dstDir, file.Name()))
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// Copy content from source file to destination file.
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}
	}

	return nil
}
