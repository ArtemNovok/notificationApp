package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewMux()
	mux.Use(middleware.Recoverer)
	mux.Get("/email", app.HandleGetRequestEmails)
	mux.Post("/", app.HandlePostReq)
	mux.Post("/exp", app.HandlePostExp)
	mux.Get("/", app.HandleGetReq)
	// mux.Get("/sms", app.HandleGetRequestSMS)
	return mux
}
