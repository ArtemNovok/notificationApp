package main

import (
	"os"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

var (
	accountSid = os.Getenv("TWILIO_ACCOUNT_SID")
	authToken  = os.Getenv("TWILIO_ACCOUNT_TOKEN")
)

func (app *Config) SendSMS() error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})
	params := &twilioApi.CreateMessageParams{}
	params.SetTo("+17473319360")
	params.SetFrom("+18449381236")
	params.SetBody("Test message")
	_, err := client.Api.CreateMessage(params)
	if err != nil {
		return err
	}
	return nil
}
