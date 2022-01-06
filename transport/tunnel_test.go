package transport

import (
	"encoding/base64"
	"testing"

	"github.com/matryer/is"
	"lukechampine.com/frand"

	"github.com/awnumar/rosen/router"
)

func TestReadWriteTunnel(t *testing.T) {
	is := is.New(t)

	A, B, err := setupLocalConn()
	is.NoErr(err)
	defer A.Close()
	defer B.Close()

	key := frand.Bytes(32)

	tA, err := NewTunnel(A, key)
	is.NoErr(err)
	tB, err := NewTunnel(B, key)
	is.NoErr(err)

	refData := randomPacketSeq(100)

	is.NoErr(tA.Send(refData))

	readData, err := tB.Recv()
	is.NoErr(err)

	// ugly hack since testing library doesn't consider nil and len(0) slices to be equivalent
	for i := range readData {
		if len(readData[i].Data) == 0 {
			readData[i].Data = nil
		}
	}

	is.Equal(refData, readData)
}

func randomPacketSeq(length int) (packets []router.Packet) {
	for i := 0; i < length; i++ {
		packets = append(packets, randomPacket())
	}
	return
}

func randomPacket() router.Packet {
	return router.Packet{
		ID: base64.RawURLEncoding.EncodeToString(frand.Bytes(16)),
		Dest: router.Endpoint{
			Network: base64.RawURLEncoding.EncodeToString(frand.Bytes(16)),
			Address: base64.RawURLEncoding.EncodeToString(frand.Bytes(16)),
		},
		Data: frand.Bytes(frand.Intn(4096)),
		Type: router.Data,
	}
}
