package node

import (
	"bytes"
	"fmt"
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
	CODE_NODE_CONN = "00"
	CODE_NODE_DIS  = "01"
)

type TorHeaders struct {
	closeSocket int
	nextIp      string
	rest        []byte
}


/*
	trying to perform with router, login to tor-network

	routerAddress string: "ip:port"

	return-value: true - connection agreed. false - connection refused.
*/
func NetworkLogon(routerAddress string) bool {
	routerConn, err := net.Dial("tcp", routerAddress)
	if err != nil {
		log.Println("ERROR: ", err)
		return false
	}

	routerConn.Write([]byte(CODE_NODE_CONN))
	
	respBuf := make([]byte, 1)
	_, err = routerConn.Read(respBuf)
	if err != nil {
		return false
	}

	routerConn.Close()

	passed := []byte("1") // "1" for true and performed logon succesfully
	res := bytes.Compare(respBuf, passed)


	return res==0 //if res==0-> connection passed. else-> connection refused
}

/*
	disconnect from tor-network

	routerAddress string: "ip:port"
*/
func NetworkLogout(routerAddress string) {
	routerConn, err := net.Dial("tcp", routerAddress)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	routerConn.Write([]byte(CODE_NODE_DIS))

	routerConn.Close()

	log.Println("Logged out from tor-network")
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
	var nextNodeConn net.Conn
	socketOpenFlag := false

	// first we do a key exchange with the client
	aes_key, err := ExchangeKey(conn)
	if err != nil {
		log.Println("ERROR: ", err)
	}

	// the transfering loop will end once
	// the client will turn the CLOSE_SOCKET flag on
	for {
		allData, err := GetAllDataFromSocket(conn)
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
	function get all read all data from socket by first 5 bytes that are the size of the rest of th content
*/
func GetAllDataFromSocket(conn net.Conn) ([]byte, error) {
	dataSizeBuf := make([]byte, DATA_SIZE_SEGMENT_SIZE)
	_, err := conn.Read(dataSizeBuf)
	if err != nil {
		return nil, err
	}

	dataSize, err := strconv.Atoi(string(RemoveLeadingChars(dataSizeBuf, '0')))
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

	nextIp := string(RemoveLeadingChars(nextIpBuf, '0')) // ip might come with padding

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
	lenBuf := make([]byte, DATA_SIZE_SEGMENT_SIZE)
	_, err := conn.Read(lenBuf)
	if err != nil {
		return nil, err
	}

	length, err := strconv.Atoi(string(RemoveLeadingChars(lenBuf, '0')))
	if err != nil {
		return nil, err
	}
	
	pemKey := make([]byte, length)
	_, err = conn.Read(pemKey)
	if err != nil {
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
