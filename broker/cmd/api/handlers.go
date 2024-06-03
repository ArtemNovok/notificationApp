package main

import (
	"embed"
	"html/template"
	"net/http"
	"os"
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

type ErrorResponse struct {
	Month    string `json:"month"`
	Day      string `json:"day"`
	Hour     string `json:"hour"`
	Minute   string `json:"minute"`
	Sender   string `json:"sender"`
	Password string `json:"password"`
	Subject  string `json:"subject"`
	Error    bool   `json:"error"`
	Message  string `json:"message"`
}

var Mysecret = os.Getenv("SECRET")

func newErrorRes(month, day, hour, minute, sender, password, subject string) ErrorResponse {
	return ErrorResponse{Month: month, Day: day, Hour: hour, Minute: minute, Sender: sender, Password: password, Subject: subject}
}

//go:embed templates/*
var templatesFS embed.FS

func (app *Config) HandlePostExp(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(16 << 20)
	month := r.FormValue("month")
	day := r.FormValue("day")
	hour := r.FormValue("hour")
	minute := r.FormValue("minute")
	sender := r.FormValue("sender")
	password := r.FormValue("password")
	subject := r.FormValue("subject")
	file, _, err := r.FormFile("recipient")
	temp, _, err := r.FormFile("template")
	errResp := newErrorRes(month, day, hour, minute, sender, password, subject)
	templ := template.Must(template.ParseFS(templatesFS, "templates/index.html.gohtml"))
	isValid := ValidEmail(sender)
	if !isValid {
		errResp.Error = true
		errResp.Message = "Invalid email address"
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	if err != nil {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	defer file.Close()
	defer temp.Close()
	records, err := app.ReadContactsFile(&file)
	if err != nil {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	var recipients []string
	for _, record := range records {
		recipients = append(recipients, record[0])
	}
	if sender == "" || password == "" || subject == "" {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	ecrPassword, err := Encrypt(password, Mysecret)
	if err != nil {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}

	intMonth, intDay, intHour, intMinute, err := ValidateConvertData(month, day, hour, minute)
	if err != nil {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	date, err := CreateDate(intMonth, intDay, intHour, intMinute, &app.loc)
	if err != nil {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	if time.Now().After(date) {
		errResp.Error = true
		errResp.Message = "This date in the past"
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	tp, err := ParseTemp(&temp)
	if err != nil {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	id, err := data.InsertTosend(sender, ecrPassword, subject, date)
	if err != nil {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	tp.Id = id
	err = data.InsertDocument(tp)
	if err != nil {
		errResp.Error = true
		errResp.Message = err.Error()
		templ.ExecuteTemplate(w, "index", errResp)
		return
	}
	for _, recipient := range recipients {
		err := data.InsertRecipients(id, recipient)
		if err != nil {
			errResp.Error = true
			errResp.Message = err.Error()
			templ.ExecuteTemplate(w, "index", errResp)
			return
		}
	}
	templ.ExecuteTemplate(w, "success", nil)
}

func (app *Config) HandleMainPage(w http.ResponseWriter, r *http.Request) {
	template := template.Must(template.ParseFS(templatesFS, "templates/index.html.gohtml"))
	template.ExecuteTemplate(w, "index", nil)
}
