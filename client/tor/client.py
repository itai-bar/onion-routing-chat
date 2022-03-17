import socket

from . import serialize, const, crypto, validation, key_exchange as ke

class TorClient:
    def __init__(self, rsa : crypto.Rsa, router_ip : str, dst_ip : str):
        self._rsa       = rsa
        self._router_ip = router_ip
        self._dst_ip    = dst_ip

    def send(self, msg : bytes):
        """Sends a message using the tor protocol

        Args:
            msg (str): message for the final server
        Returns:
            resp (str)
        """
        route = self.get_nodes()  # [1st node, 2nd node, 3rd node, dst_ip]

        with connect_to_server(route[const.ST_NODE_IP_IDX], 8989) as sock:
            # getting the aes keys of all nodes
            aes_keys = ke.key_exchange(route[:-1], sock, self._rsa)

            # encrypting the message with the aes layers and sending
            tor_msg = serialize.serialize_tor_message(msg, route[1:], True, aes_keys)
            sock.sendall(tor_msg)

            # reading plaintext size of the response
            resp_size = int(sock.recv(const.MESSAGE_SIZE_LEN).decode()) 
            encrypted_resp = sock.recv(resp_size)

            # decrypting all aes layers
            return crypto.decrypt_by_order(encrypted_resp , aes_keys)

    def get_nodes(self):
        """get router_ip and ip_of_destination from arguments and return nodes route and dst

        Returns:
            list(string): Received ip's from router
        """
        sock_with_router  = connect_to_server(self._router_ip, const.ROUTER_PORT)

        # seralizing and sending the get route request
        get_route_message = const.CODE_ROUTE + str(len(self._rsa.pem_public_key)).zfill(const.MESSAGE_SIZE_LEN) + self._rsa.pem_public_key.decode()
        sock_with_router.sendall(get_route_message.encode())

        response_size = int(sock_with_router.recv(const.MESSAGE_SIZE_LEN).decode())

        response = sock_with_router.recv(response_size)
        sock_with_router.close()

        list_of_ips = self._rsa.decrypt(response).decode().split("&")
        list_of_ips.append(self._dst_ip)
        return list_of_ips
    

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

