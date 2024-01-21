package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	serverURL := "0.0.0.0:" + *port

	r := gin.Default()
	r.StaticFS("/passes", gin.Dir("./b2wData/passes", false))

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

	r.POST("/create", func(c *gin.Context) {
		plan := c.PostForm("plan")
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
