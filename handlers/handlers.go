package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Quote struct {
	Id         uuid.UUID `json:"id,omitempty" db:"id"`
	Book       string    `json:"book,omitempty" db:"book"`
	Quote      string    `json:"quote,omitempty" db:"quote"`
	InsertedAt time.Time `json:"insertedAt,omitempty" db:"inserted_at"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty" db:"updated_at"`
}

func (self Quote) New(book, quote string) Quote {
	var now time.Time = time.Now()
	return Quote{
		Id:         uuid.New(),
		Book:       book,
		Quote:      quote,
		InsertedAt: now,
		UpdatedAt:  now,
	}
}

type quoteCreationArgs struct {
	Book  string `json:"book,omitempty"`
	Quote string `json:"quote,omitempty"`
}

func CreateQuote(postgreSQLPool *pgxpool.Pool) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		arguments := quoteCreationArgs{}
		if err := json.NewDecoder(request.Body).Decode(&arguments); err != nil {
			WriteError(responseWriter, err, http.StatusBadRequest)
			return
		}
		fmt.Printf("arguments %#+v \n", arguments)
		var quote Quote
		quote = quote.New(arguments.Book, arguments.Quote)
		fmt.Printf("quote %#+v \n", quote)

		var sqlQuery string = "INSERT INTO quote (id, book, quote, inserted_at, updated_at) VALUES ($1, $2, $3, $4, $5)"

		_, err := postgreSQLPool.Exec(context.TODO(), sqlQuery, quote.Id, quote.Book, quote.Quote, quote.InsertedAt, quote.UpdatedAt)

		if err != nil {
			WriteError(responseWriter, err, http.StatusInternalServerError)
			return
		}
		fmt.Printf("quote again %#+v \n", quote)
		WriteJSON(responseWriter, http.StatusCreated, quote)
		return
	}
}

func ReadQuote(postgreSQLPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := postgreSQLPool.Query(context.TODO(), "SELECT * FROM quote")
		if err != nil {
			WriteError(w, err, http.StatusInternalServerError)
			return
		}
		quotes, err := pgx.CollectRows[Quote](rows, pgx.RowToStructByPos[Quote])
		if err != nil {
			WriteError(w, err, http.StatusInternalServerError)
			return
		}

		WriteJSON(w, http.StatusOK, quotes)
		return

	}
}

func WriteJSON(responseWriter http.ResponseWriter, status int, value any) error {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(status)
	return json.NewEncoder(responseWriter).Encode(value)
}

type Error struct {
	Error string `json:"error"`
} // no need for omitempty here; we'll never send an empty error.
// WriteError logs an error, then writes it as a JSON object in the form {"error": <error>}, setting the Content-Type header to application/json.
func WriteError(w http.ResponseWriter, err error, statusCode int) {
	log.Printf("%d %v: %v", statusCode, http.StatusText(statusCode), err) // log the error; http.StatusText gets "Not Found" from 404, etc.
	w.Header().Set("Content-Type", "encoding/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Error{err.Error()})
}

// ReadJSON reads a JSON object from an io.ReadCloser, closing the reader when it's done. It's primarily useful for reading JSON from *http.Request.Body.
func ReadJSON[T any](r io.ReadCloser) (T, error) {
	var v T                               // declare a variable of type T
	err := json.NewDecoder(r).Decode(&v)  // decode the JSON into v
	return v, errors.Join(err, r.Close()) // close the reader and return any errors.
}
