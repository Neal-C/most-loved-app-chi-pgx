package main


import (
    "net/http"
    "net/http/httptest"
    _ "os"
    "testing"

    _ "github.com/stretchr/testify/require"
)

//TODO : make stuff into functions so that it's easier to test

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