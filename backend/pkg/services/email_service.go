package services

import (
	"fmt"
	"log"
	"net/smtp"
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/models"
)

type EmailService struct {
	smtpHost    string
	smtpPort    string
	senderEmail string
	senderPass  string
}

func NewEmailService() *EmailService {
	log.Println("Email username:", config.EmailUsername)
	return &EmailService{
		smtpHost:    "smtp.cs.hmc.edu",
		smtpPort:    "587",
		senderEmail: config.EmailUsername + "@cs.hmc.edu",
		senderPass:  config.EmailPassword,
	}
}

func (s *EmailService) SendBumpNotification(user models.UserRaw, roomID string, dormName string) error {
	auth := smtp.PlainAuth("", s.senderEmail, s.senderPass, s.smtpHost)

	to := []string{user.Email}
	// for testing, set to tlam@g.hmc.edu
	to = []string{"tlam@g.hmc.edu"}

	subject := fmt.Sprintf("(no-reply) Digital Draw Notification - Bumped from %s, %s", dormName, roomID)
	body := fmt.Sprintf(
		"Dear %s %s,\n\n"+
			"This email is to notify you that you have been bumped from room %s in %s Dorm.\n"+
			"Please log in to the room draw system to view more details.\n\n"+
			"Best regards,\nDigiDraw System",
		user.FirstName, user.LastName, roomID, dormName,
	)

	message := fmt.Sprintf("Subject: %s\r\n"+
		"From: %s\r\n"+
		"To: %s\r\n"+
		"\r\n"+
		"%s", subject, s.senderEmail, to[0], body)

	err := smtp.SendMail(
		s.smtpHost+":"+s.smtpPort,
		auth,
		s.senderEmail,
		to,
		[]byte(message),
	)

	if err != nil {
		log.Printf("Failed to send bump notification email: %v", err)
		return err
	}

	return nil
}
