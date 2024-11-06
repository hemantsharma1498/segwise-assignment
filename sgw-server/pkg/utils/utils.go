package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	saltSize int    = 16
	sTime    uint32 = 6
	memory   uint32 = 32
	keyLen   uint32 = 32
)

func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeBase64(encodedData string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedData)
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func CreateHash(password string, salt []byte) []byte {
	hash := argon2.Key([]byte(password), salt, sTime, memory, 8, keyLen)
	return hash
}

func DecodeReqBody(r *http.Request, d any) error {
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		return err
	}
	return nil
}

func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func WriteResponse(w http.ResponseWriter, err error, msg any, httpStatus int) error {
	if err != nil {
		log.Println(err)
	}
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)
	return json.NewEncoder(w).Encode(msg)
}

func IsoDateToTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339, date)
}
