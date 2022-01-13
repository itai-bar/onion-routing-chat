import socket

from . import serialize, const, crypto, validation, key_exchange as ke

class TorClient:
    def __init__(self, rsa : crypto.Rsa, router_ip : str, dst_ip : str):
        self._rsa = rsa
        self._router_ip = router_ip
        self._dst_ip = dst_ip
        self._cookie = ""
        self._server_aes_key = ""

    def send(self, msg):
        """Sends a message using the tor protocol

        Args:
            msg (str): message for the final server
        Returns:
            resp (str)
        """

        route = self.get_nodes()  # [1st node, 2nd node, 3rd node, dst_ip]
        print("got ip's succesfully:", route)

        with connect_to_server(route[const.ST_NODE_IP_IDX], 8989) as sock:
            aes_keys = ke.key_exchange(route[:-1], sock, self._rsa)

            tor_msg = serialize.serialize_tor_message(msg, route[1:], True, aes_keys)
            print(tor_msg)

            sock.sendall(tor_msg)

            size = int(sock.recv(const.MESSAGE_SIZE_LEN).decode()) # reading plaintext size
            resp = sock.recv(size)
    
            resp = crypto.decrypt_by_order(resp, aes_keys)

            return resp

    def get_nodes(self):
        """get router_ip and ip_of_destination from arguments and return nodes route and dst

        Returns:
            list(string): Received ip's from router
        """
        sock_with_router = connect_to_server(self._router_ip, const.ROUTER_PORT)
        get_route_message = const.CODE_ROUTE + str(len(self._rsa.pem_public_key)).zfill(const.MESSAGE_SIZE_LEN) + self._rsa.pem_public_key.decode()
        sock_with_router.sendall(get_route_message.encode())

        size = int(sock_with_router.recv(const.MESSAGE_SIZE_LEN).decode())

        response = sock_with_router.recv(size)
        sock_with_router.close()

        list_of_ips = self._rsa.decrypt(response).decode().split("&")
        list_of_ips.append(self._dst_ip)
        return list_of_ips
    
    def deserialize_auth_msg(self, auth_msg : bytes):
        try:
            self._cookie = auth_msg[:const.COOKIE_SIZE]
            self._server_aes_key = auth_msg[const.COOKIE_SIZE:]
        except IndexError as e:
            print("Error: auth msg is invalid:", e)

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

