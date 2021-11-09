import socket
import os

def main():
    message_to_send = ""
    sock_with_server = connect_to_server("127.0.0.1", 8989)
    while message_to_send != "Exit":
        message_to_send = input("Enter message('Exit' to exit):")
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

def get_nodes():
    """Get nodes from file

    Returns:
        list: Ip's that stored in nodes.txt file
    """

    curr_path = os.path.dirname(__file__)
    path_to_nodes = curr_path + r"\..\nodes.txt"

    with open(path_to_nodes, "r") as nodes:
        return [node.strip() for node in nodes.readlines()]
    


def pad_ips(list_of_ips):
    """This function get list of ip's not padding, for example 1.2.3.4 and return each padded 001.002.003.004

    Args:
        list_of_ips (list): Ip's to pad, each byte to fill with 000

    Returns:
        list: Padded ip's as list
    """

    padded_ip, padded_ips = [], []
    
    for ip in list_of_ips:
        padded_ip = []
        for byte in ip.split("."):
            padded_ip.append(byte.zfill(3))
        padded_ips.append(".".join(padded_ip))

    return padded_ips


if __name__ == "__main__":
    main()
