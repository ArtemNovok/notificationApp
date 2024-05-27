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
	r.ParseMultipartForm(10 << 20)
	month := r.FormValue("month")
	day := r.FormValue("day")
	hour := r.FormValue("hour")
	minute := r.FormValue("minute")
	sender := r.FormValue("sender")
	password := r.FormValue("password")
	senderName := r.FormValue("sendername")
	subject := r.FormValue("subject")
	text := r.FormValue("text")
	file, _, err := r.FormFile("recipient")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer file.Close()
	records, err := app.ReadContactsFile(&file)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	var recipients []string
	for _, record := range records {
		recipients = append(recipients, record[0])
	}
	if sender == "" || password == "" || senderName == "" || subject == "" || text == "" {
		app.errorJSON(w, errors.New("empty fields"))
		return
	}
	intMonth, intDay, intHour, intMinute, err := ValidateConvertData(month, day, hour, minute)
	if err != nil {
		log.Println("5")
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	date, err := CreateDate(intMonth, intDay, intHour, intMinute, &app.loc)
	if err != nil {
		log.Println("4")
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Println(date)
	if time.Now().After(date) {
		log.Println("3")
		app.errorJSON(w, errors.New("this date in the past"), http.StatusBadRequest)
		return
	}
	id, err := data.InsertTosend(sender, senderName, password, subject, text, date)
	if err != nil {
		log.Println("2")
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	for _, recipient := range recipients {
		err := data.InsertRecipients(id, recipient)
		if err != nil {
			log.Println("1")
			app.errorJSON(w, err, http.StatusInternalServerError)
			return
		}
	}
	app.writeJSON(w, http.StatusAccepted, "Successfully handled!")
}
