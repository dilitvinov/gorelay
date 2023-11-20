package main

import (
	"fmt"
	"net"
)

const (
	target = "95.165.1.34:1194" // home vpn
	size   = 2 << 10
)

type connHandler struct {
	to   chan []byte
	from chan []byte

	client net.Conn
	server net.Conn
}

func handle(client net.Conn) {

	server, err := net.Dial("tcp", target)
	Err(err)

	h := connHandler{
		to:     make(chan []byte),
		from:   make(chan []byte),
		client: client,
		server: server,
	}

	h.startExchange()
}

func (h *connHandler) startExchange() {
	retransmit(h.client, h.server, h.to)
	retransmit(h.server, h.client, h.from)
}

func retransmit(from, to net.Conn, buff chan []byte) {
	go read(from, buff)
	go write(to, buff)
}

func read(conn net.Conn, buff chan []byte) {
	slice := make([]byte, size)
	for {
		n, err := conn.Read(slice)
		if err != nil {
			fmt.Println(err)
			err = nil
		}
		buff <- slice[:n]
	}
}

func write(conn net.Conn, buff chan []byte) {
	for {
		n, err := conn.Write(<-buff)
		Err(err)
		fmt.Println("write bytes: ", n)
	}
}
