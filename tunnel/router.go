package tunnel

import "github.com/awnumar/rosen/router"

const bufferSize = 4096

// ProxyWithRouter starts some routines that continuously proxy data between the local Router
// endpoint and the local Tunnel endpoint. This function will block while proxying.
//
// For example, server-side proxy implementations can attach the client-side socket to a Tunnel,
// and and then attach a Router that holds connections to the outside world.
func (t *Tunnel) ProxyWithRouter(r *router.Router) error {
	routerToTunnelErr := make(chan error)
	go func() {
		buffer := make([]router.Packet, bufferSize)
		for {
			size := r.Fill(buffer)
			if err := t.Send(buffer[:size]); err != nil {
				routerToTunnelErr <- err
				close(routerToTunnelErr)
				return
			}
		}
	}()

	tunnelToRouterErr := make(chan error)
	go func() {
		for {
			data, err := t.Recv()
			if err != nil {
				tunnelToRouterErr <- err
				close(tunnelToRouterErr)
				return
			}
			r.Ingest(data)
		}
	}()

	var err error
	select {
	case err = <-routerToTunnelErr:
	case err = <-tunnelToRouterErr:
	}
	return err
}
