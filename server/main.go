package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} else {
		log.Println("Loaded .env file successfully")
	}

	db, err = getDBConnection()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	} else {
		log.Println("Connected to the database successfully")
	}
}

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	serverURL := "localhost:" + *port

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

	r.POST("pass/v1/create", AuthRequired(), func(c *gin.Context) {
		companyID := c.PostForm("companyID")
		cashback := c.PostForm("cashback") + "€"
		log.Println(c.Request.MultipartForm)
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
			log.Println("Failed to create pass:", err)
			c.JSON(500, gin.H{
				"message":   "Failed to create pass",
				"error":     err.Error(),
				"companyID": companyID,
			})
			return
		}

		pkpassFilePath := serverURL + "/passes/" + pass.ID.String() + ".pkpass"

		log.Printf("Pass was created successfully!\nLink: %s\n", pkpassFilePath)

		c.JSON(200, gin.H{
			"message":   "Pass was created successfully",
			"link":      pkpassFilePath,
			"companyID": pass.CompanyID,
			"passID":    pass.ID,
		})
	})

	r.POST("pass/v1/getPass", AuthRequired(), func(c *gin.Context) {
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

	})

	r.POST("pass/v1/updateCashback", AuthRequired(), func(c *gin.Context) {
		companyID := c.PostForm("companyID")
		cashback := c.PostForm("cashback") + "€"

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
		SendNotificationPushAboutUpdate()

		c.JSON(200, gin.H{
			"message":   "Cashback was updated successfully",
			"link":      serverURL + "/passes/" + pass.ID.String() + ".pkpass",
			"companyID": companyID,
		})
	})

	// --- Apple Wallet Requests BEGIN --- //
	r.POST("/pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier/:serialNumber", AuthRequired(), registerDeviceRequest)
	r.GET("/pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier", checkPassUpdatesRequest)
	r.GET("/pass/v1/registerDevice/v1/passes/:passTeamIdentifier/:serialNumber", AuthRequired(), getUpdatedPasses)
	r.DELETE("/pass/v1/registerDevice/v1/devices/:deviceLibraryIdentifier/registrations/:passTeamIdentifier/:serialNumber", deletePassRequest)
	r.POST("/pass/v1/registerDevice/v1/log", logRequest)
	// --- Apple Wallet Requests END --- //

	if err := r.Run(serverURL); err != nil {
		log.Fatal("Server run failed:", err)
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

// --- Apple Wallet Requests BEGIN --- //

type pushTokenRequest struct {
	PushToken string `json:"pushToken"`
}

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

	log.Printf("New pass registered for notifications. SerialNumber: %v, PushToken: %v\n", deviceReg.SerialNumber, deviceReg.PushToken)

	if exists {
		c.JSON(200, gin.H{})
	} else {
		c.JSON(201, gin.H{})
	}
}

func checkPassUpdatesRequest(c *gin.Context) {
	log.Printf("Request if there's any updates. Query: %v\n", c.Request.URL.Query())

	previousLastUpdated := c.Query("passesUpdatedSince")
	deviceLibraryIdentifier := c.Param("deviceLibraryIdentifier")

	if len(previousLastUpdated) == 0 {
		passesSNByDevice, err := GetPassesByDevice(db, deviceLibraryIdentifier)
		if err != nil {
			log.Printf("Error getting passes by device: %v\n", err)
		} else {
			serialNumbers := make([]string, 0, len(passesSNByDevice))
			// Loop through the passes and extract the serial number from each
			for _, pass := range passesSNByDevice {
				serialNumbers = append(serialNumbers, pass.SerialNumber)
			}

			lastUpdated := time.Now().UTC().Format(time.RFC3339)
			// Construct the response object
			response := gin.H{
				"lastUpdated":   lastUpdated, // Use the actual last update timestamp of your passes here
				"serialNumbers": serialNumbers,
			}
			log.Println(response)
			c.JSON(200, response)
		}
	} else {
		updatedPasses, err := CheckPassUpdatesRequest(db, deviceLibraryIdentifier, previousLastUpdated)
		if err != nil {
			log.Println(err)
		}

		if len(updatedPasses) == 0 {
			log.Println("No Matching Passes")
			c.JSON(204, gin.H{})
		} else {
			log.Println("Matching Passes Found")
			serialNumbers := make([]string, 0, len(updatedPasses))
			// Loop through the passes and extract the serial number from each
			for _, pass := range updatedPasses {
				serialNumbers = append(serialNumbers, pass.ID.String())
			}
			lastUpdated := time.Now().UTC().Format(time.RFC3339)
			// Construct the response object
			response := gin.H{
				"lastUpdated":   lastUpdated, // Use the actual last update timestamp of your passes here
				"serialNumbers": serialNumbers,
			}
			log.Println(response)
			c.JSON(200, response)
		}
	}
}

func getUpdatedPasses(c *gin.Context) {
	serialNumber := c.Param("serialNumber")
	log.Println("Request for pass with serial number:", serialNumber)

	filePath := "./b2wData/passes/" + serialNumber + ".pkpass"
	// Read the .pkpass file content
	pkpassContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read .pkpass file: %v", err)
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

	log.Printf("Pass on device was unregistered. SerialNumber: %v\n", serialNumber)
}

func logRequest(c *gin.Context) {
	log.Println(ReadRequestBody(c.Request.Body))
	c.JSON(http.StatusOK, gin.H{})
}

// --- Apple Wallet Requests END --- //
