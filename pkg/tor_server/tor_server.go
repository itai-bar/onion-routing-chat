package tor_server

import (
	"net"
	"strconv"
	"torbasedchat/pkg/tor_logger"
)

// the servers has to make their handler like that
type ClientHandler func(net.Conn)

/*
	inits the server with an address and
	waits for client to handle with the given client handler
*/
func RunServer(address string, clientHandler ClientHandler, logger *tor_logger.TorLogger) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Err.Fatal("Error in listening:\t", err)
	}
	defer listener.Close() // will close the listener when the function exits
	logger.Info.Println("Listening on:\t", address)

	for {
		conn, err := listener.Accept() // new client
		if err != nil {
			logger.Err.Fatal("Error on accepting client:\t", err)
		}
		logger.Info.Println("New client:\t", conn.RemoteAddr().String())

		go clientHandler(conn) // new thread to handle the client
	}
}

/*
	a simple wrap on conn.Read to make it nicer to read
*/
func ReadSize(conn net.Conn, size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := conn.Read(buf)
	return buf, err
}

/*
	helper that read data based on a size segment at
	the start of a message
*/
func ReadDataFromSizeHeader(conn net.Conn, sizeSegmentLen int) ([]byte, error) {
	dataSize, err := GetDataSize(conn, sizeSegmentLen)
	if err != nil {
		return nil, err
	}

	allData := make([]byte, dataSize)
	conn.Read(allData)
	if err != nil {
		return nil, err
	}

	return allData, nil
}

func GetDataSize(conn net.Conn, size int) (int, error) {
	dataSizeBuf := make([]byte, size)
	_, err := conn.Read(dataSizeBuf)
	if err != nil {
		return 0, err
	}

	dataSize, err := strconv.Atoi(string(RemoveLeadingChars(dataSizeBuf, '0')))
	if err != nil {
		return 0, err
	}

	return dataSize, nil
}

/*
	removes every char c from the start of byte array s
*/
func RemoveLeadingChars(s []byte, c byte) []byte {
	for i := range s {
		if s[i] != c {
			return s[i:]
		}
	}
	return []byte{}
}
