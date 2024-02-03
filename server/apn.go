package main

import (
	"github.com/rs/zerolog/log"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

func SendNotificationPushAboutUpdate() {

	cert, err := certificate.FromP12File("./certificates/Certificates.p12", "")
	if err != nil {
		log.Error().
			Err(err).
			Msg("Push Certificate Error")
	}

	notification := &apns2.Notification{}
	notification.DeviceToken = "b820c3afeda0a58ae605bfb0b9a29c5c97c74b4d21d6fa32b0ead17c1e917fc2"
	notification.Topic = "pass.com.finom.bank2wallet"
	notification.Payload = []byte(`{"aps":{"alert":"Your cashback balance was updated!"}}`)

	// If you want to test push notifications for builds running directly from XCode (Development), use
	// client := apns2.NewClient(cert).Development()
	// For apps published to the app store or installed as an ad-hoc distribution use Production()

	client := apns2.NewClient(cert).Production()
	res, err := client.Push(notification)

	if err != nil {
		log.Error().
			Err(err).
			Msg("Error sending push notification")
	}

	log.Debug().
		Interface("Result", res).
		Msg("Notification result")
}
