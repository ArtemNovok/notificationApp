package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/ArtemNovok/Sender/data"
)

type Email struct {
	Id          int64     `json:"id,omitempty"`
	Sender      string    `json:"sender"`
	Password    string    `json:"password"`
	Subject     string    `json:"subject"`
	Recipient   string    `json:"recipient"`
	ExpDate     time.Time `json:"expDate"`
	FullySended bool      `json:"fullysended"`
}

func (app *Config) HandlePostExp(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(16 << 20)
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
	temp, _, err := r.FormFile("template")
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
	tp, err := ParseTemp(&temp)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	id, err := data.InsertTosend(sender, password, subject, date)
	if err != nil {
		log.Println("2")
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	tp.Id = id
	err = data.InsertDocumet(tp)
	if err != nil {
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
