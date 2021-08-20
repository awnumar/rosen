<img align="right" src="icon.svg" width="150px">

# Rosen

[![documentation](https://godoc.org/github.com/awnumar/rosen?status.svg)](https://godoc.org/github.com/awnumar/rosen)

This package implements a modular framework for proxies that encapsulate traffic within some cover protocol to circumvent censorship based on deep packet inspection and endpoint fingerprinting techniques. For more information, check out [this](http://spacetime.dev/rosen-censorship-resistant-proxy-tunnel) post.

### Goals

1. **Unobservability**. It should be difficult to distinguish obfuscated traffic from innocent background traffic using the same protocol.
2. **Endpoint-fingerprinting resistance**. It should be difficult to use active probing to ascertain that a given server is actually a proxy server. This is accomplished by responding as a proxy if and only if a valid key is provided and falling back to some default behaviour otherwise.
3. **Modularity**. It should be relatively easy to add support for another cover protocol or configure the behaviour of an existing protocol to adapt to changing adversarial conditions. This is facilitated by a modular architecture.
4. **Compatibility**. It should be possible to route most application traffic through the proxy. This is why a SOCKS interface was chosen, but TUN support is also a goal.
5. **Performance**. It should be fast and have minimal overhead.
6. **Usability**. It should be easy to use.

### Supported protocols

- Encrypted WebSockets
- HTTPS

### Installation

Requires Go version 1.16 or above.

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
rosen -mode client -config example.json
```

This will launch a SOCKS server on the default port (23579). Use the `-help` flag to see other options.

### Future development

- Verify SOCKS server supports UDP and IPv6.
- TUN support in addition to SOCKS.
- Support other cover protocols.
- Support multiple clients per server.
- Tests.

### License

This is public domain software. See [LICENSE](/LICENSE) for details.
