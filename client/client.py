import socket
import sys
import key_exchange as ke
import const
import serialize
import crypto

def main():
    message = ""
    route_ips = get_nodes_and_dst_ips()  # [1st node, 2nd node, 3rd node, dst_ip]
    rsa_key_pair = crypto.Rsa()  # creating Rsa class with random keypair for all sessions

    while message != "Exit":
        message = input("enter message: ")
        if message == '':  # empty message is not allowed 
            continue
        resp = tor_message(message, route_ips, rsa_key_pair)
        print(resp.decode())

def tor_message(msg : str, route : list, rsa_key_pair : crypto.Rsa) -> bytes:
    """Sends a message using the tor protocol

    Args:
        msg (str): message for the final server
        route (list[str]): list of tor node ips
    Returns:
        resp (str)
    """
    sock_with_server = connect_to_server(route[const.ST_NODE_IP_IDX], 8989)
    
    aes_keys = ke.key_exchange(route[:-1], sock_with_server, rsa_key_pair)

    tor_msg = serialize.serialize_tor_message(msg, route[1:], True, aes_keys)
    print(tor_msg)

    sock_with_server.sendall(tor_msg)

    size = int(sock_with_server.recv(const.MESSAGE_SIZE_LEN).decode()) # reading plaintext size
    resp = sock_with_server.recv(size)
    
    resp = crypto.decrypt_by_order(resp, aes_keys)

    sock_with_server.close()
    return resp

def connect_to_server(ip : str, port : int) -> socket.socket:
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
    return sock

def get_nodes_and_dst_ips() -> list:
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
