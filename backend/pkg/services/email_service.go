package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"roomdraw/backend/pkg/models"

	"github.com/joho/godotenv"
)

type EmailService struct {
	smtpHost    string
	smtpPort    string
	senderEmail string
	senderPass  string
}

func NewEmailService() *EmailService {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	log.Println("os.Getenv(EMAIL_USERNAME): ", os.Getenv("EMAIL_USERNAME"))
	return &EmailService{
		smtpHost:    "smtp.cs.hmc.edu",
		smtpPort:    "587",
		senderEmail: os.Getenv("EMAIL_USERNAME") + "@cs.hmc.edu",
		senderPass:  os.Getenv("EMAIL_PASSWORD"),
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
