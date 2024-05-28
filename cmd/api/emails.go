package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"

	"github.com/ArtemNovok/Sender/data"
)

func (app *Config) SendEmailViaDB(email *data.Email) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	// parsing template
	t, err := data.GetDocument(email.Id)
	if err != nil {
		return err
	}
	temp, err := template.New("temp").Parse(t.Str)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	//setting up headers
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	_, err = body.Write([]byte(fmt.Sprintf("Subject:%s \n%s\n\n", email.Subject, mimeHeaders)))
	if err != nil {
		return err
	}
	// Executing template with given data
	err = temp.Execute(&body, nil)
	if err != nil {
		return err
	}
	// Authenticate
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", email.Sender, email.Password, smtpHost)
	//Send email to given addresses
	err = smtp.SendMail(addr, auth, email.Sender, email.Recipient, body.Bytes())
	if err != nil {
		return err
	}
	log.Printf("Sended mails from: %s", email.Sender)
	return nil
}
