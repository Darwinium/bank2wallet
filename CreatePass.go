package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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
func CreatePass(plan, companyName, iban, bic, address string) (string, error) {
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
					Value: plan,
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
	}

	passName := sanitizeCompanyName(companyName)

	// Directories and files to create
	dirs := []string{passName + ".pass"}
	manifest := make(map[string]string)

	// Create directories
	for _, dir := range dirs {
		if err := os.MkdirAll("./b2wData/tmp/"+dir, 0755); err != nil {
			return "", fmt.Errorf("error creating directory %s: %v", dir, err)
		}
	}

	// Create pass.json
	passJSON, err := json.MarshalIndent(pass, "", " ")
	if err != nil {
		return "", fmt.Errorf("error marshalling pass.json: %v", err)
	}
	passFilePath := "./b2wData/tmp/" + passName + ".pass" + "/pass.json"
	if err := os.WriteFile(passFilePath, passJSON, 0644); err != nil {
		return "", fmt.Errorf("error writing pass.json: %v", err)
	}
	manifest["pass.json"] = sha1Hash(passJSON)

	// Move images from template directory to pass directory
	imageManifest, err := copyImages("./template", "./b2wData/tmp/"+passName+".pass")
	if err != nil {
		return "", fmt.Errorf("error copying images: %v", err)
	}

	manifest = mergeMaps(manifest, imageManifest)
	// Create manifest.json
	manifestJSON, err := json.MarshalIndent(manifest, "", " ")
	if err != nil {
		return "", fmt.Errorf("error marshalling manifest.json: %v", err)
	}
	if err := os.WriteFile("./b2wData/tmp/"+passName+".pass/manifest.json", manifestJSON, 0644); err != nil {
		return "", fmt.Errorf("error writing manifest.json: %v", err)
	}

	//Sign the pass
	err = signingPass(passName)
	if err != nil {
		return "", fmt.Errorf("error signing pass: %v", err)
	}

	pkpassName, err := createPKPass(passName)
	if err != nil {
		return "", fmt.Errorf("error creating pkpass: %v", err)
	}

	err = os.RemoveAll("./b2wData/tmp/" + passName + ".pass")
	if err != nil {
		log.Printf("error removing tmp directory: %v", err)
	}

	return pkpassName, nil
}

func sanitizeCompanyName(companyName string) string {
	// Remove all non-alphanumeric characters.
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	sanitized := reg.ReplaceAllString(companyName, "")

	// Remove spaces.
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	// Truncate to 15 characters.
	if len(sanitized) > 15 {
		sanitized = sanitized[:15]
	}

	return sanitized
}

func copyImages(srcDir, dstDir string) (map[string]string, error) {
	// Ensure destination directory exists.
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return nil, err
	}

	// Get list of files in source directory.
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, err
	}

	// Manifest for images
	manifest := make(map[string]string)

	// Iterate over each file.
	for _, file := range files {
		// Skip directories.
		if file.IsDir() {
			continue
		}

		// Open source file.
		srcFile, err := os.Open(filepath.Join(srcDir, file.Name()))
		if err != nil {
			return nil, err
		}
		defer srcFile.Close()

		// Create destination file.
		newFilePath := filepath.Join(dstDir, file.Name())
		dstFile, err := os.Create(newFilePath)
		if err != nil {
			return nil, err
		}
		defer dstFile.Close()

		// Copy content from source file to destination file.
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return nil, err
		}

		// Read the content of the destination file.
		dstContent, err := os.ReadFile(newFilePath)
		if err != nil {
			return nil, err
		}

		manifest[file.Name()] = sha1Hash(dstContent)
	}

	return manifest, nil
}

func signingPass(passName string) error {
	CERT_PASSWORD := os.Getenv("CERT_PASSWORD")

	cmd := exec.Command("openssl", "smime", "-binary", "-sign", "-certfile", "./certificates/WWDR.pem", "-signer", "./certificates/passcertificate.pem", "-inkey", "./certificates/passkey.pem", "-in", "./b2wData/tmp/"+passName+".pass/manifest.json", "-out", "./b2wData/tmp/"+passName+".pass/signature", "-outform", "DER", "-passin", "pass:"+CERT_PASSWORD)
	log.Println(cmd.Path)
	err := cmd.Run()
	if err != nil {
		log.Println("Error executing OpenSSL command: ", err)
		return err
	}
	log.Printf("Signing of the pass %s executed successfully\n", passName)
	return nil

}

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

// sha1Hash returns the SHA1 hash of the given data as a hex string
func sha1Hash(data []byte) string {
	hash := sha1.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func mergeMaps(m1 map[string]string, m2 map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}
