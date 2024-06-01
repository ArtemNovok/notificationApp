package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ArtemNovok/Sender/data"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	loc time.Location
}

var Conn *kafka.Conn
var Writer *kafka.Writer

const (
	webPort     = "8000"
	postgResurl = "host=postgres user=postgres password=mysecretpassword dbname=postgres sslmode=disable timezone=GMT-7 connect_timeout=5"
	mongourl    = "mongodb://mongodb"
	kafkaHost   = "localhost:9092"
)

func main() {
	log.Println("Giving time to kafka (10 second)")
	time.Sleep(time.Second * 10)
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("location loaded")
	app := Config{
		loc: *loc,
	}
	log.Println(time.Now())
	db, err := ConnectToDB(postgResurl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	data.NewDB(db)
	con, err := ConectToMongo(mongourl)
	if err != nil {
		log.Fatal(err)
	}
	data.NewClient(con)
	w := &kafka.Writer{
		Addr:     kafka.TCP("broker:9093"),
		Topic:    "Emails",
		Balancer: &kafka.LeastBytes{},
	}
	Writer = w
	defer Writer.Close()
	r := NewReader()
	defer r.Close()
	go ReadMessages(r)
	log.Printf("Starting server on port:%s ...", webPort)
	var wg sync.WaitGroup
	var wg2 sync.WaitGroup
	go BackgroundChecker(&app, &wg)
	go MissedEmailsChecker(&app, &wg2)
	go BackgroundCleaner()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func ConnectToDB(url string) (*sql.DB, error) {
	count := 0
	for {
		con, err := sql.Open("pgx", url)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		err = con.Ping()
		if err == nil {
			log.Println("Successfully connected to database")
			return con, nil
		}
		if count > 8 {
			return nil, errors.New("failed to connect to db")
		}
		log.Println("Baking off for 2 seconds...")
		time.Sleep(time.Second * 2)
		continue
	}
}

func MissedEmailsChecker(app *Config, wg *sync.WaitGroup) {
	for {
		time.Sleep(time.Second * 120)
		missedEmails, err := data.CheckMissedMessages()
		if err != nil {
			log.Println("failed to check missed data: ", err.Error())
			continue
		}
		for _, email := range missedEmails {
			wg.Add(1)
			go func(app *Config, email data.Email) {
				err = SendTranEmail(app, &email)
				if err != nil {
					log.Println("Failed to push emails in que: ", err.Error())
				}
				defer wg.Done()
			}(app, email)
		}
		wg.Wait()
	}
}

func BackgroundChecker(app *Config, wg *sync.WaitGroup) {
	for {
		emails, err := data.CheckExpData()
		if err != nil {
			log.Println("Failed to check exp data in background:", err.Error())
			continue
		}
		for _, email := range emails {
			wg.Add(1)
			go func(app *Config, email data.Email) {
				err = SendTranEmail(app, &email)
				if err != nil {
					log.Println("Failed to push emails in que: ", err.Error())
				}
				defer wg.Done()
			}(app, email)
		}
		wg.Wait()
		time.Sleep(time.Second * 30)
	}
}
func SendTranEmail(app *Config, email *data.Email) error {
	recipents, err := data.FindRecipients(email.Id)
	if err != nil {
		return err
	}
	email.Recipient = recipents
	templ, err := data.GetDocument(email.Id)
	if err != nil {
		return err
	}
	email.Template = templ.Str
	err = PlaceEmail(*email)
	if err != nil {
		log.Printf("failed to push email into que: %s", err.Error())
		return err
	}
	err = data.ChangeQueStatus(email.Id)
	if err != nil {
		return err
	}
	return nil
}
func ConectToMongo(url string) (*mongo.Client, error) {
	count := 0
	for {
		count++
		cl, err := mongo.Connect(context.Background(), options.Client().ApplyURI(url))
		if err != nil {
			log.Println(err)
		}
		err = cl.Ping(context.Background(), &readpref.ReadPref{})
		if err == nil {
			log.Println("Successfully connected to db!")
			return cl, nil
		}
		if count > 8 {
			return nil, errors.New("failed to connect to mongo")
		}
		log.Println("Baking off for 2 seconds ...")
		time.Sleep(time.Second * 2)
	}
}

func BackgroundCleaner() {
	for {
		err := data.DeleteSendedMessages()
		if err != nil {
			log.Println("failed to clean tables: ", err.Error())
		}
		time.Sleep(time.Second * 60)
	}
}
