package main

import (
	"errors"

	"github.com/awnumar/rosen/config"
	"github.com/awnumar/rosen/protocols/https"
	"github.com/awnumar/rosen/protocols/wss"
	"github.com/awnumar/rosen/router"
)

func server(conf config.Configuration) (err error) {
	var server router.Server

	switch conf["protocol"] {
	case "":
		return errors.New("protocol must be specified in config file")
	case "wss":
		server, err = wss.NewServer(conf)
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
