package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/ArtemNovok/Sender/data"
)

type JSONResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Writes json  using given status code, data, headers and returns error if something goes wrong
func (app *Config) writeJSON(w http.ResponseWriter, statusCode int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for key, val := range headers[0] {
			w.Header()[key] = val
		}
	}
	w.WriteHeader(statusCode)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// Reads json and marshal given data struct and returns error if something goes wrong
func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megaByte

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		log.Println("3")
		return err
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only one JSON value")
	}
	return nil
}

// Writes error response using given error, and status codes in JSONResponse struct
func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}

func ValidateConvertData(month, day, hour, minute string) (int, int, int, int, error) {
	if month == "" || day == "" || hour == "" || minute == "" {
		return -1, -1, -1, -1, errors.New("empty fields")
	}
	intMonth, err := strconv.Atoi(month)
	if err != nil || intMonth > 12 || intMonth < 1 {
		return -1, -1, -1, -1, errors.New("invalid month val")
	}
	intDay, err := strconv.Atoi(day)
	if err != nil || intDay < 1 || intDay > 31 {
		return -1, -1, -1, -1, errors.New("invalid day val")
	}
	intHour, err := strconv.Atoi(hour)
	if err != nil || intHour < 0 || intHour > 24 {
		return -1, -1, -1, -1, errors.New("invalid hour val")
	}
	intMinute, err := strconv.Atoi(minute)
	if err != nil || intMinute < 0 || intMinute > 59 {
		return -1, -1, -1, -1, errors.New("invalid minute val")
	}
	return intMonth, intDay, intHour, intMinute, nil
}

func CreateDate(month, day, hour, min int, loc *time.Location) (time.Time, error) {
	switch month {
	case 1:

		return time.Date(time.Now().Year(), time.January, day, hour, min, 0, 0, loc), nil
	case 2:
		if day > 28 {
			return time.Time{}, errors.New("invalid day for that month")
		}
		return time.Date(time.Now().Year(), time.February, day, hour, min, 0, 0, loc), nil
	case 3:
		return time.Date(time.Now().Year(), time.March, day, hour, min, 0, 0, loc), nil
	case 4:
		if day > 30 {
			return time.Time{}, errors.New("invalid day for that month")
		}
		return time.Date(time.Now().Year(), time.April, day, hour, min, 0, 0, loc), nil
	case 5:
		return time.Date(time.Now().Year(), time.May, day, hour, min, 0, 0, loc), nil
	case 6:
		if day > 30 {
			return time.Time{}, errors.New("invalid day for that month")
		}
		return time.Date(time.Now().Year(), time.June, day, hour, min, 0, 0, loc), nil
	case 7:
		return time.Date(time.Now().Year(), time.July, day, hour, min, 0, 0, loc), nil
	case 8:
		return time.Date(time.Now().Year(), time.August, day, hour, min, 0, 0, loc), nil
	case 9:
		if day > 30 {
			return time.Time{}, errors.New("invalid day for that month")
		}
		return time.Date(time.Now().Year(), time.September, day, hour, min, 0, 0, loc), nil
	case 10:
		return time.Date(time.Now().Year(), time.October, day, hour, min, 0, 0, loc), nil
	case 11:
		if day > 30 {
			return time.Time{}, errors.New("invalid day for that month")
		}
		return time.Date(time.Now().Year(), time.November, day, hour, min, 0, 0, loc), nil
	default:
		return time.Date(time.Now().Year(), time.December, day, hour, min, 0, 0, loc), nil
	}
}

func (app *Config) ReadContactsFile(file *multipart.File) ([][]string, error) {
	records, err := csv.NewReader(*file).ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	return records, nil
}

func ParseTemp(file *multipart.File) (data.Template, error) {
	bytes, err := io.ReadAll(*file)
	if err != nil {
		return data.Template{}, err
	}
	str := string(bytes)
	var temp data.Template
	temp.Str = str
	return temp, nil
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Encrypt(text, Mysecret string) (string, error) {
	bytes := []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}
	block, err := aes.NewCipher([]byte(Mysecret))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, bytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func Decrypt(text, MySecret string) (string, error) {
	bytes := []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}
	block, err := aes.NewCipher([]byte(MySecret))

	if err != nil {
		return "", err
	}
	cipherText := Decode(text)
	cfb := cipher.NewCFBDecrypter(block, bytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}
