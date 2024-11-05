package server

import (
	"net/http"
)

/*
*ConfirmSale
 */

func (s *Server) Routes() {
	s.Router.HandleFunc("/api/login", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		s.Login(w, r)
	})))
}

func withCORS(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the origin from the request
		origin := r.Header.Get("Origin")

		// List of allowed origins
		allowedOrigins := []string{
			//	"https://main.d3to1cludkqj3l.amplifyapp.com",
			"http://localhost:3000",
			"https://main.d3to1cludkqj3l.amplifyapp.com/",
		}

		// Check if the request origin is in the list of allowed origins
		allowedOrigin := ""
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				allowedOrigin = origin
				break
			}
		}

		// If the origin is allowed, set the CORS headers
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Call the original handler
		handler.ServeHTTP(w, r)

	}
}
