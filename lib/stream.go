package lib

import (
	"fmt"
	"io"
)

// indAuthStream is a queue that implements io.ReadWriter such that,
//   All data that is written to the queue is 
type indAuthStream struct {
	r   io.Reader
	k   []byte
	b   []byte
	p   int
	err error
}

func NewAeGob(r io.Reader, key []byte) *Ae_gob {
	return &Ae_gob{r: r, b: nil, k: key}
}

func (g *Ae_gob) Read(b []byte) (int, error) {
	read, err := io.ReadAll(g.r)
	if err != nil {
		fmt.Println(err)
	}
	g.b, err = Ae_decrypt(read, g.k)
	if err != nil {
		fmt.Println(err)
	}

	return 0, g.err
}

func (g *Ae_gob)