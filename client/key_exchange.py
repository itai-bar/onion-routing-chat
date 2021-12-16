import crypto
import socket
import const
from serialize import serialize_tor_message

def key_exchange(ip_path, sock_with_first_node : socket.socket, rsa_obj):
    aes_keys = []  # list of aes.Aes keys
    for idx, ip in enumerate(ip_path): # entering this after exchanged keys with first node
        if idx > const.ST_NODE_IP_IDX:
            message = serialize_tor_message(rsa_obj.pem_public_key.decode(), ip_path[1:idx+1], False)
        else:
            message = serialize_tor_message(rsa_obj.pem_public_key.decode(), None, False)

        # encrypted_message = crypto.encrypt_by_order(message, aes_keys)
        encrypted_message = message
        sock_with_first_node.sendall(encrypted_message.encode())

        # reading data len
        size = int(sock_with_first_node.recv(const.MESSAGE_SIZE_LEN).decode())

        response = sock_with_first_node.recv(size)
        response = crypto.decrypt_by_order(response, aes_keys)
        aes_keys.append(crypto.Aes(rsa_obj.decrypt(response)))  # appending to aes_keys the aes key of curr iteration node

    return aes_keys
