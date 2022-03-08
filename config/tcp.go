package config

import (
	"errors"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
)

var tcp = specification{
	protocol: "tcp",
	options: []option{
		{
			key:    "serverAddr",
			prompt: "Enter the hostname or IP address of the server.\n> ",
			process: func(resp string) (string, error) {
				resp = strings.TrimSpace(resp)
				if !(govalidator.IsDNSName(resp) || govalidator.IsIP(resp)) {
					return "", errors.New("must be a valid hostname or IP address")
				}
				return resp, nil
			},
		},
		{
			key:    "serverPort",
			prompt: "Enter the port that the server will listen on.\n> ",
			process: func(resp string) (string, error) {
				resp = strings.TrimSpace(resp)
				port, err := strconv.Atoi(resp)
				if err != nil {
					return "", err
				}
				if port < 1 || port > 65535 {
					return "", errors.New("must be a valid port number in the range 1-65535")
				}
				return resp, nil
			},
		},
	},
}
