package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Email struct {
	Id          int64     `json:"id,omitempty"`
	Sender      string    `json:"sender"`
	Password    string    `json:"password"`
	Subject     string    `json:"subject"`
	Recipient   []string  `json:"recipient"`
	ExpDate     time.Time `json:"expDate"`
	FullySended bool      `json:"fullysended"`
	Template    string    `json:"template"`
}

type Config struct {
}

func main() {
	log.Println("Starting emails sender service")
	app := Config{}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"broker:9093"},

		Topic:     "Emails",
		Partition: 0,
		MaxBytes:  10e6, // 10MB
		GroupID:   "EmailConsumers",
	})
	defer r.Close()
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("failed to read message from que due to error: %s\n", err.Error())
			break
		}
		var email Email
		err = json.Unmarshal(m.Value, &email)
		if err != nil {
			log.Printf("failed to unmarshal data due to error: %s", err.Error())
			break
		}
		log.Printf("Sending emails from %s \n", email.Sender)
		err = app.SendEmailViaDB(&email)
		if err != nil {
			log.Printf("Failed to send emails: %s\n", err.Error())
		}
	}
}
