package main

const (
	portForBackend = 2777
	portForClients = 2778
)

var HelloPacket = []byte("HELLO_TURN")

// Мы релей. Ждем приветственный пакет от прокси бекенда.
// Далее ждем подключений от клиентов.
// TODO Err(clientSock.SetReadBuffer(size))
func main() {
	s := NewServer(portForBackend, portForClients)
	Err(s.Serve())
}

func Err(err error) {
	if err != nil {
		panic(err)
	}
}
