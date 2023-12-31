package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

// checkResponseCode is a simple utility to check the response code
// of the response
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
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
		user:     get("POSTGRES_USER"),
		database: get("POSTGRES_DB"),
		host:     get("POSTGRES_HOST"),
		password: get("POSTGRES_PASSWORD"),
		port:     get("POSTGRES_PORT"),
		sslMode:  os.Getenv("POSTGRES_SSLMODE"), // optional, so we don't add it to missing
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

func setupForTest(t *testing.T) *Server {
	var server *Server = createNewServer(os.Getenv("PORT"))
	postgreSQLConfig, err := pgConfigFromEnv()
	if err != nil {
		t.Fatalf("Couldn't parse environment variables")
	}
	var databaseURL string = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", postgreSQLConfig.user, postgreSQLConfig.password, postgreSQLConfig.host, postgreSQLConfig.port, postgreSQLConfig.database, postgreSQLConfig.sslMode)

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Unable to connect to the database")
	}
	server.MountHandlers(pool)

	return server
}

func TestCreateQuote(t *testing.T) {
	server := setupForTest(t)

	// Create a New Request
	req, _ := http.NewRequest("GET", "/", nil)

	// Execute Request
	response := executeRequest(req, server)

	// Check response Code
	checkResponseCode(t, http.StatusOK, response.Code)

	require.Equal(t, "healthchecked", response.Body.String())
}
