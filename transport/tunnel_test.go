package transport

import (
	"encoding/base64"
	"testing"

	"github.com/awnumar/rosen/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lukechampine.com/frand"
)

func TestReadWriteTunnel(t *testing.T) {
	A, B, err := setupLocalConn()
	require.NoError(t, err)
	defer A.Close()
	defer B.Close()

	key := frand.Bytes(32)

	tA, err := NewTunnel(A, key)
	require.NoError(t, err)
	tB, err := NewTunnel(B, key)
	require.NoError(t, err)

	refData := randomPacketSeq(100)

	require.NoError(t, tA.Send(refData))

	readData, err := tB.Recv()
	require.NoError(t, err)

	// ugly hack since assert doesn't consider nil and len(0) to be equal
	for i := range readData {
		if len(readData[i].Data) == 0 {
			readData[i].Data = nil
		}
	}

	assert.Equal(t, len(refData), len(readData))
	assert.Equal(t, refData, readData)
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
