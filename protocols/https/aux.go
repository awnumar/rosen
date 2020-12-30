package https

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	clientBufferSize = 4096
	serverBufferSize = 4096
)

var staticWebsiteHandler = http.FileServer(http.Dir("public"))

func getResponseText(resp *http.Response) (string, error) {
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	return string(respBytes), nil
}

// custom logger for http client

type logger struct{}

func (l logger) Error(msg string, keysAndValues ...interface{}) {
	output(msg, keysAndValues...)
}

func (l logger) Info(msg string, keysAndValues ...interface{}) {
	output(msg, keysAndValues...)
}

func (l logger) Debug(msg string, keysAndValues ...interface{}) {}

func (l logger) Warn(msg string, keysAndValues ...interface{}) {
	output(msg, keysAndValues...)
}

func output(msg string, keysAndValues ...interface{}) {
	fmt.Print(msg, " ")
	for _, kv := range keysAndValues {
		fmt.Print(kv, " ")
	}
	fmt.Println("")
}
