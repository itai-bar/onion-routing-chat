import socket
import sys
import key_exchange as ke
import const
import serialize

def main():
    message = ""
    route_ips = get_nodes_and_dst_ips()  # [1st node, 2nd node, 3rd node, dst_ip]
    sock_with_server = connect_to_server(route_ips[const.ST_NODE_IP_IDX], 8989)
    ke.key_exchange(route_ips, sock_with_server)
    
    while message != "Exit":
        message = input("Enter message('Exit' to exit):")
        message_to_send = serialize.serialize_tor_message(message, route_ips[const.ND_NODE_IP_IDX:])
        print(message_to_send)
        print("response:", send_message_and_get_response(sock_with_server, message_to_send).decode())
        
    sock_with_server.close()


def send_message_and_get_response(sock_with_server, message_to_send):
    """The function send message to the given socket and return the response

    Args:
        sock_with_server (socket): Socket to send it the message
        message_to_send (string): Message to send to the server through the socket

    Returns:
        string.encode(): Response from the server
    """

    sock_with_server.sendall(message_to_send.encode())
    return sock_with_server.recv(len(message_to_send))  # now we are dealing with echo server, going to be changed in the progress of the project


def connect_to_server(ip, port):
    """The function creates TCP socket, create connection with given 'ip' and 'port' and returns the connected socket

    Args:
        ip (string): Ip of server
        port (int): Port to get his service

    Returns:
        socket: Socket connected to given ip and port
    """
    
    # Create a TCP/IP socket
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    # Connect the socket to the server
    server_address = (ip, port)
    sock.connect(server_address)
    print('connected to ', server_address)
    return sock

def get_nodes_and_dst_ips():
    """Get nodes ip's from argv

    Returns:
        list(string): Ip's that stored in nodes.txt file
    """

    nodes = []
    for i in range(1,5):
        try:
            nodes.append(sys.argv[i])
        except IndexError as e:
            print("Please enter 4 ip's as arguments: client.py node_ip1 node_ip2 node_ip3 ip_of_destination")
            exit(0)
    return nodes


if __name__ == "__main__":
    main()
