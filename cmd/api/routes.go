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
	// mux.Get("/sms", app.HandleGetRequestSMS)
	return mux
}
