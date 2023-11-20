package main

import "net"

// try twice
func main() {
	conn, err := net.Dial("udp", "127.0.0.1:2778")
	Err(err)
	for i := 0; i < 1; i++ {
		conn.Write([]byte("HELLO FROM VPN CLIENT!"))
		conn.Write([]byte("AND AGAIN"))
	}
}

func Err(err error) {
	if err != nil {
		panic(err)
	}
}
