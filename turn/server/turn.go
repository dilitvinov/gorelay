package server

import (
	"bytes"
	"fmt"
	"net"
	"runtime"
)

var HelloPacket = []byte("HELLO_TURN")

const (
	size      = 2 << 15
	protocol  = "udp"
	localAddr = "0.0.0.0"
	// localAddr = "127.0.0.1"
)

type (
	Server struct {
		vacantBackends          chan net.Addr
		addrMap                 map[string]sThread
		forBackend, forClients  *net.UDPAddr
		backendSock, clientSock *net.UDPConn
	}

	sThread struct {
		addr   net.Addr
		sock   *net.UDPConn
		dataCh chan []byte
	}
)

func NewServer(backendPort, clientsPort int) *Server {
	s := &Server{
		vacantBackends: make(chan net.Addr, 10),
		addrMap:        make(map[string]sThread),
		forBackend:     &net.UDPAddr{IP: net.ParseIP(localAddr), Port: backendPort},
		forClients:     &net.UDPAddr{IP: net.ParseIP(localAddr), Port: clientsPort},
	}
	return s
}

func (s *Server) Serve() error {
	var err error
	s.backendSock, err = net.ListenUDP(protocol, s.forBackend)
	if err != nil {
		return fmt.Errorf("failed to listen backendSock: %w", err)
	}
	err = s.backendSock.SetReadBuffer(size)
	if err != nil {
		return fmt.Errorf("failed to SetReadBuffer for backendSock: %w", err)
	}

	s.clientSock, err = net.ListenUDP(protocol, s.forClients)
	if err != nil {
		return fmt.Errorf("failed to listen clientSock: %w", err)
	}
	err = s.clientSock.SetReadBuffer(size)
	if err != nil {
		return fmt.Errorf("failed to SetReadBuffer for clientSock: %w", err)
	}

	go s.readFromBackend()
	s.readFromClients()
	return nil
}

func (s *Server) readFromBackend() {
	for {
		slice := make([]byte, size)
		n, remoteAddr, err := s.backendSock.ReadFrom(slice)
		Err(err)
		if sock, ok := s.addrMap[remoteAddr.String()]; !ok {
			// the first packet from remote. Check HelloPacket
			if bytes.Equal(slice[:len(HelloPacket)], HelloPacket) {
				println("accepted new connection from backend " + remoteAddr.String())
				// prepare backendSock for write
				s.vacantBackends <- remoteAddr
			} else {
				// drop packet
				continue
			}
		} else {
			// known backend
			sock.dataCh <- slice[:n]
		}
	}
}

func (s *Server) readFromClients() {
	for {
		slice := make([]byte, size)
		n, clientAddr, err := s.clientSock.ReadFromUDP(slice)
		println("accepted from client " + clientAddr.String())
		Err(err)
		if sock, ok := s.addrMap[clientAddr.String()]; !ok {
			// new client
			backendAddr := <-s.vacantBackends

			stBackend := sThread{
				addr:   backendAddr,
				sock:   s.backendSock,
				dataCh: make(chan []byte, 10),
			}
			go stBackend.transmitData()
			s.addrMap[clientAddr.String()] = stBackend

			stClient := sThread{
				addr:   clientAddr,
				sock:   s.clientSock,
				dataCh: make(chan []byte, 10),
			}
			go stClient.transmitData()
			s.addrMap[backendAddr.String()] = stClient
			stBackend.dataCh <- slice[:n]
		} else {
			sock.dataCh <- slice[:n]
			Err(err)
		}
	}
}

func (st *sThread) transmitData() {
	for data := range st.dataCh {
		n, err := st.sock.WriteTo(data, st.addr)
		if err != nil {
			fmt.Println("error transmitData: %w", err)
		} else {
			fmt.Println(fmt.Sprintf("wrote from %s to %s %d byte!", st.sock.LocalAddr().String(), st.addr.String(), n))
		}
	}
}

func Err(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Println(fmt.Errorf("error is occured on %s:%d: %w", file, line, err))
	}
}
