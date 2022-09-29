package util

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeJson(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	wantP := person{Name: "raymond", Age: 12}

	var p person
	r := strings.NewReader(`{"name":"raymond", "age":12}`)

	err := DecodeJson(r, &p)
	require.NoError(t, err)

	require.Equal(t, wantP, p)
}

func TestEncodeJson(t *testing.T) {
	haveErrStr := "invalid password"
	haveErr := errors.New(haveErrStr)

	wantStr := `{"status":false,"message":"invalid password"}`

	var b bytes.Buffer
	err := EncodeJson(&b, haveErr)
	require.NoError(t, err)

	require.Equal(t, wantStr, strings.Trim(b.String(), "\n"))
}

func TestEncodeJsonStatus(t *testing.T) {

	type person struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	}

	wantP := person{Name: "chukwudi", Country: "nigeria"}

	haveMessageStr := "successful"
	haveStatusCode := http.StatusOK

	wantObject := `{"status":true,"message":"successful","data":{"name":"chukwudi","country":"nigeria"}}`

	var b bytes.Buffer
	err := EncodeJsonStatus(&b, haveMessageStr, haveStatusCode, wantP)

	require.NoError(t, err)
	require.Equal(t, wantObject, strings.Trim(b.String(), "\n"))

}
