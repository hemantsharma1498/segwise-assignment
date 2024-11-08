package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/hemantsharma1498/segwise-assignment/pkg/scraper"
	"golang.org/x/crypto/argon2"
	"net/http"
	"net/mail"
	"time"
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

func WriteResponse(w http.ResponseWriter, msg any, httpStatus int) error {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)
	return json.NewEncoder(w).Encode(msg)
}

func IsoDateToTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339, date)
}

func WithCORS(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowedOrigins := []string{
			"http://localhost:8080",
		}

		allowedOrigin := ""
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				allowedOrigin = origin
				break
			}
		}

		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)

	}
}

func GetUsedParams(profile scraper.Profile) []string {
	checks := map[string]func() bool{
		"Posts":      func() bool { return len(profile.Posts) > 0 },
		"Experience": func() bool { return len(profile.Experience) > 0 },
		"Education":  func() bool { return len(profile.Education) > 0 },
		"Location":   func() bool { return profile.Location != "" },
		"Name":       func() bool { return profile.Name != "" },
	}
	paramsUsed := make([]string, 0, len(checks))

	for param, check := range checks {
		if check() {
			paramsUsed = append(paramsUsed, param)
		}
	}
	return paramsUsed
}
