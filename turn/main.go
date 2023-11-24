package main

import (
	"fmt"
	"gorelay/turn/server"
	"runtime"
)

const (
	portForBackend = 2777
	portForClients = 2778
)

// Мы релей. Ждем приветственный пакет от прокси бекенда.
// Далее ждем подключений от клиентов.
func main() {
	s := server.NewServer(portForBackend, portForClients)
	Err(s.Serve())
}

func Err(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Println(fmt.Errorf("error is occured on %s:%d: %w", file, line, err))
		panic(err)
	}
}
