package router

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
	"torbasedchat/pkg/tor_rsa"
	"torbasedchat/pkg/tor_server"
)

const (
	REQ_CODE_SIZE          = 2
	CODE_NODE_CONN         = "00"
	CODE_NODE_DIS          = "01"
	CODE_ROUTE             = "11"
	DATA_SIZE_SEGMENT_SIZE = 5
)

var networkLock sync.Mutex
var network TorNetwork

type TorNetwork map[string]struct{
	lastEcho time.Time
}

func init() {
	network = make(TorNetwork)
}

/*
	the router will be handling 3 different requests:
		* node connection
		* node disconnection
		* tor client route request
*/
func HandleClient(conn net.Conn) {
	msgCode, err := tor_server.ReadSize(conn, REQ_CODE_SIZE)
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
		network[ip] = struct{lastEcho time.Time}{time.Now()}             // init en empty struct to the map
		conn.Write([]byte("1"))              // "1" for true - it means joined successfully
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
func RandMapItem(network TorNetwork) (k string, v struct{lastEcho time.Time}) {
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
	allData, err := tor_server.ReadDataFromSizeHeader(conn, DATA_SIZE_SEGMENT_SIZE)
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

func CheckNodes(){
	for {
		for node := range network{
			currTime := time.Now()
			diff := currTime.Sub(network[node].lastEcho)
			if diff.Minutes() >= 0.2{
				log.Println("sending echo to " + node)
				if isAlive(node){
					log.Println(node + " is Alive")
					network[node] = struct{lastEcho time.Time}{time.Now()}
				}else{
					log.Println(node + " is Dead")
					delete(network, node)
				}
			}
		}
		time.Sleep(5*time.Second)
	}
}

func isAlive(ipAddr string) bool{
	conn, err:= net.Dial("tcp", ipAddr+":8989")
	if err != nil{ // if err occured, then node is dead
		return false
	}
	conn.Close()
	return true
}