package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {

	env := os.Getenv("CHI_PGX_ENV")
	if "" == env {
		err := godotenv.Load("./.env.development.local", "./.env.development.database")
		if err != nil {
			log.Fatal("hire me! 😮", err)
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), r))
}
