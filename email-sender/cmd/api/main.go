package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Email struct {
	Id        int64    `json:"id,omitempty"`
	Sender    string   `json:"sender"`
	Password  string   `json:"password"`
	Subject   string   `json:"subject"`
	Recipient []string `json:"recipient"`
	// ExpDate     time.Time `json:"expDate"`
	// FullySended bool      `json:"fullysended"`
	Template string `json:"template"`
}

type SendedEmail struct {
	Id     int64  `json:"id"`
	Sended bool   `json:"sended"`
	Error  string `json:"error"`
}
type Config struct {
}

func main() {
	log.Println("Giving time to kafka  (10 seconds)")
	time.Sleep(10 * time.Second)
	log.Println("Starting emails sender service")
	app := Config{}
	// Setting up kafka
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"broker:9093"},
		Topic:     "Emails",
		Partition: 0,
		MaxBytes:  20e6, // 10MB
		GroupID:   "EmailConsumers",
	})
	w := &kafka.Writer{
		Addr:     kafka.TCP("broker:9093"),
		Topic:    "SendedEmails",
		Balancer: &kafka.LeastBytes{},
	}
	// Closing connections
	defer r.Close()
	defer w.Close()
	for {
		// Read message from kafka
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("failed to read message from que due to error: %s\n", err.Error())
			break
		}
		var email Email
		//Decoding messages
		err = json.Unmarshal(m.Value, &email)
		if err != nil {
			log.Printf("failed to unmarshal data due to error: %s", err.Error())
			break
		}
		log.Printf("Sending emails from %s \n", email.Sender)
		// Sending emails
		err = app.SendEmailViaDB(&email)
		if err != nil {
			tempErr := err
			log.Printf("Failed to send emails: %s\n", err.Error())
			err = SendBadNotificationToSender(&email, err)
			if err != nil {
				log.Println("failed to send notification about fail")
			}
			err = Write(&email, w, false, tempErr.Error())
			if err != nil {
				log.Println("failed to send failed message")
			}
		} else {
			err = SendNotificationToSender(&email)
			if err != nil {
				log.Println("failed to send notification to sender")
			}
			err = Write(&email, w, true, "")
			if err != nil {
				log.Println("failed to put sended message into que with id: ", email.Id)
			} else {
				log.Println("message putted")
			}
		}
	}
}

func Write(email *Email, w *kafka.Writer, sended bool, errStr string) error {
	var sendedEmail SendedEmail
	sendedEmail.Id = email.Id
	sendedEmail.Sended = sended
	sendedEmail.Error = errStr
	bytes, err := json.Marshal(sendedEmail)
	if err != nil {
		return err
	}
	err = w.WriteMessages(context.Background(), kafka.Message{
		Value: bytes,
	})
	if err != nil {
		return err
	}
	return nil
}
