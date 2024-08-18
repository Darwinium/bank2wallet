package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Settings struct {
	host   string
	port   string
	apiKey string
}

var (
	settings Settings
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if os.Getenv("APP_ENV") != "release" {
		err := godotenv.Load()
		log.Info().Msg("Development mode. Set APP_ENV=release to load environment variables from system and not from file.")
		if err != nil {
			log.Fatal().Msg("Error loading .env file")
		} else {
			log.Info().Msg("Loaded .env file successfully")
		}
	} else {
		log.Info().Msg("Production mode. Loading environment variables from system.")
	}

	settings = Settings{
		host:   os.Getenv("HOST"),
		port:   os.Getenv("PORT"),
		apiKey: os.Getenv("API_KEY"),
	}
}

func main() {
	// Components routing:
	app.Route("/", &CreatePass{})
	app.RunWhenOnBrowser()

	// HTTP routing:
	http.Handle("/", &app.Handler{
		Name:        "Finom Bank2Wallet",
		Description: "Create a pass for Apple Wallet.",
	})

	serverString := settings.host + ":" + settings.port

	log.Info().
		Msgf("Server started on %s\n", serverString)

	if err := http.ListenAndServe(serverString, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to start the server")
	}
}
