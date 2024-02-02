package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const (
	TemplateDir     = "./template"      // Directory with the template images
	TempDir         = "./b2wData/tmp/"  // Directory to store the temporary pass files
	PassesDir       = "./passes/"       // Directory to store the generated pkpass files
	CertificatesDir = "./certificates/" // Directory with the certificates
)

// Field represents a field in the pass
type Field struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// Generic represents the generic type of pass
type Generic struct {
	HeaderFields    []Field `json:"headerFields"`
	PrimaryFields   []Field `json:"primaryFields"`
	SecondaryFields []Field `json:"secondaryFields"`
	BackFields      []Field `json:"backFields"`
	AuxiliaryFields []Field `json:"auxiliaryFields"`
}

// Barcode represents the barcode of the pass
type Barcode struct {
	Format          string `json:"format"`
	Message         string `json:"message"`
	MessageEncoding string `json:"messageEncoding"`
}

// PassData represents the data itself in the pass
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

// NewPassData creates a new pass with the given data and saves it in the database. It returns the pass data
func NewPassData(companyID, cashback, companyName, iban, bic, address string) (PassData, error) {
	db, err := getDBConnection()
	if err != nil {
		return PassData{}, err
	}
	newPass, err := AddNewPass(db, companyID, cashback, companyName, iban, bic, address)
	if err != nil {
		return PassData{}, fmt.Errorf("error adding new pass: %v", err)
	}
	serialNumber := newPass.ID.String()

	return PassData{
		FormatVersion:       1,
		PassTypeIdentifier:  "pass.com.finom.bank2wallet",
		SerialNumber:        serialNumber,
		WebServiceURL:       "https://finom.co/passes/",
		AuthenticationToken: "vxwxd7J8AlNNFPS8k0a0FfUFtq0ewzFdc",
		TeamIdentifier:      "35XPTK6L36",
		OrganizationName:    "Finom",
		Description:         "Your Finom Bank Details",
		LogoText:            "Your Bank Details",
		BackgroundColor:     "rgb(255, 76, 92)",
		ForegroundColor:     "rgb(255, 255, 255)",
		LabelColor:          "rgb(11, 0, 46)",
		Generic: Generic{
			HeaderFields: []Field{
				{
					Key:   "cashback",
					Label: "CASHBACK",
					Value: cashback,
				},
			},
			PrimaryFields: []Field{
				{
					Key:   "company-name",
					Value: companyName,
				},
			},
			SecondaryFields: []Field{
				{
					Key:   "iban",
					Label: "IBAN",
					Value: iban,
				},
				{
					Key:   "bic",
					Label: "BIC",
					Value: bic,
				},
			},
			AuxiliaryFields: []Field{
				{
					Key:   "address",
					Label: "ADDRESS",
					Value: address,
				},
			},
			BackFields: []Field{
				{
					Key:   "serialNumber",
					Label: "Serial Number",
					Value: serialNumber,
				},
				{
					Key:   "companyID",
					Label: "Company ID",
					Value: companyID,
				},
				{
					Key:   "info",
					Label: "Additional Information",
					Value: "This pass contains your bank credentials in Finom and is valid for SEPA payments only. \nGo to https://finom.co/passes/ for more information.",
				},
			},
		},
		Barcode: Barcode{
			Format:          "PKBarcodeFormatQR",
			Message:         "BCD\n001\n1\nSCT\n" + bic + "\n" + companyName + "\n" + iban,
			MessageEncoding: "iso-8859-1",
		},
	}, nil
}

// CreatePass generates the necessary directories and files for a pass. It returns the name of the pass file in the pkpass format
func CreatePass(companyID, cashback, companyName, iban, bic, address string) (string, error) {
	pass, err := NewPassData(companyID, cashback, companyName, iban, bic, address)
	if err != nil {
		return "", err
	}

	passName := SanitizeText(companyName)

	// Directories and files to create
	if err := CreateDir(TempDir + passName + ".pass"); err != nil {
		return "", err
	}

	// Create pass.json
	passJSON, err := json.MarshalIndent(pass, "", " ")
	if err != nil {
		return "", fmt.Errorf("error marshalling pass.json: %v", err)
	}
	passFilePath := TempDir + passName + ".pass" + "/pass.json"
	if err := os.WriteFile(passFilePath, passJSON, 0644); err != nil {
		return "", fmt.Errorf("error writing pass.json: %v", err)
	}
	manifest := make(map[string]string)
	manifest["pass.json"] = Sha1Hash(passJSON)

	// Move images from template directory to pass directory
	imageManifest, err := CopyImages(TemplateDir, TempDir+passName+".pass")
	if err != nil {
		return "", fmt.Errorf("error copying images: %v", err)
	}

	// Create manifest.json
	manifest = MergeMaps(manifest, imageManifest)
	manifestJSON, err := json.MarshalIndent(manifest, "", " ")
	if err != nil {
		return "", fmt.Errorf("error marshalling manifest.json: %v", err)
	}
	if err := os.WriteFile(TempDir+passName+".pass/manifest.json", manifestJSON, 0644); err != nil {
		return "", fmt.Errorf("error writing manifest.json: %v", err)
	}

	//Sign the pass
	err = signingPass(passName)
	if err != nil {
		return "", fmt.Errorf("error signing pass: %v", err)
	}

	// Create pkpass
	pkpassName, err := createPKPass(passName)
	if err != nil {
		return "", fmt.Errorf("error creating pkpass: %v", err)
	}

	// Remove tmp directory
	err = os.RemoveAll(TempDir + passName + ".pass")
	if err != nil {
		log.Printf("error removing tmp directory: %v", err)
	}

	return pkpassName, nil
}

// signingPass signs the pass with the certificates
func signingPass(passName string) error {
	CERT_PASSWORD := os.Getenv("CERT_PASSWORD")
	cmd := exec.Command("openssl", "smime", "-binary", "-sign", "-certfile", CertificatesDir+"WWDR.pem", "-signer", CertificatesDir+"passcertificate.pem", "-inkey", CertificatesDir+"passkey.pem", "-in", TempDir+passName+".pass/manifest.json", "-out", TempDir+passName+".pass/signature", "-outform", "DER", "-passin", "pass:"+CERT_PASSWORD)
	log.Println(cmd.Path)
	err := cmd.Run()
	if err != nil {
		log.Println("Error executing OpenSSL command: ", err)
		return err
	}
	log.Printf("Signing of the pass %s executed successfully\n", passName)
	return nil

}

// createPKPass creates the pkpass file. It returns the name of the pkpass file
func createPKPass(passName string) (string, error) {
	// Change working directory
	err := os.Chdir("./b2wData/tmp/" + passName + ".pass")
	if err != nil {
		log.Fatalf("os.Chdir() failed with %s\n", err)
		return "", err
	}

	cmd := exec.Command("zip", "-r", "../../passes/"+passName+".pkpass", ".")
	err = cmd.Run()
	if err != nil {
		log.Println("Error executing ZIP command: ", err)
		return "", err
	}
	os.Chdir("../../../")
	log.Printf("Creation of the pkpass %s executed successfully\n", passName)
	return passName + ".pkpass", nil
}
