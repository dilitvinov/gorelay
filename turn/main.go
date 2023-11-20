package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
)

const (
	size           = 2 << 15 // ITS VERY IMPORTANT THING
	Protocol       = "udp"
	portForBackend = "2777"
	portForClients = "2778"
)

var HelloPacket = []byte("HELLO_TURN")

// Мы релей. Ждем приветственный пакет от прокси бекенда.
// Далее ждем подключений от клиентов.
func main() {
	vacantBackends := make(chan net.Addr, 10)
	addrMap := make(map[string]net.Addr)

	backAddr, err := getUDPAddr(portForBackend)
	Err(err)
	backendSock, err := net.ListenUDP(Protocol, backAddr)

	cliAddr, err := getUDPAddr(portForClients)
	Err(err)
	clientSock, err := net.ListenUDP(Protocol, cliAddr)
	Err(err)

	Err(err)
	err = backendSock.SetReadBuffer(size)
	Err(err)
	go func() {
		for {
			slice := make([]byte, size)
			n, remoteAddr, err := backendSock.ReadFrom(slice)
			Err(err)
			if clientAddr, ok := addrMap[remoteAddr.String()]; !ok {
				// the first packet from remote. Check HelloPacket
				if bytes.Equal(slice[:len(HelloPacket)], HelloPacket) {
					println("accepted from backend " + remoteAddr.String())
					// prepare backendSock for write
					vacantBackends <- remoteAddr
				} else {
					// drop packet
					continue
				}
			} else {
				// known backend
				n, err = clientSock.WriteTo(slice[:n], clientAddr)
				Err(err)
				fmt.Println(fmt.Sprintf("wrote to %s %d byte!", clientAddr.String(), n))
			}
		}
	}()

	// TODO Err(clientSock.SetReadBuffer(size))

	for {
		slice := make([]byte, size)
		n, clientAddr, err := clientSock.ReadFromUDP(slice)
		println("accepted from client " + clientAddr.String())
		Err(err)
		if backendAddr, ok := addrMap[clientAddr.String()]; !ok {
			// new client
			// todo add new connection and start goroutine
			backendAddr := <-vacantBackends
			addrMap[clientAddr.String()] = backendAddr
			addrMap[backendAddr.String()] = clientAddr

			n, err = backendSock.WriteTo(slice[:n], backendAddr)
			Err(err)
			fmt.Println(fmt.Sprintf("wrote to %s %d byte!", backendAddr.String(), n))
		} else {
			n, err = backendSock.WriteTo(slice[:n], backendAddr)
			fmt.Println(fmt.Sprintf("wrote to %s %d byte!", backendAddr.String(), n))
			Err(err)
		}
	}
}

func Err(err error) {
	if err != nil {
		panic(err)
	}
}

func getUDPAddr(s string) (*net.UDPAddr, error) {
	host, port, err := net.SplitHostPort("0.0.0.0:" + s) // TODO
	if err != nil {
		return nil, err
	}

	nPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	return &net.UDPAddr{IP: net.ParseIP(host), Port: nPort}, nil
}
