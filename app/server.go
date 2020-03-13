package main

import (
	"fmt"
	"karen/handlers"
	"karen/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func logging(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path: %s, method: %s", r.URL.Path, r.Method)
		f.ServeHTTP(w, r)
	})
}

func main() {
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s)/?interpolateParams=true",
		os.Getenv("KAREN_DB_USERNAME"),
		os.Getenv("KAREN_DB_PASSWORD"),
		os.Getenv("KAREN_DB_LOCATION"))
	db, err := models.NewDB(connectionString)
	if err != nil {
		log.Fatal(err)
		return
	}

	env := &handlers.Env{db}

	r := mux.NewRouter()
	r.HandleFunc("/karen/v1/users/auth", env.PostAuthHandler).Methods("POST")
	r.HandleFunc("/karen/v1/users/self", env.PostUserHandler).Methods("POST")
	r.HandleFunc("/karen/v1/users/self", env.GetUserHandler).Methods("GET")
	r.HandleFunc("/karen/v1/users/{user-id}", env.GetUserHandler).Methods("GET")
	r.HandleFunc("/karen/v1/users/self", env.PatchUserHandler).Methods("PATCH")
	r.Use(logging)

	httpSrv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      r,
	}

	log.Fatal(httpSrv.ListenAndServe())
}
