import crypto
import socket
import const
from serialize import serialize_tor_message

def key_exchange(ip_path : list, sock_for_exchanging : socket.socket, rsa_obj : crypto.Rsa) -> list:
    """Function exchange aes keys using the given rsa object and returns the exchanged aes-keys as ordered(by ip's) list

    Args:
        ip_path (list): ip's for key exchange
        sock_for_exchanging (socket.socket): socket with first network component
        rsa_obj (crypto.Rsa): rsa private and public key as crypto.Rsa object

    Returns:
        list: aes-keys(ordered by given ip's order)
    """
    aes_keys = []  # list of aes.Aes keys
    for idx, ip in enumerate(ip_path): # entering this after exchanged keys with first node
        message = serialize_tor_message(rsa_obj.pem_public_key.decode(), ip_path[1:idx+1], False, aes_keys)

        sock_for_exchanging.sendall(message)

        # reading data len
        size = int(sock_for_exchanging.recv(const.MESSAGE_SIZE_LEN).decode())

        response = sock_for_exchanging.recv(size)
        response = crypto.decrypt_by_order(response, aes_keys)
        aes_keys.append(crypto.Aes(rsa_obj.decrypt(response)))  # appending to aes_keys the aes key of curr iteration node

    return aes_keys
