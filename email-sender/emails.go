package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
)

var Mysecret = os.Getenv("SECRET")

func (app *Config) SendEmailViaDB(email *Email) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	// parsing template
	temp, err := template.New("temp").Parse(email.Template)
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
	//Decrypt password
	password, err := Decrypt(email.Password, Mysecret)
	if err != nil {
		return err
	}
	// Authenticate
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", email.Sender, password, smtpHost)
	//Send email to given addresses
	err = smtp.SendMail(addr, auth, email.Sender, email.Recipient, body.Bytes())
	if err != nil {
		return err
	}
	log.Printf("Sended mails from: %s", email.Sender)
	return nil
}
