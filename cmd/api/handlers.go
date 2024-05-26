package main

import (
	"encoding/json"
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Handled!"))
}

func (app *Config) HandleGetReq(w http.ResponseWriter, r *http.Request) {
	resp, err := data.CheckExpData()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	var js data.Response
	js.Data = resp
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(js)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to encode response"))
		return
	}
}

func (app *Config) HandlePostExp(w http.ResponseWriter, r *http.Request) {
	month := r.FormValue("month")
	day := r.FormValue("day")
	hour := r.FormValue("hour")
	minute := r.FormValue("minute")
	intMonth, intDay, intHour, intMinute, err := ValidateConvertData(month, day, hour, minute)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to load location"))
	}
	date, err := CreateDate(intMonth, intDay, intHour, intMinute, &app.loc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	log.Println(date)
	if time.Now().After(date) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("This date in the Past !!!"))
		return
	}
	err = data.InsertWithExpDate(date)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to insert data into table"))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Successfully handled!"))
}
