package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ArtemNovok/Sender/data"
)

type Email struct {
	Id        int64     `json:"id,omitempty"`
	Sender    string    `json:"sender"`
	Password  string    `json:"password"`
	Subject   string    `json:"subject"`
	Message   Message   `json:"message"`
	Recipient string    `json:"recipient"`
	ExpDate   time.Time `json:"expDate"`
}

type Message struct {
	SenderName string `json:"sendername"`
	Text       string `json:"text"`
}

var password = os.Getenv("APP_PASSWORD")

// This handler handle get request by decode req body and calling sendEmail func
func (app *Config) HandlePostRequestEmails(w http.ResponseWriter, r *http.Request) {
	//Decode req to get required data
	var req Email
	err := app.readJSON(w, r, &req)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	// close req body after performing logic
	defer r.Body.Close()
	log.Println(req.Message)
	// calling sendEmail func and handle error if it occurs
	err = app.SendEmail(req, password)
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

func (app *Config) HandlePostExp(w http.ResponseWriter, r *http.Request) {
	month := r.FormValue("month")
	day := r.FormValue("day")
	hour := r.FormValue("hour")
	minute := r.FormValue("minute")
	sender := r.FormValue("sender")
	password := r.FormValue("password")
	senderName := r.FormValue("sendername")
	subject := r.FormValue("subject")
	recipient := r.FormValue("recipient")
	text := r.FormValue("text")
	if sender == "" || password == "" || senderName == "" || subject == "" || recipient == "" || text == "" {
		app.errorJSON(w, errors.New("empty fields"))
		return
	}
	intMonth, intDay, intHour, intMinute, err := ValidateConvertData(month, day, hour, minute)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	date, err := CreateDate(intMonth, intDay, intHour, intMinute, &app.loc)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Println(date)
	if time.Now().After(date) {
		app.errorJSON(w, errors.New("this date in the past"), http.StatusBadRequest)
		return
	}
	err = data.InsertData(sender, senderName, password, recipient, subject, text, date)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, http.StatusAccepted, "Successfully handled!")
}
