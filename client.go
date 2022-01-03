package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/awnumar/rosen/config"
	"github.com/awnumar/rosen/protocols/https"
	"github.com/awnumar/rosen/protocols/tcp"
	"github.com/awnumar/rosen/router"

	"github.com/eahydra/socks"
)

func client(conf config.Configuration) (err error) {
	var client router.Client

	switch conf["protocol"] {
	case "":
		return errors.New("protocol must be specified in config file")
	case "tcp":
		client, err = tcp.NewClient()
	case "https":
		client, err = https.NewClient(conf)
	default:
		return errors.New("unknown protocol: " + conf["protocol"])
	}
	if err != nil {
		return err
	}

	dialer, err := newDialer(client)
	if err != nil {
		return err
	}

	s, err := socks.NewSocks5Server(dialer)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", socksPort))
	if err != nil {
		return err
	}

	return s.Serve(listener)
}

type dialer struct {
	tun        router.Client
	server     *net.TCPListener
	serverAddr *net.TCPAddr
}

func newDialer(tun router.Client) (*dialer, error) {
	server, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		return nil, err
	}

	return &dialer{
		tun:        tun,
		server:     server,
		serverAddr: server.Addr().(*net.TCPAddr),
	}, nil
}

func (d *dialer) Dial(network, address string) (net.Conn, error) {
	connChannel := make(chan net.Conn)
	errChannel := make(chan error)
	defer close(connChannel)
	defer close(errChannel)

	go func() {
		conn, err := d.server.Accept()
		if err != nil {
			errChannel <- err
		} else {
			connChannel <- conn
		}
	}()

	clientConn, err := net.DialTCP("tcp", nil, d.serverAddr)
	if err != nil {
		return nil, err
	}

	select {
	case err := <-errChannel:
		return nil, err
	case serverConn := <-connChannel:
		if err := d.tun.HandleConnection(router.NewEndpoint(network, address), serverConn); err != nil {
			fmt.Println(err)
		}
	}

	return clientConn, nil
}
