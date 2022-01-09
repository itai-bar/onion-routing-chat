package node

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"torbasedchat/pkg/tor_aes"
	"torbasedchat/pkg/tor_rsa"
	"torbasedchat/pkg/tor_server"
)

const (
	CLOSE_SOCKET_SIZE      = 1
	IP_SEGMENT_SIZE        = 15
	DATA_SIZE_SEGMENT_SIZE = 5
	CODE_NODE_CONN         = "00"
	CODE_NODE_DIS          = "01"
)

type TorHeaders struct {
	closeSocket int
	nextIp      string
	rest        []byte
}

func HandleClient(conn net.Conn) {
	var nextNodeConn net.Conn
	socketOpenFlag := false

	// first we do a key exchange with the client
	aes_key, err := ExchangeKey(conn)
	if err != nil {
		conn.Close()
		return
	}

	// the transfering loop will end once
	// the client will turn the CLOSE_SOCKET flag on
	for {
		allData, err := tor_server.ReadDataFromSizeHeader(conn, DATA_SIZE_SEGMENT_SIZE)
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}

		allData, err = aes_key.Decrypt(allData)
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}

		headers, err := GetTorHeaders(allData)
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}

		// avoiding opening the socket in a loop
		if !socketOpenFlag {
			nextNodeConn, err = net.Dial("tcp", headers.nextIp+":8989")
			if err != nil {
				log.Println("ERROR: ", err)
				return
			}
			socketOpenFlag = true
		}

		resp, err := TransferMessage(nextNodeConn, headers.rest, aes_key)
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}
		conn.Write(resp)
		log.Println("sent response back")

		if headers.closeSocket == 1 {
			conn.Close()
			nextNodeConn.Close()
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
func GetTorHeaders(allData []byte) (*TorHeaders, error) {
	allDataReader := bytes.NewReader(allData)
	closeSocketBuf := make([]byte, CLOSE_SOCKET_SIZE)

	_, err := allDataReader.Read(closeSocketBuf)
	if err != nil {
		return nil, err
	}

	closeSocket, _ := strconv.Atoi(string(closeSocketBuf))

	// reading the next node/dst ip
	nextIpBuf := make([]byte, IP_SEGMENT_SIZE)
	_, err = allDataReader.Read(nextIpBuf)
	if err != nil {
		return nil, err
	}

	nextIp := string(tor_server.RemoveLeadingChars(nextIpBuf, '0')) // ip might come with padding

	rest := make([]byte, allDataReader.Len())
	_, err = allDataReader.Read(rest)
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
func TransferMessage(conn net.Conn, req []byte, aes_key *tor_aes.Aes) ([]byte, error) {
	//adding the size of the rest of the request
	paddedLen := fmt.Sprintf("%05d", len(req))
	req = append([]byte(paddedLen), req...)

	// sending the request to the next part of the path
	conn.Write(req)
	log.Println("forwarded the rest of the message")
	// from now we expect a response from the rest of the network

	data, err := tor_server.ReadDataFromSizeHeader(conn, DATA_SIZE_SEGMENT_SIZE)
	if err != nil {
		return nil, err
	}

	encryptedData, err := aes_key.Encrypt(data)
	if err != nil {
		return nil, err
	}

	// appending the data size and data back together
	paddedLen = fmt.Sprintf("%05d", len(encryptedData))
	resp := append([]byte(paddedLen), encryptedData...)

	return resp, nil
}

/*
	performs a key change with the client using a given RSA key
*/
func ExchangeKey(conn net.Conn) (*tor_aes.Aes, error) {
	pemKey, err := tor_server.ReadDataFromSizeHeader(conn, DATA_SIZE_SEGMENT_SIZE)
	if err != nil { // error can occur when router trying to check if node is alive
		return nil, err
	}

	// inits a rsa object with the key we got from the client
	// creating the aes key for the rest of comm
	rsa, err := tor_rsa.NewRsaGivenPemPublicKey(pemKey)
	aes := tor_aes.NewAesRandom()
	if err != nil {
		return nil, err
	}

	log.Println("got rsa key from client")

	buf, err := rsa.Encrypt(aes.Key)
	if err != nil {
		return nil, err
	}

	// padding the length with to 5 chars
	paddedLen := fmt.Sprintf("%05d", len(buf))
	buf = append([]byte(paddedLen), buf...)

	conn.Write(buf)

	log.Println("sent rsa encrypted aes key")
	return aes, nil
}
