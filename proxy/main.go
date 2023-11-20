package main

import (
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8888")
	Err(err)

	for {
		conn, err := listener.Accept()
		Err(err)
		handle(conn)
	}
}
func Err(err error) {
	if err != nil {
		panic(err)
	}
}
