package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
)

//go:embed templates/*
var templatesFS embed.FS

// THis func send emails with given data IMPORTANT password must match from attr
func (app *Config) SendEmail(req GetRequest, passwrod string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	// parsing template
	temp, err := template.ParseFS(templatesFS, "templates/mailtemp.html.gohtml")
	if err != nil {
		return err
	}
	var body bytes.Buffer
	//setting up headers
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	_, err = body.Write([]byte(fmt.Sprintf("Subject:%s \n%s\n\n", req.Subject, mimeHeaders)))
	if err != nil {
		return err
	}
	// Executing template with given data
	err = temp.Execute(&body, req)
	if err != nil {
		return err
	}
	// Authenticate
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", req.From, passwrod, smtpHost)
	//Send email to given addresses
	err = smtp.SendMail(addr, auth, req.From, req.To, body.Bytes())
	if err != nil {
		return err
	}
	log.Printf("Sended mails from: %s", req.From)
	return nil
}
