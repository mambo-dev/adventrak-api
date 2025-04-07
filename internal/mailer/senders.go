package mailer

type EmailSender struct {
	Name  string
	Email string
}

type EmailDetails struct {
	FromEmail   string
	FromName    string
	ToEmail     string
	ToName      string
	Subject     string
	HtmlContent string
}

var SystemEmails = map[string]EmailSender{
	"system": {
		Name:  "adventrak",
		Email: "mambo.michael.22@gmail.com",
	},
}
