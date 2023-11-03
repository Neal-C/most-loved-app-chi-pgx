package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/Neal-C/most-loved-app-go-pgx/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	/*
		Note the ‘_’ : this is a “blank import”, which means that we’re importing the package for it’s side effects (that is, registering the driver).
	*/
	// _ "github.com/jackc/pgx/v5" // register the driver // needed if database/sql
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	env := os.Getenv("CHI_PGX_ENV")
	if "" == env {
		err := godotenv.Load("./.env.development.local")
		if err != nil {
			log.Println("missing .env.development.local file in the current directory")
			log.Fatal("hire me! 😮", err)
		}
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect to database:", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.TODO()); err != nil {
		log.Fatal("Pinging database failed", err)
	}

	chiRouter := chi.NewRouter()
	chiRouter.Use(middleware.Logger)
	chiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthchecked"))
	})
	chiRouter.Post("/quote", handlers.CreateQuote(pool))
	chiRouter.Get("/quote", handlers.ReadQuote(pool))
	chiRouter.Patch("/quote", func(w http.ResponseWriter, r *http.Request) {

	})
	chiRouter.Delete("/quote", func(w http.ResponseWriter, r *http.Request) {

	})
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), chiRouter))
}

// pgconfig is a struct that holds the configuration for connecting to a postgres database.
// each field corresponds to a component of the connection string.
// the following required environment variables are used to populate the struct:
//
//	PG_USER
//	 PG_PASSWORD
//	 PG_HOST
//	 PG_PORT
//	 PG_DATABASE
//
// additionally, the following optional environment variable is used to populate the sslmode:
//
//	PG_SSLMODE: must be one of  "", "disable", "allow", "require", "verify-ca", or "verify-full"
type pgconfig struct {
	user, database, host, password, port string // required
	sslMode                              string // optional
}

func pgConfigFromEnv() (pgconfig, error) {
	var missing []string
	// small closures like this can help reduce code duplication and make intent clearer.
	// you generally pay a small performance penalty for this, but for configuration, it's not a big deal;
	// you can spare the nanoseconds.
	// i prefer little helper functions like this to a complicated configuration framework like viper, cobra, envconfig, etc.
	get := func(key string) string {
		val := os.Getenv(key)
		if val == "" {
			missing = append(missing, key)
		}
		return val
	}
	cfg := pgconfig{
		user:     get("PG_USER"),
		database: get("PG_DATABASE"),
		host:     get("PG_HOST"),
		password: get("PG_PASSWORD"),
		port:     get("PG_PORT"),
		sslMode:  os.Getenv("PG_SSLMODE"), // optional, so we don't add it to missing
	}
	switch cfg.sslMode {
	case "", "disable", "allow", "require", "verify-ca", "verify-full":
		// valid sslmode
	default:
		return cfg, fmt.Errorf(`invalid sslmode "%s": expected one of "", "disable", "allow", "require", "verify-ca", or "verify-full"`, cfg.sslMode)
	}

	if len(missing) > 0 {
		sort.Strings(missing) // sort for consistency in error message
		return cfg, fmt.Errorf("missing required environment variables: %v", missing)
	}
	return cfg, nil
}
