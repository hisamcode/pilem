package helper

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type Envelope map[string]any

// WriteJSON write json with http.ResponseWriter
func WriteJSON(w http.ResponseWriter, status int, data Envelope, headers http.Header) error {

	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// AnyToJSON Create From envelope to json string
// Usually this helper for testing
func AnyToJSON(data any) (string, error) {
	js, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB.
	// to prevent denial-of-service attack
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(dst)
	if err != nil {
		var (
			syntaxError           *json.SyntaxError
			unmarshallTypeError   *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
			maxBytesError         *http.MaxBytesError
		)

		switch {
		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON.
		case errors.As(err, &syntaxError):
			return errors.New("body contains badly-formed JSON")

		// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the
		// JSON value is the wrong type for the target destination.
		case errors.As(err, &unmarshallTypeError):
			if unmarshallTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshallTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshallTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format "json: unknown
		// field "<name>"".
		// Note that there's an open
		// issue at https://github.com/golang/go/issues/29035 regarding turning this
		// into a distinct error type in the future.
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		//it means the request body exceeded our
		// size limit of 1MB
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		// A json.InvalidUnmarshalError error will be returned if we pass something
		// that is not a non-nil pointer to Decode().
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a single JSON value this will
	// return an io.EOF error. So if we get anything else, we know that there is
	// additional data in the request body and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contains a single JSON value")
	}

	return nil
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := Envelope{"error": message}
	err := WriteJSON(w, status, env, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "the server encountered a problem and could not process your request"
	ErrorResponse(w, r, http.StatusInternalServerError, message)
}

func NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "the requested resource could not be found"
	ErrorResponse(w, r, http.StatusNotFound, message)
}

func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	ErrorResponse(w, r, http.StatusUnprocessableEntity, err.Error())
}

// NewSQLMock helper for stub sql
func NewSQLMock(t *testing.T, fn func(mock sqlmock.Sqlmock)) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Cant creates sqlmock database connection. Err:%v", err)
	}
	t.Cleanup(func() {
		defer db.Close()
	})

	fn(mock)
	return db, mock
}
