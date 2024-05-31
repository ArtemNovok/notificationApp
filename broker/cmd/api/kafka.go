package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ArtemNovok/Sender/data"
	"github.com/segmentio/kafka-go"
)

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
