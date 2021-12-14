package node

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"torbasedchat/pkg/tor_aes"
	"torbasedchat/pkg/tor_rsa"
)

const (
	CLOSE_SOCKET_SIZE      = 1
	IP_SEGMENT_SIZE        = 15
	DATA_SIZE_SEGMENT_SIZE = 5
)

type TorHeaders struct {
	closeSocket int
	nextIp      string
	rest        []byte
}

/*
	runs the server on a given address
	calls HandleClient for a connected client

	address string: "ip:port"
*/
func RunNode(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Error in listening:\t", err)
	}
	defer listener.Close() // will close the listener when the function exits
	log.Println("Listening on:\t", address)

	for {
		conn, err := listener.Accept() // new client
		if err != nil {
			log.Fatal("Error on accepting client:\t", err)
		}
		log.Println("New client:\t", conn.RemoteAddr().String())

		go HandleClient(conn) // new thread to handle the client
	}
}

func HandleClient(conn net.Conn) {
	// first we do a key exchange with the client
	ExchangeKey(conn)

	// the transfering loop will end once
	// the client will turn the CLOSE_SOCKET flag on
	for {
		headers, err := GetTorHeaders(conn)
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}
		log.Printf("headers here:\nclose: %d\nnext: %s\nrest: %s\n",
			headers.closeSocket, headers.nextIp, headers.rest)

		nextNodeConn, err := net.Dial("tcp", headers.nextIp+":8989")
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}

		resp, err := TransferMessage(nextNodeConn, headers.rest)
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}
		conn.Write(resp)

		if headers.closeSocket == 1 {
			conn.Close()
			break
		}
	}
}

/*
	getting headers of tor message from the socket

	data transfering message:
		1 Byte		(close socket flag)
		15 Bytes 	(2nd node ip)
		15 Bytes 	(3rd node ip)
		15 Bytes 	(dst ip)
		2 Bytes 	(data size)
		data size 	(data)

	every node will cut the ip part when receveing the message
	and send the rest to the ip until the last two segments
	will reach the dst ip.

	conn net.Conn: connection with a client
*/
func GetTorHeaders(clientConn net.Conn) (*TorHeaders, error) {
	closeSocketBuf := make([]byte, CLOSE_SOCKET_SIZE)
	_, err := clientConn.Read(closeSocketBuf)
	if err != nil {
		return nil, err
	}
	closeSocket, _ := strconv.Atoi(string(closeSocketBuf))

	// reading the next node/dst ip
	nextIpBuf := make([]byte, IP_SEGMENT_SIZE)
	_, err = clientConn.Read(nextIpBuf)
	if err != nil {
		return nil, err
	}

	nextIp := string(RemoveLeadingChars(nextIpBuf, '0')) // ip might come with padding
	// reading the rest of the message
	bufReader := bufio.NewReader(clientConn)
	rest, err := bufReader.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	return &TorHeaders{closeSocket, nextIp, rest}, nil
}

/*
	sending the message to a given connection, wating for response
	that looks likes that:
		2 Bytes 	(data size)
		data size 	(data)
	and returns it

	conn net.Conn: connection with next node or final dst
	req []byte: the req to send forward
*/
func TransferMessage(conn net.Conn, req []byte) ([]byte, error) {
	// sending the request to the next part of the path
	conn.Write(req)
	// from now we expect a response from the rest of the network

	// reading data size
	dataSizeBuf := make([]byte, DATA_SIZE_SEGMENT_SIZE)
	_, err := conn.Read(dataSizeBuf)
	if err != nil {
		return nil, err
	}

	dataSize, _ := strconv.Atoi(string(RemoveLeadingChars(dataSizeBuf, '0')))
	data := make([]byte, dataSize)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}

	// appending the data size and data back together
	return append(dataSizeBuf, data...), nil
}

/*
	performs a key change with the client using a given RSA key
*/
func ExchangeKey(conn net.Conn) ([]byte, error) {
	lenBuf := make([]byte, DATA_SIZE_SEGMENT_SIZE)
	_, err := conn.Read(lenBuf)
	if err != nil {
		return nil, err
	}

	leng, _ := strconv.Atoi(string(RemoveLeadingChars(lenBuf, '0')))
	pemKey := make([]byte, leng)
	_, err = conn.Read(pemKey)
	if err != nil {
		return nil, err
	}

	pemKey = pemKey[0 : len(pemKey)-1]

	// inits a rsa object with the key we got from the client
	// creating the aes key for the rest of comm
	rsa, err := tor_rsa.NewRsaGivenPemPublicKey(pemKey)
	aes := tor_aes.NewAesRandom()
	if err != nil {
		return nil, err
	}

	buf, err := rsa.Encrypt(aes.Key)
	if err != nil {
		return nil, err
	}

	conn.Write(buf)
	return aes.Key, nil
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
