package main

import (
	"net/smtp"

	"github.com/jirbthagoras/hon/shared"
)

var config = shared.NewConfig()

type Mailer struct {
	Auth smtp.Auth
}

func NewMailer() *Mailer {

	// take the username
	username := config.GetString("SMTP_USERNAME")
	password := config.GetString("SMTP_PASSWORD")
	host := config.GetString("SMTP_HOST")

	// returns
	return &Mailer{Auth: smtp.PlainAuth("", username, password, host)}
}

func (m *Mailer) SendMail(data *SendMail) error {
	from := config.GetString("SMTP_FROM")
	port := config.GetString("SMTP_PORT")
	host := config.GetString("SMTP_HOST")

	msg := "From: " + from + "\n" +
		"To: " + data.To + "\n" +
		"Subject: " + data.Subject + "\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		data.Body

	err := smtp.SendMail(host+":"+port, m.Auth, from, []string{data.To}, []byte(msg))
	return err
}
