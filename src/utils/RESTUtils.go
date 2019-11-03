package utils

import (
	"net/http"
	"runtime/debug"
)

func GET(url string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		debug.PrintStack()
	}
	response, err := HttpClient.Do(req)
	if err != nil {
		debug.PrintStack()
	}
	return response
}
