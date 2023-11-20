package main

import (
	"fmt"
	"log"
	"net"
)

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
		_, err := relayConn.Read(firstBytes)
		if err != nil {
			fmt.Println(err)
			err = nil
		}

		h.server, err = net.ListenUDP(Protocol, h.backendAddr)
		Err(err)

		h.to <- firstBytes
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
