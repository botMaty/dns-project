package transport

import (
	"context"
	"dns-server/types"
	"net"
	"time"
)

type UDPServer struct {
	addr     string
	resolver types.Resolver
}

func NewUDPServer(addr string, r types.Resolver) *UDPServer {
	return &UDPServer{
		addr:     addr,
		resolver: r,
	}
}

func (s *UDPServer) ListenAndServe() error {
	conn, err := net.ListenPacket("udp", s.addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, 512)

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		}
		go s.handlePacket(conn, addr, buf[:n])
	}
}

func (s *UDPServer) handlePacket(
	conn net.PacketConn,
	addr net.Addr,
	data []byte,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.resolver.Resolve(ctx, data)
	if err != nil {
		return
	}

	conn.WriteTo(resp, addr)
}
