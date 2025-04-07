package mailer

import (
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendEmail(emailDetails EmailDetails, apiKey string) error {
	from := mail.NewEmail(emailDetails.FromName, emailDetails.FromEmail)
	to := mail.NewEmail(emailDetails.ToName, emailDetails.ToEmail)
	htmlContent := emailDetails.HtmlContent
	message := mail.NewSingleEmail(from, emailDetails.Subject, to, "", htmlContent)
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	log.Printf("Email sent succesfully received: %v code, with %v ", response.StatusCode, response.Body)
	return nil

}
