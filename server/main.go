package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var db *gorm.DB
var serverURL string

type pushTokenRequest struct {
	PushToken string `json:"pushToken"`
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	err := godotenv.Load()
	if err != nil {
		log.Fatal().Msg("Error loading .env file")
	} else {
		log.Info().Msg("Loaded .env file successfully")
	}

	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()
	serverURL = os.Getenv("SERVER_URL") + ":" + *port

	db, err = getDBConnection()
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to the database")
	} else {
		log.Info().Msg("Connected to the database successfully")
	}
}

func main() {
	r := gin.Default()
	// Configuring CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Be careful with this in production
		// AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true // allow any origin or you could put your specific one
		},
		MaxAge: 12 * time.Hour,
	}))

	r.StaticFS("/passes", gin.Dir("./b2wData/passes", false))

	r.POST("pass/v1/create", AuthRequired(), createPass)
	r.POST("pass/v1/getPass", AuthRequired(), getPass)
	r.POST("pass/v1/updateCashback", AuthRequired(), updateCashback)

	// --- Apple Wallet Requests BEGIN --- //
	r.POST("/pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier/:serialNumber", AuthRequired(), registerDeviceRequest)
	r.GET("/pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier", checkPassUpdatesRequest)
	r.GET("/pass/v1/registerDevice/v1/passes/:passTeamIdentifier/:serialNumber", AuthRequired(), getUpdatedPass)
	r.DELETE("/pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier/:serialNumber", deletePassRequest)
	r.POST("/pass/v1/registerDevice/v1/log", logRequest)
	// --- Apple Wallet Requests END --- //

	if err := r.Run(serverURL); err != nil {
		log.Fatal().Err(err).Msg("Server run failed")
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "ApplePass ")

		if token != os.Getenv("AUTH_TOKEN") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func createPass(c *gin.Context) {
	companyID := c.PostForm("companyID")
	cashback := c.PostForm("cashback") + "€"
	log.Debug().Any("Request", c.Request.MultipartForm)
	companyName := c.PostForm("companyName")
	iban := c.PostForm("iban")
	bic := c.PostForm("bic")
	address := c.PostForm("address")

	missingFields := []string{}
	if companyID == "" {
		missingFields = append(missingFields, "companyID")
	}
	if cashback == "" {
		cashback = "0" + "€"
	}
	if companyName == "" {
		missingFields = append(missingFields, "companyName")
	}
	if iban == "" {
		missingFields = append(missingFields, "iban")
	}
	if bic == "" {
		missingFields = append(missingFields, "bic")
	}
	if address == "" {
		missingFields = append(missingFields, "address")
	}

	if len(missingFields) > 0 {
		c.JSON(400, gin.H{
			"message": "Missing required fields",
			"fields":  missingFields,
		})
		return
	}

	pass, err := GeneratePass(
		db,
		companyID,
		cashback,
		companyName,
		iban,
		bic,
		address,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create pass")
		c.JSON(500, gin.H{
			"message":   "Failed to create pass",
			"error":     err.Error(),
			"companyID": companyID,
		})
		return
	}

	pkpassFilePath := serverURL + "/passes/" + pass.ID.String() + ".pkpass"
	log.Debug().Msgf("Pass was created successfully!\nLink: %s\n", pkpassFilePath)

	c.JSON(200, gin.H{
		"message":   "Pass was created successfully",
		"link":      pkpassFilePath,
		"companyID": pass.CompanyID,
		"passID":    pass.ID,
	})
}

func getPass(c *gin.Context) {
	companyID := c.PostForm("companyID")

	if len(companyID) == 0 {
		c.JSON(400, gin.H{
			"message": "Missing required fields",
			"fields":  "companyID",
		})
		return
	}

	pass, err := GetPassByCompanyID(db, companyID)
	if err != nil {
		c.JSON(500, gin.H{
			"message":   "Failed to get pass",
			"error":     err.Error(),
			"companyID": companyID,
		})
		return
	}

	c.JSON(200, gin.H{
		"message":   "Pass was retrieved successfully",
		"link":      serverURL + "/passes/" + pass.ID.String() + ".pkpass",
		"companyID": companyID,
		"passID":    pass.ID,
	})

}

