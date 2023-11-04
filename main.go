package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/Neal-C/most-loved-app-go-pgx/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

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

	// if err := pool.Ping(context.Background()); err != nil {
	// 	log.Fatal("Pinging database failed", err)
	// }

	chiRouter := chi.NewRouter()
	chiRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{`https:\/\/*`, `http:\/\/*`},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	// Enable httprate request limiter of 100 requests per minute.
	//
	// In the code example below, rate-limiting is bound to the request IP address
	// via the LimitByIP middleware handler.
	//
	// To have a single rate-limiter for all requests, use httprate.LimitAll(..).
	//
	// Please see _example/main.go for other more, or read the library code.
	chiRouter.Use(httprate.LimitByIP(100, 1*time.Minute))
	chiRouter.Use(middleware.RedirectSlashes)
	chiRouter.Use(middleware.Logger)
	chiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthchecked"))
	})
	chiRouter.Post("/quote", handlers.CreateQuote(pool))
	chiRouter.Get("/quote", handlers.ReadQuote(pool))
	chiRouter.Patch("/quote", handlers.UpdateQuote(pool))
	chiRouter.Delete("/quote", handlers.DeleteQuote(pool))
	// log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), chiRouter))

	server := &http.Server{Addr: "0.0.0.0:" + os.Getenv("PORT"), Handler: chiRouter}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		// lint:ignore
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out... forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
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
