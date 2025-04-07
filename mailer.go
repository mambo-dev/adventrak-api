package main

import (
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailDetails struct {
	FromEmail   string
	FromName    string
	ToEmail     string
	ToName      string
	Subject     string
	HtmlContent string
}

func (cfg apiConfig) sendEmail(emailDetails EmailDetails) error {
	from := mail.NewEmail(emailDetails.FromName, emailDetails.FromEmail)
	to := mail.NewEmail(emailDetails.ToName, emailDetails.ToEmail)
	htmlContent := emailDetails.HtmlContent
	message := mail.NewSingleEmail(from, emailDetails.Subject, to, "", htmlContent)
	client := sendgrid.NewSendClient(cfg.sendGridApiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	log.Printf("Email sent succesfully received: %v code, with %v and headers %v", response.StatusCode, response.Body, response.Headers)
	return nil

}