func updateCashback(c *gin.Context) {
	companyID := c.PostForm("companyID")
	cashback := c.PostForm("cashback")

	missingFields := []string{}
	if companyID == "" {
		missingFields = append(missingFields, "companyID")
	}
	if cashback == "" {
		missingFields = append(missingFields, "cashback")
	}

	if len(missingFields) > 0 {
		c.JSON(400, gin.H{
			"message": "Missing required fields during the update of cashback",
			"fields":  missingFields,
		})
		return
	} else {
		cashback += "€"
	}

	pass, err := UpdatePassByCompanyID(db, companyID, cashback)
	if err != nil {
		c.JSON(500, gin.H{
			"message":   "Failed to get pass",
			"error":     err.Error(),
			"companyID": companyID,
		})
		return
	}

	// TODO: Add generating updated file
	pass, err = GeneratePass(
		db,
		pass.CompanyID,
		pass.Cashback,
		pass.CompanyName,
		pass.IBAN,
		pass.BIC,
		pass.Address,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate new pass")
	}

	SendNotificationPushAboutUpdate()

	c.JSON(200, gin.H{
		"message":   "Cashback was updated successfully",
		"link":      serverURL + "/passes/" + pass.ID.String() + ".pkpass",
		"companyID": companyID,
	})
}

// --- Apple Wallet Requests BEGIN --- //

func registerDeviceRequest(c *gin.Context) {
	deviceLibraryIdentifier := c.Param("deviceLibraryIdentifier")
	serialNumber := c.Param("serialNumber")

	var req pushTokenRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pushToken := req.PushToken

	deviceReg, err, exists := RegisterDevice(db, deviceLibraryIdentifier, serialNumber, pushToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info().
		Str("SerialNumber", deviceReg.SerialNumber).
		Str("PushToken", deviceReg.PushToken).
		Msg("Registration of the new pass")

	if exists {
		c.JSON(200, gin.H{})
	} else {
		c.JSON(201, gin.H{})
	}
}

func checkPassUpdatesRequest(c *gin.Context) {
	log.Info().
		Interface("Query", c.Request.URL.Query()).
		Msg("Request to check updates")

	previousLastUpdated := c.Query("passesUpdatedSince")
	deviceLibraryIdentifier := c.Param("deviceLibraryIdentifier")

	var (
		serialNumbers []string
		err           error
	)

	// Check if it is the first request from the device
	if len(previousLastUpdated) == 0 {
		// Get all passes for the device
		serialNumbers, err = GetPassesByDeviceID(db, deviceLibraryIdentifier)
	} else {
		// Get updated passes for the device
		serialNumbers, err = GetUpdatedPasses(db, deviceLibraryIdentifier, previousLastUpdated)
	}

	if err != nil {
		log.Error().Err(err)
		return
	}

	if len(serialNumbers) == 0 {
		log.Info().
			Str("DeviceLibraryIdentifier", deviceLibraryIdentifier).
			Msg("No matching passes found for the device")

		// 204 — No Matching Passes
		c.JSON(204, gin.H{})
		return
	}

	log.Info().
		Str("DeviceLibraryIdentifier", deviceLibraryIdentifier).
		Msg("Matching passes found for the device")

	// Update the lastUpdated timestamp of the device
	lastUpdated := time.Now().UTC().Format(time.RFC3339)
	response := gin.H{
		"lastUpdated":   lastUpdated,
		"serialNumbers": serialNumbers,
	}
	log.Debug().
		Interface("Response", response).
		Str("DeviceLibraryIdentifier", deviceLibraryIdentifier).
		Str("LastUpdated", lastUpdated)

	// 200 — Matching Passes Found
	c.JSON(200, response)
}

func getUpdatedPass(c *gin.Context) {
	serialNumber := c.Param("serialNumber")
	log.Info().
		Str("SerialNumber", serialNumber).
		Msg("Request for updated pass")

	filePath := "./b2wData/passes/" + serialNumber + ".pkpass"
	// Read the .pkpass file content
	pkpassContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Error().Msgf("Failed to read .pkpass file: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	// Set the correct Content-Type headers
	lastModifiedTime := time.Now().UTC().Format(http.TimeFormat)
	c.Header("Last-Modified", lastModifiedTime)
	c.Header("Content-Type", "application/vnd.apple.pkpass")
	// Send the .pkpass file content as the response
	c.Writer.Write(pkpassContent)
}

func deletePassRequest(c *gin.Context) {
	serialNumber := c.Param("serialNumber")
	err := DeletePassOnDevice(db, serialNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}

	log.Info().
		Str("SerialNumber", serialNumber).
		Msg("Pass was unregistered")
}

func logRequest(c *gin.Context) {
	log.Debug().Msgf("Request: %v\n", ReadRequestBody(c.Request.Body))
	c.JSON(http.StatusOK, gin.H{})
}

// --- Apple Wallet Requests END --- //
