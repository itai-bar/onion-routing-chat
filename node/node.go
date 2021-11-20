package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
)

const (
	IP_SEGMENT_SIZE        = 15
	DATA_SIZE_SEGMENT_SIZE = 5
)

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

		go TransferMessage(conn) // new thread to handle the client
	}
}

/*
	handles connection with a client
	for now there is only one available service - transfering data

	data transfering message:
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
func TransferMessage(conn net.Conn) {
	defer conn.Close()
	// reading the next node/dst ip
	nextIpBuf := make([]byte, IP_SEGMENT_SIZE)
	_, err := conn.Read(nextIpBuf)
	if err != nil {
		log.Println("err: ", err)
		return
	}

	nextIp := string(RemoveLeadingChars(nextIpBuf, '0')) // ip might come with padding

	// reading the rest of the message
	bufReader := bufio.NewReader(conn)
	buf, err := bufReader.ReadBytes(0)
	if err != nil {
		log.Println("err: ", err)
		return
	}

	log.Printf("sending %s to %s", string(buf), nextIp)

	// sending the rest of the message forward
	resp, err := SendToNextNode(nextIp, buf)
	if err != nil {
		log.Println("err: ", err)
		return
	}

	log.Println("sending back: ", string(resp))
	// returning the resp to the original requester
	conn.Write(resp)
}

/*
	sending the message to a given ip, wating for response
	that looks likes that:
		2 Bytes 	(data size)
		data size 	(data)
	and returns it

	nextIp string: the ip cutted from the original data,
	should connect to it
	req []byte: the req to send forward
*/
func SendToNextNode(nextIp string, req []byte) ([]byte, error) {
	c, err := net.Dial("tcp", nextIp+":8989")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// sending the request to the next part of the path
	c.Write(req)
	// from now we expect a response from the rest of the network

	// reading data size
	dataSizeBuf := make([]byte, DATA_SIZE_SEGMENT_SIZE)
	_, err = c.Read(dataSizeBuf)
	if err != nil {
		return nil, err
	}

	dataSize, _ := strconv.Atoi(string(RemoveLeadingChars(dataSizeBuf, '0')))
	data := make([]byte, dataSize)
	_, err = c.Read(data)
	if err != nil {
		return nil, err
	}

	// appending the data size and data back together
	return append(dataSizeBuf, data...), nil
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
