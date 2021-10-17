package config

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"lukechampine.com/frand"
)

// LoadConfig loads a configuration from a file.
func LoadConfig(filepath string) (Configuration, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var conf Configuration
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return conf, verify(conf)
}

func writeConfig(config Configuration) (string, error) {
	data, err := config.JSON()
	if err != nil {
		return "", err
	}
	filename, err := filepath.Abs(randConfigFileName())
	if err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return filename, err
	}
	return filename, nil
}

func randConfigFileName() string {
	return base64.RawURLEncoding.EncodeToString(frand.Bytes(6)) + ".json"
}
