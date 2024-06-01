package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ArtemNovok/Sender/data"
	"github.com/segmentio/kafka-go"
)

type SendedEmail struct {
	Id     int64 `json:"id"`
	Sended bool  `json:"sended"`
}

func ConnectKafka(ctx context.Context, conType, host, topic string, partition int) (*kafka.Conn, error) {
	return kafka.DialLeader(ctx, conType, host, topic, partition)
}

func PlaceEmail(email data.Email) error {
	payload, err := json.Marshal(email)
	if err != nil {
		return err
	}
	err = Writer.WriteMessages(context.Background(), kafka.Message{
		Value: payload,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func NewReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"broker:9093"},
		Topic:     "SendedEmails",
		Partition: 0,
		MaxBytes:  20e6, // 10MB
		GroupID:   "SendedEmailsConsumer",
	})
}

func ReadMessages(reader *kafka.Reader) {
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("failed to read sended message")
			continue
		}
		var email SendedEmail
		err = json.Unmarshal(m.Value, &email)
		if err != nil {
			log.Println("failed to unmarshal message value")
			continue
		}
		err = data.DeleteSendedTans(email.Id)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Message was sended and its status is changed")
	}
}