package router

// PacketType gives some information about the state of the connection that a packet belongs to.
type PacketType int

const (
	// Open signals to open a new connection.
	Open PacketType = iota

	// Data is a normal packet containing data.
	Data PacketType = iota

	// Close signals to close the connection and clean up.
	Close PacketType = iota
)

// Packet holds a single message to or from the server.
type Packet struct {
	ID   string
	Dest Endpoint
	Data []byte
	Type PacketType
}

// Endpoint refers to the remote host that the client wants to connect to.
type Endpoint struct {
	Network string
	Address string
}

// NewEndpoint initialises an Endpoint object.
func NewEndpoint(network, address string) Endpoint {
	return Endpoint{
		Network: network,
		Address: address,
	}
}

// NewConnection returns true if the message refers to a connection that has not yet been opened.
func (p Packet) NewConnection() bool {
	return p.Type == Open
}

// Closed checks if a message indicates a closed connection.
func (p Packet) Closed() bool {
	return p.Type == Close
}

// NewPacket returns a message for a new connection.
func NewPacket(id string, dest Endpoint) Packet {
	return Packet{
		ID:   id,
		Dest: dest,
		Type: Open,
	}
}

// DataPacket returns a data-containing message for an existing connection.
func DataPacket(id string, data []byte) Packet {
	return Packet{
		ID:   id,
		Data: data,
		Type: Data,
	}
}

// ClosePacket returns a message indicating a closed connection.
func ClosePacket(id string) Packet {
	return Packet{
		ID:   id,
		Type: Close,
	}
}
