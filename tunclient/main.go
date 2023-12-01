package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

const (
	// TODO is it enough?
	size = 2 << 15
)

const (
	HelloPacket   = "HELLO_TURN"
	TurnRelayAddr = "176.107.133.219:2777"
	BackendAddr   = "0.0.0.0:1194"
	// TurnRelayAddr = "0.0.0.0:2777"
	// BackendAddr   = "95.165.1.34:1194"

	Protocol = "udp"
)

// TODO
// 0. Реализовать множественные подключения на сервере
// 0. Глушить соединения и горутину, если простой 10 минут
// 1. Смержить файлы, сделать параметры
// 2. Выложить в github
// 3. Сделать Dockerfile
// 4. Подумать над оптимизацией (больше буфер? горутины на запись?)
// 5. Написать тесты

// Мы подключаемся к релею, ждем первые байты.
// Дальше открываем соединение с бекендом и начинаем обмен
func main() {

	turnAddr, err := getUDPAddr(TurnRelayAddr)
	Err(err)
	backendAddr, err := getUDPAddr(BackendAddr)
	Err(err)

	for {
		turnSock, err := net.DialUDP(Protocol, nil, turnAddr)
		Err(err)
		<-handle(backendAddr, turnSock)
	}

	// send HelloPacket
	//_, err = turnSock.Write([]byte(HelloPacket))
	//Err(err)
	//
	//go func() {
	//	for {
	//		slice := make([]byte, size)
	//		n, _, err := backendSock.ReadFrom(slice)
	//		Err(err)
	//		n, err = turnSock.Write(slice[:n])
	//		Err(err)
	//		fmt.Println(fmt.Sprintf("from %s to %s: success!", backendSock.RemoteAddr().String(), turnAddr.String()))
	//	}
	//}()
	//
	//for {
	//	slice := make([]byte, size)
	//	n, _, err := turnSock.ReadFrom(slice)
	//	Err(err)
	//	n, err = backendSock.Write(slice[:n])
	//	Err(err)
	//	fmt.Println(fmt.Sprintf("wrote to %s %d byte!", backendAddr.String(), n))
	//}
}

func getUDPAddr(s string) (*net.UDPAddr, error) {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil, err
	}

	nPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	return &net.UDPAddr{IP: net.ParseIP(host), Port: nPort}, nil
}

func Err(err error) {
	if err != nil {
		panic(err)
	}
}

type (
	connHandler struct {
		to   chan []byte
		from chan []byte

		relayConn   net.Conn
		server      net.Conn
		backendAddr *net.UDPAddr
		sigChan     chan Signal
	}

	Signal int
)

const (
	SigFirstBatch Signal = iota + 1
	SigConnClosed
)

// TODO дропать после 10 мин простоя
func handle(
	backendAddr *net.UDPAddr,
	relayConn net.Conn,
) <-chan Signal {

	h := connHandler{
		to:          make(chan []byte),
		from:        make(chan []byte),
		relayConn:   relayConn,
		sigChan:     make(chan Signal, 2),
		backendAddr: backendAddr,
	}

	go func() {
		// wait for the first bytes
		firstBytes := make([]byte, size)
		n, err := relayConn.Read(firstBytes)
		if err != nil {
			fmt.Println(err)
			err = nil
		}

		h.server, err = net.DialUDP(Protocol, nil, backendAddr)
		Err(err)

		h.to <- firstBytes[:n]
		h.sigChan <- SigFirstBatch
		h.startExchange()
	}()

	return h.sigChan
}

func (h *connHandler) startExchange() {
	h.retransmit(h.relayConn, h.server, h.to)
	h.retransmit(h.server, h.relayConn, h.from)
}

func (h *connHandler) retransmit(from, to net.Conn, buff chan []byte) {
	go h.read(from, buff)
	go h.write(to, buff)
}

func (h *connHandler) read(conn net.Conn, buff chan []byte) {
	slice := make([]byte, size)
	for {
		n, err := conn.Read(slice)
		if err != nil {
			log.Println("read err is occurred: %w", err)
			h.sigChan <- SigConnClosed
			return
		}
		buff <- slice[:n]
	}
}

func (h *connHandler) write(conn net.Conn, buff chan []byte) {
	for bts := range buff {
		_, err := conn.Write(bts)
		if err != nil {
			log.Println("write err is occurred: %w", err)
			h.sigChan <- SigConnClosed
			return
		}
	}
}
