package main

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
)

type GetRequest struct {
	From    string   `json:"from"`
	Message string   `json:"message"`
	To      []string `json:"to"`
}

const password = ""

func (app *Config) HandleGetRequest(w http.ResponseWriter, r *http.Request) {
	var req GetRequest
	err := app.readJSON(w, r, &req)
	if err != nil {
		log.Println("1")
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	log.Println(req.Message)
	err = SendEmail(req.From, password, req.Message, req.To)
	if err != nil {
		log.Println("2")
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	jsResp := JSONResponse{
		Error:   false,
		Message: "Emails successfully sended!!",
	}
	app.writeJSON(w, http.StatusAccepted, jsResp)

}

func SendEmail(from, passwrod, message string, to []string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	byteMessage := []byte(message)
	auth := smtp.PlainAuth("", from, passwrod, smtpHost)
	err := smtp.SendMail(addr, auth, from, to, byteMessage)
	if err != nil {
		return err
	}
	log.Printf("Sended mails from: %s", from)
	return nil
}
