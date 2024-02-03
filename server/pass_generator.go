package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
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

// CreatePassStructure creates the structure of the pass card with the given data
func CreatePassStructure(pass Pass) PassData {

	serialNumber := pass.ID.String()

	return PassData{
		FormatVersion:       1,
		PassTypeIdentifier:  "pass.com.finom.bank2wallet",
		SerialNumber:        serialNumber,
		WebServiceURL:       "https://creative-smoothly-cockatoo.ngrok-free.app/pass/v1/registerDevice",
		AuthenticationToken: os.Getenv("AUTH_TOKEN"),
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
					Value: pass.Cashback,
				},
			},
			PrimaryFields: []Field{
				{
					Key:   "company-name",
					Value: pass.CompanyName,
				},
			},
			SecondaryFields: []Field{
				{
					Key:   "iban",
					Label: "IBAN",
					Value: pass.IBAN,
				},
				{
					Key:   "bic",
					Label: "BIC",
					Value: pass.BIC,
				},
			},
			AuxiliaryFields: []Field{
				{
					Key:   "address",
					Label: "ADDRESS",
					Value: pass.Address,
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
					Value: pass.CompanyID,
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
			Message:         "BCD\n001\n1\nSCT\n" + pass.BIC + "\n" + pass.CompanyName + "\n" + pass.IBAN,
			MessageEncoding: "iso-8859-1",
		},
	}
}

// CreatePass generates the necessary directories and files for a pass. It returns the name of the pass file in the pkpass format
func GeneratePass(db *gorm.DB, companyID, cashback, companyName, iban, bic, address string) (Pass, error) {
	passDB, err := AddNewPass(db, companyID, cashback, companyName, iban, bic, address)
	if err != nil {
		return Pass{}, fmt.Errorf("error adding new pass: %v", err)
	}

	passCard := CreatePassStructure(passDB)
	passName := passDB.ID.String()

	// Directories and files to create
	if err := CreateDir(TempDir + passName + ".pass"); err != nil {
		return Pass{}, err
	}

	// Create pass.json
	passJSON, err := json.MarshalIndent(passCard, "", " ")
	if err != nil {
		return Pass{}, fmt.Errorf("error marshalling pass.json: %v", err)
	}
	passFilePath := TempDir + passName + ".pass" + "/pass.json"
	if err := os.WriteFile(passFilePath, passJSON, 0644); err != nil {
		return Pass{}, fmt.Errorf("error writing pass.json: %v", err)
	}
	manifest := make(map[string]string)
	manifest["pass.json"] = Sha1Hash(passJSON)

	// Move images from template directory to pass directory
	imageManifest, err := CopyImages(TemplateDir, TempDir+passName+".pass")
	if err != nil {
		return Pass{}, fmt.Errorf("error copying images: %v", err)
	}

	// Create manifest.json
	manifest = MergeMaps(manifest, imageManifest)
	manifestJSON, err := json.MarshalIndent(manifest, "", " ")
	if err != nil {
		return Pass{}, fmt.Errorf("error marshalling manifest.json: %v", err)
	}
	if err := os.WriteFile(TempDir+passName+".pass/manifest.json", manifestJSON, 0644); err != nil {
		return Pass{}, fmt.Errorf("error writing manifest.json: %v", err)
	}

	//Sign the pass
	err = signingPassFile(passName)
	if err != nil {
		return Pass{}, fmt.Errorf("error signing pass: %v", err)
	}

	// Create pkpass
	err = createPKPassFile(passName)
	if err != nil {
		return Pass{}, fmt.Errorf("error creating pkpass: %v", err)
	}

	// Remove tmp directory of the pass
	err = os.RemoveAll(TempDir + passName + ".pass")
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error removing tmp directory")
	}

	return passDB, nil
}

// signingPassFile signs the pass with the certificates
func signingPassFile(passName string) error {
	CERT_PASSWORD := os.Getenv("CERT_PASSWORD")
	cmd := exec.Command("openssl", "smime", "-binary", "-sign", "-certfile", CertificatesDir+"WWDR.pem", "-signer", CertificatesDir+"passcertificate.pem", "-inkey", CertificatesDir+"passkey.pem", "-in", TempDir+passName+".pass/manifest.json", "-out", TempDir+passName+".pass/signature", "-outform", "DER", "-passin", "pass:"+CERT_PASSWORD)

	log.Debug().
		Str("Command", cmd.Path).
		Strs("Args", cmd.Args).
		Msg("Signing pass with OpenSSL command")

	err := cmd.Run()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error executing OpenSSL command")
		return err
	}
	log.Debug().
		Str("passName", passName).
		Msg("Signing of the pass executed successfully")
	return nil

}

// createPKPassFile creates the pkpass file. It returns the name of the pkpass file
func createPKPassFile(passName string) error {
	// Change working directory
	err := os.Chdir("./b2wData/tmp/" + passName + ".pass")
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error changing working directory")
		return err
	}

	cmd := exec.Command("zip", "-r", "../../passes/"+passName+".pkpass", ".")
	err = cmd.Run()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error executing ZIP command")
		return err
	}
	os.Chdir("../../../")
	log.Debug().
		Str("passName", passName).
		Msg("PKPass file created successfully")
	return nil
}
