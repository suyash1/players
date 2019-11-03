package utils

import (
	"fmt"
	"net/http"
	"time"
)

var HttpClient *http.Client

const (
	MaxIdleConnections int = 100
	RequestTimeOut     int = 30
)

func init() {
	HttpClient = createHttpClient()
}

func createHttpClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
		},
		Timeout: time.Duration(RequestTimeOut) * time.Second,
	}
	fmt.Println(client)
	return client
}
