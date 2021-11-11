import socket
import sys

ST_NODE_IP_IDX = 0
ND_NODE_IP_IDX = 1
RD_NODE_IP_IDX = 2
DST_IP_IDX = 3

def main():
    message = ""
    route_ips = get_nodes_and_dst_ips()  # [1st node, 2nd node, 3rd node, dst_ip]
    sock_with_server = connect_to_server(route_ips[ST_NODE_IP_IDX], 8989)
    while message != "Exit":
        message = input("Enter message('Exit' to exit):")
        message_to_send = serialize_data_transfering_message(message, route_ips[ND_NODE_IP_IDX:])
        print("response:", send_message_and_get_response(sock_with_server, message_to_send).decode())
    sock_with_server.close()


def serialize_data_transfering_message(message, route_ips):
    """ Function creates protocoled message
    data transfering message:
        15 Bytes    (padded 2nd node ip)
        15 Bytes    (padded 3rd node ip)
        15 Bytes    (padded dst ip)
        2 Bytes     (padded data size[max is 65535])
        data size   (data)

    Args:
        message (string): message to send
        nodes_ips (list(string)): ip's of 2nd node and 3rd node. all padded with pad_ips func
        dst_ip (string): dstination ip padded with pad_ips func

    Returns:
        string: message suited to protocol
    """

    result = ""
    result += "".join(route_ips)  # Each node remove exact amount of bytes because padding
    result += str(len(message)).zfill(5)  # fill with zeros so the dst be able to read it with no problems
    result += message
    return result



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


def pad_ips(list_of_ips):
    """This function get list of ip's not padded, for example 1.2.3.4 and return each ip padded 00000001.2.3.4

    Args:
        list_of_ips (list): Ip's to pad, fill with leading zeros to len 15

    Returns:
        list: Padded ip's as list
    """

    padded_ips = []
    
    for ip in list_of_ips:
        padded_ips.append(ip.zfill(15))

    return padded_ips


if __name__ == "__main__":
    main()

