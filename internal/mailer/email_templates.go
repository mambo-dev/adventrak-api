package mailer

import (
	"fmt"
)

func MakeEmailTemplate(title, body, link string) string {

	buttonHTML := ""
	if link != "" {
		buttonHTML = fmt.Sprintf(`
			<div style="text-align: center; margin-top: 20px;">
				<a href="%s" style="
					display: inline-block;
					padding: 12px 24px;
					background-color: #4CAF50;
					color: white;
					text-decoration: none;
					border-radius: 6px;
					font-weight: bold;
				">Click Here</a>
			</div>
		`, link)
	}

	emailHTML := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>%s</title>
		</head>
		<body style="font-family: sans-serif; padding: 24px; background-color: #f7f7f7; color: #333;">
			<div style="max-width: 600px; margin: auto; background-color: #fff; padding: 24px; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.05);">
				<h2 style="text-align: center; color: #4CAF50;">%s</h2>
				<p style="font-size: 16px; line-height: 1.5;">%s</p>
				%s
			</div>
		</body>
		</html>
	`, title, title, body, buttonHTML)

	return emailHTML
}
