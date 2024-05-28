package data

import (
	"database/sql"
	"log"
	"time"
)

type Email struct {
	Id          int64     `json:"id,omitempty"`
	Sender      string    `json:"sender"`
	Password    string    `json:"password"`
	Subject     string    `json:"subject"`
	Message     Message   `json:"message"`
	Recipient   []string  `json:"recipient"`
	ExpDate     time.Time `json:"expDate"`
	FullySended bool      `json:"fullysended"`
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
	err := CreateTosendTable()
	if err != nil {
		log.Fatal(err)
	}
	err = CreateRecipientTable()
	if err != nil {
		log.Fatal(err)
	}
}
func CreateRecipientTable() error {
	query := `create table if not exists recipients (
			id serial primary key,
			transid int,
			recipient varchar(500),
			constraint fk_tran 
			foreign key (transid) references tosend(id)
		)	`
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

func CreateTosendTable() error {
	query := `create table if not exists tosend (
		id serial primary key,
		sender varchar(500),
		sendername varchar(500),
		password varchar(500),
		subject varchar(777),
		text text,
		expdate timestamp,
		fullysended boolean)`
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

func InsertTosend(sender, sendername, password, subject, text string, expdate time.Time) (int64, error) {
	query := `insert into tosend(sender, sendername ,password,subject, text, expdate, fullysended) values($1, $2, $3, $4, $5, $6, $7) returning id`
	stmt, err := DB.Prepare(query)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(sender, sendername, password, subject, text, expdate, false)
	if err != nil {
		return -1, err
	}
	newQuery := `select id from tosend where expdate = $1 and sender = $2 and subject = $3`
	row := DB.QueryRow(newQuery, expdate, sender, subject)
	var id int64
	err = row.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}
func InsertRecipients(id int64, recipient string) error {
	query := `insert into recipients (transid, recipient) values ($1, $2)`
	stmt, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id, recipient)
	if err != nil {
		return err
	}
	return nil
}

func ChangeStatusTosended(id int64) error {
	query := `select id from recipients where transid = $1`
	res, err := DB.Query(query, id)
	if err != nil {
		return err
	}
	defer res.Close()
	var ids []int64
	for res.Next() {
		var recId int64
		err := res.Scan(&recId)
		if err != nil {
			return err
		}
		ids = append(ids, recId)
	}
	if len(ids) > 0 {
		return nil
	}
	secondQuery := `update tosend set fullysended = $1 where id = $2`
	stmt, err := DB.Prepare(secondQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(true, id)
	if err != nil {
		return err
	}
	return nil
}

func CheckExpData() ([]Email, error) {
	query := `select * from tosend where expdate < $1 and fullysended = $2 `
	rows, err := DB.Query(query, time.Now().UTC(), false)
	if err != nil {
		return []Email{}, err
	}
	defer rows.Close()
	var resp []Email
	for rows.Next() {
		var data Email
		err := rows.Scan(&data.Id, &data.Sender, &data.Message.SenderName, &data.Password, &data.Subject, &data.Message.Text, &data.ExpDate, &data.FullySended)
		if err != nil {
			log.Println(err)
			return []Email{}, err
		}
		resp = append(resp, data)
	}
	return resp, nil
}
func FindRecipients(id int64) ([]string, error) {
	query := `select recipient from recipients where transid = $1`
	res, err := DB.Query(query, id)
	if err != nil {
		return []string{}, err
	}
	defer res.Close()
	var recipients []string
	for res.Next() {
		var recipient string
		err = res.Scan(&recipient)
		if err != nil {
			return []string{}, err
		}
		recipients = append(recipients, recipient)
	}
	return recipients, nil
}
func DeleteRecipients(id int64) error {
	query := `delete from recipients where transid = $1`
	_, err := DB.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
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
