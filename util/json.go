package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var ErrEmptyBody = errors.New("body must not be empty")

func DecodeJson(body io.Reader, o interface{}) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError

	err := json.NewDecoder(body).Decode(o)
	if err != nil {
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return ErrEmptyBody

		default:
			return err
		}

	}

	return nil
}

func EncodeJson(w io.Writer, err error) error {
	return json.NewEncoder(w).Encode(NewServiceErrResponse(err))
}

func EncodeJsonStatus(w io.Writer, message string, statusCode int, o interface{}) error {
	return json.NewEncoder(w).Encode(NewServerResponse(message, o, statusCode))
}
