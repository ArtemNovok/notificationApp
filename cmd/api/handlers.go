package main

import (
	"log"
	"net/http"
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
