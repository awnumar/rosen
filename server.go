package main

import (
	"errors"

	"github.com/awnumar/rosen/protocols/config"
	"github.com/awnumar/rosen/protocols/https"
	"github.com/awnumar/rosen/proxy"
)

func server(conf config.Configuration) (err error) {
	var server proxy.Server

	switch conf["protocol"] {
	case "":
		return errors.New("protocol must be specified in config file")
	case "https":
		server, err = https.NewServer(conf)
	default:
		return errors.New("unknown protocol: " + conf["protocol"])
	}
	if err != nil {
		return err
	}

	return server.Start()
}
