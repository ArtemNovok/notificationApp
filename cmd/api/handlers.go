package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

type GetRequest struct {
	From    string   `json:"from"`
	Subject string   `json:"subject"`
	Message Message  `json:"message"`
	To      []string `json:"to"`
}

type Message struct {
	SenderName string `json:"sendername"`
	Text       string `json:"text"`
}

var password = os.Getenv("APP_PASSWORD")

//go:embed templates/*
var templatesFS embed.FS

// This handler handle get request by decode req body and calling sendEmail func
func (app *Config) HandleGetRequest(w http.ResponseWriter, r *http.Request) {
	//Decode req to get required data
	var req GetRequest
	err := app.readJSON(w, r, &req)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	// close req body after performing logic
	defer r.Body.Close()
	log.Println(req.Message)
	// calling sendEmail func and handle error if it occurs
	err = SendEmail(req, password)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	// If call is successful response with success
	jsResp := JSONResponse{
		Error:   false,
		Message: "Emails successfully sended!!",
	}
	app.writeJSON(w, http.StatusAccepted, jsResp)

}

// THis func send emails with given data IMPORTANT password must match from attr
func SendEmail(req GetRequest, passwrod string) error {
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
