package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
)

var Mysecret = os.Getenv("SECRET")

//go:embed templates/*
var tempalateFS embed.FS

type Fail struct {
	Subject string `json:"subject"`
	Error   string `json:"error"`
}

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

func SendNotificationToSender(email *Email) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	// parsing template
	temp := template.Must(template.ParseFS(tempalateFS, "templates/success.html.gohtml"))
	var body bytes.Buffer
	//setting up headers
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	_, err := body.Write([]byte(fmt.Sprintf("Subject: Your emails sended! \n%s\n\n", mimeHeaders)))
	if err != nil {
		return err
	}
	// Executing template with given data
	err = temp.Execute(&body, email)
	if err != nil {
		return err
	}
	password := os.Getenv("APP_PASSWORD")
	sender := os.Getenv("APP_EMAIL")
	// Authenticate
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", sender, password, smtpHost)
	//Send email to given addresses
	err = smtp.SendMail(addr, auth, sender, []string{email.Sender}, body.Bytes())
	if err != nil {
		return err
	}
	log.Printf("Sended mail about success to sender")
	return nil
}

func SendBadNotificationToSender(email *Email, SendError error) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	// parsing template
	temp := template.Must(template.ParseFS(tempalateFS, "templates/fail.html.gohtml"))
	var body bytes.Buffer
	//setting up headers
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	_, err := body.Write([]byte(fmt.Sprintf("Subject: Your emails weren't sended! \n%s\n\n", mimeHeaders)))
	if err != nil {
		return err
	}
	fail := Fail{
		Subject: email.Subject,
		Error:   SendError.Error(),
	}
	// Executing template with given data
	err = temp.Execute(&body, fail)
	if err != nil {
		return err
	}
	password := os.Getenv("APP_PASSWORD")
	sender := os.Getenv("APP_EMAIL")
	// Authenticate
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", sender, password, smtpHost)
	//Send email to given addresses
	err = smtp.SendMail(addr, auth, sender, []string{email.Sender}, body.Bytes())
	if err != nil {
		return err
	}
	log.Printf("Sended mail about fail to sender")
	return nil
}
