package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ArtemNovok/Sender/data"
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

// This handler handle get request by decode req body and calling sendEmail func
func (app *Config) HandleGetRequestEmails(w http.ResponseWriter, r *http.Request) {
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

func (app *Config) HandleGetRequestSMS(w http.ResponseWriter, r *http.Request) {
	err := app.SendSMS()
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
	}
	app.writeJSON(w, http.StatusAccepted, nil)
}

func (app *Config) HandlePostReq(w http.ResponseWriter, r *http.Request) {
	err := data.InsertNow()
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, http.StatusAccepted, "Handled")
}

func (app *Config) HandleGetReq(w http.ResponseWriter, r *http.Request) {
	resp, err := data.CheckExpData()
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	var js data.Response
	js.Data = resp
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
	}
	app.writeJSON(w, http.StatusAccepted, js)
}

func (app *Config) HandlePostExp(w http.ResponseWriter, r *http.Request) {
	month := r.FormValue("month")
	day := r.FormValue("day")
	hour := r.FormValue("hour")
	minute := r.FormValue("minute")
	intMonth, intDay, intHour, intMinute, err := ValidateConvertData(month, day, hour, minute)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
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
		app.errorJSON(w, errors.New("This date in the past"), http.StatusBadRequest)
		return
	}
	err = data.InsertWithExpDate(date)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, http.StatusAccepted, "Successfully handled")
}
