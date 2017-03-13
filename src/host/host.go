package host

import (
	"errors"
	"fmt"
	"log"
	"net"

	dbg "github.com/JuanCabre/go-debug"
)

var debug = dbg.Debug("Host")

type Host struct {
	IPAddr net.IP

	UDPConns []*net.UDPConn
	TCPConns []*net.TCPConn

	Services map[string]string
	Ports    map[string]string
}

// NewHost Creates a new host with the given IP Address
func NewHost(addr string) (*Host, error) {
	h := new(Host)
	h.Services = make(map[string]string)
	h.IPAddr = net.ParseIP(addr) // IP Address

	return h, nil
}

// NewService creates a new service in the host. Network can be tcp or udp, and
// it specify the transport protocol for the service. The service will be
// running in the given port. The handler function will handle the packets that
// arrive for the given service. The signature of handler can be:
// func(net.PacketConn) or func(*net.UDPConn) for udp services;
// func(net.Conn) or func(*net.TCPConn) for tcp services.
func (h *Host) NewService(network, name, port string, handler interface{}) error {

	// Service will be of the form IPV4:Port. E.g., 127.0.0.1:9000
	service := h.IPAddr.String() + ":" + port

	// Add service to the supported services map
	if port, present := h.Ports[name]; present {
		return fmt.Errorf("The service %q is alredy registered at the host at port %q", name, port)
	}
	if service, present := h.Services[port]; present {
		return fmt.Errorf("The port %q is alredy used by the service %q", port, service)
	}
	h.Services[port] = name
	h.Ports[name] = port

	if network == "udp" { // If the network is udp
		// Resolve the UDP Address
		UDPAddr, err := net.ResolveUDPAddr(network, service)
		if err != nil {
			return err
		}
		// Create the packet-oriented network
		conn, err := net.ListenUDP(network, UDPAddr)
		if err != nil {
			return err
		}

		// Do a type assertion to make sure that the handler function has the
		// correct signature
		switch assertedHandler := handler.(type) {
		case func(net.PacketConn):
			go h.listenPackets(conn, assertedHandler) // Call the handler in the background
		case func(*net.UDPConn):
			go h.listenPackets(conn, assertedHandler)
		default:
			return errors.New("NewService: The handler is not a valid function")
		}

	} else if network == "tcp" { // If the network is tcp
		addr, err := net.ResolveTCPAddr("tcp", service)
		if err != nil {
			return err
		}

		listener, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}

		// Do a type assertion to make sure that the handler function has the
		// correct signature
		switch assertedHandler := handler.(type) {
		case func(net.Conn):
			go h.listenConn(listener, assertedHandler)
		case func(*net.TCPConn):
			go h.listenConn(listener, assertedHandler)
		default: // Return error if the handler has not a valid signature
			return errors.New("NewService: The handler is not a valid function")
		}

	} else { // If it is an unknown network, return error
		return errors.New("NewService: Unknown network")
	}

	return nil
}

// listenConn method listens and accepts a connection on the given listener.
// When the connection is received, it calls the handler function as a go
// routine.
func (h *Host) listenConn(listener *net.TCPListener, handler interface{}) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Error listenConn: ", err)
			continue
		}
		switch assertedHandler := handler.(type) {
		case func(net.Conn):
			go assertedHandler(conn)
		case func(*net.TCPConn):
			go assertedHandler(conn)
		default:
			log.Fatalln("listenConn: The handler is not a valid function")
		}
	}
}

func (h *Host) listenPackets(conn *net.UDPConn, handler interface{}) {
	switch assertedHandler := handler.(type) {
	case func(net.PacketConn):
		for {
			assertedHandler(conn)
		}
	case func(*net.UDPConn):
		for {
			assertedHandler(conn)
		}
	default:
		log.Fatalln("listenConn: The handler is not a valid function")
	}
}
