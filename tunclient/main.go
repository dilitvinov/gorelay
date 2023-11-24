package main

import (
	"fmt"
	"net"
	"strconv"
)

const (
	// TODO is it enough?
	size = 2 << 17
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

// Мы подключаемся к релею, ждем первые байты.
// Дальше открываем соединение с бекендом и начинаем обмен
func main() {

	turnAddr, err := getUDPAddr(TurnRelayAddr)
	Err(err)
	turnSock, err := net.DialUDP(Protocol, nil, turnAddr)
	Err(err)

	backendAddr, err := getUDPAddr(BackendAddr)
	Err(err)
	backendSock, err := net.DialUDP(Protocol, nil, backendAddr)
	Err(err)

	// send HelloPacket
	_, err = turnSock.Write([]byte(HelloPacket))
	Err(err)

	go func() {
		for {
			slice := make([]byte, size)
			n, _, err := backendSock.ReadFrom(slice)
			Err(err)
			n, err = turnSock.Write(slice[:n])
			Err(err)
			fmt.Println(fmt.Sprintf("from %s to %s: success!", backendSock.RemoteAddr().String(), turnAddr.String()))
		}
	}()

	for {
		slice := make([]byte, size)
		n, _, err := turnSock.ReadFrom(slice)
		Err(err)
		n, err = backendSock.Write(slice[:n])
		Err(err)
		fmt.Println(fmt.Sprintf("wrote to %s %d byte!", backendAddr.String(), n))
	}
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
