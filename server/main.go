package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	serverURL := "0.0.0.0:" + *port

	r := gin.Default()
	r.StaticFS("/passes", gin.Dir("./b2wData/passes", false))

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

	r.GET("/test", func(c *gin.Context) {
		pkpassName, err := CreatePass(
			"Premium",
			"Ivo Dimitrov Super Puper&&& Company",
			"NL24 FNOM 0698 9885 95",
			"FNOMNL22",
			"Kastanienallee 99, 10435, Berlin, Germany",
		)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Failed to create pass",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Pass was created successfully",
			"link":    serverURL + "/passes/" + pkpassName,
		})
	})

	r.POST("/create", AuthRequired(), func(c *gin.Context) {
		plan := c.PostForm("plan")
		log.Println(c.Request.MultipartForm)
		companyName := c.PostForm("companyName")
		iban := c.PostForm("iban")
		bic := c.PostForm("bic")
		address := c.PostForm("address")

		missingFields := []string{}
		if plan == "" {
			missingFields = append(missingFields, "plan")
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

		pkpassName, err := CreatePass(
			plan,
			companyName,
			iban,
			bic,
			address,
		)
		if err != nil {
			log.Println("Failed to create pass:", err)
			c.JSON(500, gin.H{
				"message": "Failed to create pass",
				"error":   err.Error(),
			})
			return
		}

		log.Println("Pass was created successfully", "link", serverURL+"/passes/"+pkpassName)

		c.JSON(200, gin.H{
			"message": "Pass was created successfully",
			"link":    serverURL + "/passes/" + pkpassName,
		})
	})

	if err := r.Run(serverURL); err != nil {
		log.Fatal("Server run failed:", err)
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		// Here we are just checking if the token is what we expect. In a real-world application,
		// you would probably use a more sophisticated way to validate the token, like JWT.
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
