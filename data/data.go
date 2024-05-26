package data

import (
	"database/sql"
	"log"
	"time"
)

type TestData struct {
	Text    string    `json"text"`
	ExpData time.Time `json:"expdata"`
}

type Response struct {
	Data []TestData `json:"data"`
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
	query := `create table if not exists test(
		text varchar(200),
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

func InsertNow() error {
	query := `insert into test(text, expdate) values($1, $2)`
	_, err := DB.Exec(query, "test", time.Now())
	if err != nil {
		return err
	}
	return nil
}

func InsertWithExpDate(date time.Time) error {
	query := `insert into test(text, expdate) values($1, $2)`
	_, err := DB.Exec(query, "test", date)
	if err != nil {
		return err
	}
	return nil
}

func CheckExpData() ([]TestData, error) {
	query := `select * from test where expdate < $1`
	rows, err := DB.Query(query, time.Now().UTC())
	if err != nil {
		return []TestData{}, err
	}
	defer rows.Close()
	var resp []TestData
	for rows.Next() {
		var data TestData
		err := rows.Scan(&data.Text, &data.ExpData)
		if err != nil {
			log.Println(err)
			return []TestData{}, err
		}
		resp = append(resp, data)
	}
	return resp, nil
}

func BackgroundChecker() {
	for {
		reps, err := CheckExpData()
		if err != nil {
			log.Println("Failed to check exp data in background")
			continue
		}
		log.Println(reps)
		time.Sleep(time.Second * 60)
	}
}
