// transport/tcp.go
package transport

import (
	"context"
	"dns-server/types"
	"encoding/binary"
	"io"
	"net"
	"time"
)

type TCPServer struct {
	addr     string
	resolver types.Resolver
}

func NewTCPServer(addr string, r types.Resolver) *TCPServer {
	return &TCPServer{
		addr:     addr,
		resolver: r,
	}
}

func (s *TCPServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *TCPServer) handleConn(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// خواندن طول پیام (2 بایت Big Endian)
	var length uint16
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return
	}

	// خواندن بدنه پیام
	buf := make([]byte, length)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return
	}

	// resolve کردن
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.resolver.Resolve(ctx, buf)
	if err != nil {
		return
	}

	// نوشتن پاسخ (طول + بدنه)
	respLength := uint16(len(resp))
	binary.Write(conn, binary.BigEndian, respLength)
	conn.Write(resp)
}
