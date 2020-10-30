<img align="right" src="icon.svg" width="150px">

# Rosen

[![documentation](https://godoc.org/github.com/awnumar/rosen?status.svg)](https://godoc.org/github.com/awnumar/rosen)

This package implements a modular framework for proxies that encapsulate traffic within some cover protocol to circumvent censorship based on deep packet inspection and endpoint fingerprinting techniques.

**This package is currently pre-alpha and is considered experimental.**

### Supported protocols

- https

### Installation

```
go get github.com/awnumar/rosen
```

### Usage

Run the configuration tool to create a config file.

```
rosen -configure
```

Then on the server side run

```
rosen -mode server -config example.json
```

And finally on the client side run

```
rosen -mode client -proxyAddr https://example.com -config example.json
```

This will launch a SOCKS server on the default port.

### Future development

- Support other cover protocols.
- Auto HTTPS certificate provision with LetsEncrypt.
- Verify SOCKS server supports UDP and IPv6.
- TUN/TAP support in addition to SOCKS.
- Support multiple clients per server.

### License

This is public domain software. See [LICENSE](/LICENSE) for details.
