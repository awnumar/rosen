package config

import (
	"errors"
	"strings"

	"github.com/asaskevich/govalidator"
)

var https = specification{
	protocol: "https",
	options: []option{
		{
			key:    "proxyAddr",
			prompt: "Enter the address that the client will use to connect to the proxy server.\nIt must start with https://\n> ",
			process: func(resp string) (string, error) {
				resp = strings.TrimSpace(resp)
				if !govalidator.IsURL(resp) {
					return "", errors.New("input must be an URL")
				}
				if !strings.HasPrefix(resp, "https://") {
					return "", errors.New("input must start with https://")
				}
				return resp, nil
			},
		},
		{
			key:    "hostname",
			prompt: "Enter the public hostname that your server will be accessible from.\nThis will be used for TLS certificate provisioning.\n> ",
			process: func(resp string) (string, error) {
				resp = strings.TrimSpace(resp)
				if !govalidator.IsDNSName(resp) {
					return "", errors.New("input must be a DNS hostname")

				}
				return resp, nil
			},
		},
		{
			key:    "email",
			prompt: "Enter an email for LetsEncrypt registration.\nThis will be used when provisioning a TLS certificate.\n> ",
			process: func(resp string) (string, error) {
				resp = strings.TrimSpace(resp)
				if !govalidator.IsEmail(resp) {
					return "", errors.New("input must be an email address")

				}
				return resp, nil
			},
		},
		{
			key:    "pinRootCA",
			prompt: "Should the LetsEncrypt CA root certificate be pinned by the client? (yes/no)\n> ",
			process: func(resp string) (string, error) {
				resp = strings.TrimSpace(resp)
				if resp != "yes" && resp != "no" {
					return "", errors.New("input must be yes or no")

				}
				return resp, nil
			},
		},
		{
			key:    "tlsMaxVersion",
			prompt: "Set the maximum TLS version that should be used, 1.2 or 1.3\n> ",
			process: func(resp string) (string, error) {
				resp = strings.TrimSpace(resp)
				if resp != "1.2" && resp != "1.3" {
					return "", errors.New("input must be one of 1.2 or 1.3")
				}
				return resp, nil
			},
		},
	},
}
