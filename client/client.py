import socket
import sys
import key_exchange as ke
import const
import serialize
import crypto
import validation

def main():
    message = ""
    if not validation.check_arguments_validation():
        exit(0)
    rsa_key_pair = crypto.Rsa()  # creating Rsa class with random keypair for all sessions
    route_ips = get_nodes_and_dst_ips(rsa_key_pair)  # [1st node, 2nd node, 3rd node, dst_ip]
    print("got ip's succesfully:", route_ips)
    
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

def get_nodes_and_dst_ips(rsa_obj : crypto.Rsa) -> list:
    """get router_ip and ip_of_destination from arguments and return nodes route and dst

    Returns:
        list(string): Received ip's from router
    """
    sock_with_router = connect_to_server(sys.argv[const.ROUTER_IP_IDX], const.ROUTER_PORT)
    get_route_message = const.CODE_ROUTE + str(len(rsa_obj.pem_public_key)).zfill(const.MESSAGE_SIZE_LEN) + rsa_obj.pem_public_key.decode()
    sock_with_router.sendall(get_route_message.encode())

    size = int(sock_with_router.recv(const.MESSAGE_SIZE_LEN).decode())

    response = sock_with_router.recv(size)
    sock_with_router.close()

    list_of_ips = rsa_obj.decrypt(response).decode().split("&")
    list_of_ips.append(sys.argv[const.DESTINATION__IP_IDX])
    return list_of_ips
 #  TODO: Change node that it will be able to connect to the tor-network using the router
 #  TODO: Add router option to return encrypted ip's delimetered with &

if __name__ == "__main__":
    main()
