package data

import (
	"database/sql"
	"log"
	"time"
)

type Email struct {
	Id        int64     `json:"id,omitempty"`
	Sender    string    `json:"sender"`
	Password  string    `json:"password"`
	Subject   string    `json:"subject"`
	Message   Message   `json:"message"`
	Recipient string    `json:"recipient"`
	ExpDate   time.Time `json:"expDate"`
}

type Message struct {
	SenderName string `json:"sendername"`
	Text       string `json:"text"`
}

type Emails struct {
	Emails []Email `json:"data"`
}

var DB *sql.DB

func NewDB(db *sql.DB) {
	DB = db
	err := CreateTestTable()
	if err != nil {
		log.Fatal(err)
	}

}

func CreateTestTable() error {
	query := `create table if not exists tosend (
		id serial primary key,
		sender varchar(500),
		sendername varchar(500),
		password varchar(500),
		recipient  varchar(500), 
		subject varchar(777),
		text text,
		expdate timestamp
	)`
	stmt, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	log.Println("Test table was created!")
	return nil
}

func InsertData(sender, sendername, password, recipient, subject, text string, expdate time.Time) error {
	query := `insert into tosend(sender, sendername ,password, recipient, subject, text, expdate) values($1, $2, $3, $4, $5, $6, $7)`
	stmt, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(sender, sendername, password, recipient, subject, text, expdate)
	if err != nil {
		return err
	}
	return nil
}

func CheckExpData() ([]Email, error) {
	query := `select * from tosend where expdate < $1`
	rows, err := DB.Query(query, time.Now().UTC())
	if err != nil {
		return []Email{}, err
	}
	defer rows.Close()
	var resp []Email
	for rows.Next() {
		var data Email
		err := rows.Scan(&data.Id, &data.Sender, &data.Message.SenderName, &data.Password, &data.Recipient, &data.Subject, &data.Message.Text, &data.ExpDate)
		if err != nil {
			log.Println(err)
			return []Email{}, err
		}
		resp = append(resp, data)
	}
	return resp, nil
}

func DeleteEmail(id int64) error {
	query := `delete from tosend where id = $1`
	stmt, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}
