package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
)

type Server struct {
	Router *chi.Mux
	Addr   string
}

func createNewServer(port string) *Server {
	server := &Server{
		Addr:   "0.0.0.0:" + port,
		Router: chi.NewRouter(),
	}

	return server
}

func (server *Server) MountHandlers(connectionPool *pgxpool.Pool) {
	server.Router.Use(cors.Handler(cors.Options{
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
	server.Router.Use(httprate.LimitByIP(100, 1*time.Minute))
	server.Router.Use(middleware.StripSlashes)
	server.Router.Use(middleware.RedirectSlashes)
	server.Router.Use(middleware.Logger)
	server.Router.Use(middleware.Recoverer)
	server.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthchecked"))
	})
	server.Router.Post("/quote", handlers.CreateQuote(connectionPool))
	server.Router.Get("/quote", handlers.ReadQuote(connectionPool))
	server.Router.Patch("/quote", handlers.UpdateQuote(connectionPool))
	server.Router.Delete("/quote", handlers.DeleteQuote(connectionPool))
}

func main() {

	env := os.Getenv("CHI_PGX_ENV")
	if env == "" {
		log.Println("missing .env.development.local file in the current directory")
		log.Fatal("hire me! 😮")
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect to database:", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Pinging database failed", err)
	}

	server := createNewServer(os.Getenv("PORT"))
	server.MountHandlers(pool)

	serverHTTP := &http.Server{Addr: server.Addr, Handler: server.Router}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancelFn := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancelFn()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out... forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := serverHTTP.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	err = serverHTTP.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
