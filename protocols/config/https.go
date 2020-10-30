package config

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
)

var https = Specification{
	protocol: "https",
	options: []Option{
		{
			key:     "tlsCert",
			prompt:  "Enter the path to a TLS certificate file. This certificate will be pinned by the client.\n> ",
			process: importCert,
		},
		{
			key:     "tlsKey",
			prompt:  "Enter the path to a TLS certificate key file.\n> ",
			process: importCert,
		},
		{
			key:    "tlsMaxVersion",
			prompt: "Set the maximum TLS version that should be used, 1.2 or 1.3\n> ",
			process: func(resp string) (string, error) {
				if resp != "1.2" && resp != "1.3" {
					return "", errors.New("input must be one of 1.2 or 1.3")
				}
				return resp, nil
			},
		},
	},
}

func importCert(path string) (string, error) {
	if path == "" {
		return "", errors.New("option is required")
	}
	pemData, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(pemData), nil
}
