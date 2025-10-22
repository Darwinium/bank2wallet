package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"gorm.io/gorm"
)

// SendNotificationPushAboutUpdate sends push notifications to all registered devices about pass updates
func SendNotificationPushAboutUpdate(db *gorm.DB) {
	certPassword := os.Getenv("CERT_PASSWORD")
	cert, err := certificate.FromP12File("./certificates/Certificates.p12", certPassword)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Push Certificate Error")
		return
	}

	// Get all registered devices
	var deviceRegs []DeviceRegistration
	if err := db.Find(&deviceRegs).Error; err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get registered devices")
		return
	}

	if len(deviceRegs) == 0 {
		log.Info().Msg("No registered devices found for push notifications")
		return
	}

	// If you want to test push notifications for builds running directly from XCode (Development), use
	// client := apns2.NewClient(cert).Development()
	// For apps published to the app store or installed as an ad-hoc distribution use Production()
	client := apns2.NewClient(cert).Production()

	// Send notification to all registered devices
	for _, deviceReg := range deviceRegs {
		notification := &apns2.Notification{}
		notification.DeviceToken = deviceReg.PushToken
		notification.Topic = "pass.com.finom.bank2wallet"
		notification.Payload = []byte(`{"aps":{"alert":"Your cashback balance was updated!"}}`)

		res, err := client.Push(notification)
		if err != nil {
			log.Error().
				Err(err).
				Str("DeviceToken", deviceReg.PushToken).
				Msg("Error sending push notification")
			continue
		}

		log.Debug().
			Interface("Result", res).
			Str("DeviceToken", deviceReg.PushToken).
			Msg("Notification sent successfully")
	}
}
