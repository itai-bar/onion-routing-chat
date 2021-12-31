package router

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"torbasedchat/pkg/tor_rsa"
)

const (
	REQ_CODE_SIZE          = 2
	CODE_NODE_CONN         = "00"
	CODE_NODE_DIS          = "01"
	CODE_ROUTE             = "11"
	DATA_SIZE_SEGMENT_SIZE = 5
)

var networkLock sync.Mutex

type TorNetwork map[string]struct{}

/*
	runs the server on a given address
	calls HandleClient for a connected client

	address string: "ip:port"
*/
func RunRouter(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Error in listening:\t", err)
	}
	defer listener.Close() // will close the listener when the function exits
	log.Println("Listening on:\t", address)

	network := make(TorNetwork)

	for {
		conn, err := listener.Accept() // new client
		if err != nil {
			log.Fatal("Error on accepting client:\t", err)
		}
		log.Println("New client:\t", conn.RemoteAddr().String())

		go HandleClient(conn, network) // new thread to handle the client
	}
}

/*
	the router will be handling 3 different requests:
		* node connection
		* node disconnection
		* tor client route request
*/
func HandleClient(conn net.Conn, network TorNetwork) {
	msgCode := make([]byte, REQ_CODE_SIZE)
	_, err := conn.Read(msgCode)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	// the map is a mutual resource
	networkLock.Lock()
	defer networkLock.Unlock()

	switch string(msgCode) {
	case CODE_NODE_CONN:
		log.Println("got a node connection req")

		ip := conn.RemoteAddr().String()
		ip = ip[:strings.IndexByte(ip, ':')] //slice till the port without it
		network[ip] = struct{}{}             // init en empty struct to the map
		conn.Write([]byte("1"))              // "1" for true - it means joined succesfully
	case CODE_NODE_DIS:
		log.Println("got a node disconnection req")

		delete(network, conn.RemoteAddr().String())
	case CODE_ROUTE:
		log.Println("got a client routing req")

		SendRoute(conn, network)
	default:
		// TODO: send error msg to client
		log.Println("invalid req code")
		return
	}
}

/*
	generates a random route for a client
	this function has to be locked because its changes the network!
*/
func GenerateRoute(network TorNetwork) []string {
	tmpMap := make(TorNetwork)
	var ips []string

	for i := 0; i < 3; i++ {
		k, v := RandMapItem(network)
		tmpMap[k] = v        // saving for later
		ips = append(ips, k) // appending the new ip
		delete(network, k)   // deleting the key from the original map (would be restored)
	}

	for k, v := range tmpMap {
		network[k] = v // restoring the keys
	}

	return ips
}

// returns a random item from a tor network map
func RandMapItem(network TorNetwork) (k string, v struct{}) {
	rand.Seed(time.Now().UnixNano()) // initing seed
	i := rand.Intn(len(network))

	for k := range network {
		if i == 0 {
			return k, network[k]
		}
		i--
	}
	panic("would never happen")
}

func SendRoute(conn net.Conn, network TorNetwork) {
	route := GenerateRoute(network)
	allData, err := GetAllDataFromSocket(conn)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	rsa_key, err := tor_rsa.NewRsaGivenPemPublicKey(allData)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	encrypted, err := rsa_key.Encrypt([]byte(strings.Join(route[:], "&")))
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}
	paddedLen := fmt.Sprintf("%05d", len(encrypted))
	buf := append([]byte(paddedLen), encrypted...)

	conn.Write(buf)
}

// TODO: check if we are able to move this functions to be more generic instead of duplicate it in each code
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
