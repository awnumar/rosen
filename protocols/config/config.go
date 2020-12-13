package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var protocols = []string{"https"}

// Specification is a set of config values for a protocol. One needs to be defined per supported protocol.
type specification struct {
	protocol string
	options  []option
}

// Option represents a configuration value that we will ask the user for.
type option struct {
	key     string
	prompt  string
	process func(string) (string, error)
}

// Configuration represents a set of chosen options.
type Configuration map[string]string

// JSON returns the prettified JSON representation of a configuration.
func (c Configuration) JSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "	")
}

// Configure launches the quiz that asks the user for configuration values.
// The resulting configuration is written to the working directory and the filename is returned.
func Configure() (string, error) {
	protocol := answer(fmt.Sprintf("Which protocol do you want to use?\nChoose from {%s}\n> ", strList(protocols)),
		func(resp string) (string, error) {
			if !contains(protocols, resp) {
				return "", errors.New("invalid protocol")
			}
			return resp, nil
		})

	switch protocol {
	case "https":
		return processSpec(https)
	default:
		panic("error: unknown protocol") // should never happen
	}
}

func processSpec(spec specification) (string, error) {
	config := make(Configuration)
	config["protocol"] = spec.protocol
	config["authToken"] = generateAuthToken()
	for _, q := range spec.options {
		config[q.key] = answer(q.prompt, q.process)
	}
	return writeTofile(config)
}

func answer(prompt string, verify func(string) (string, error)) (data string) {
	var err error
	for {
		data, err = verify(input(prompt))
		if err == nil {
			break
		}
		color.HiRed("error: %s", err)
	}
	return
}

func input(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n%s", prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(text)
}

func strList(list []string) string {
	return strings.Join(list, ", ")
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if item == v {
			return true
		}
	}
	return false
}
