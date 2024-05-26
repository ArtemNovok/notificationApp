package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ArtemNovok/Sender/data"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Config struct {
	loc time.Location
}

const webPort = "8000"
const postgResurl = "host=host.docker.internal port=5432 user=postgres password=mysecretpassword dbname=postgres sslmode=disable timezone=UTC connect_timeout=5"

func main() {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal(err)
	}
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
	log.Printf("Starting server on port:%s ...", webPort)
	go data.BackgroundChecker()
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
			return nil, errors.New("Failed to connect to db")
		}
		log.Println("Baking off for 2 seconds...")
		time.Sleep(time.Second * 2)
		continue
	}
}
