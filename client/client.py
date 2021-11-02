import socket


def main():
    message_to_send = ""
    sock_with_server = connect_to_server("127.0.0.1", 8989)
    while message_to_send != "Exit":
        message_to_send = input("Enter message('Exit' to exit):")
        print("response:", send_message_and_get_response(sock_with_server, message_to_send).decode())
    sock_with_server.close()


def send_message_and_get_response(sock_with_server, message_to_send):
    """
    The function send message to the given socket and return the response
    :param sock_with_server: Socket to send it the message
    :param message_to_send: Message to send to the server through the socket
    :return: Response from the server
    """
    sock_with_server.sendall(message_to_send.encode())
    return sock_with_server.recv(len(message_to_send))  # now we are dealing with echo server, going to be changed in the progress of the project


def connect_to_server(ip, port):
    """
    The function creates TCP socket, create connection with given 'ip' and 'port' and returns the connected socket
    :param ip: Ip of server
    :param port: Port to get his service
    :return: Socket connected to given ip and port
    """
    # Create a TCP/IP socket
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    # Connect the socket to the server
    server_address = (ip, port)
    sock.connect(server_address)
    return sock


if __name__ == "__main__":
    main()
