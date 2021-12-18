from os import curdir
import const
import crypto

def serialize_tor_message(message : str, route_ips : list, close_socket : bool, aes_keys : list) -> bytes:
    """ Function creates protocoled message
    data transfering message:
        5  Bytes    (rest data-size)    not encrypted
        1  Byte     (close_socket)      encrypted by node1
        15 Bytes    (padded 2nd node ip)encrypted by node1
        1  Byte     (close_socket)      encrypted by node2,node1
        15 Bytes    (padded 3rd node ip)encrypted by node2,node1
        1  Bytes    (close_socket)      encrypted by node3,node2,node1
        15 Bytes    (padded dst ip)     encrypted by node3,node2,node1
        data size   (data)              encrypted by node3,node2,node1

    Args:
        message (string): message to send
        nodes_ips (list(string)): ip's of 2nd node, 3rd node and destination ip. depending on part of communication

    Returns:
        bytes: encrypted message suited to protocol
    """
    result = message.encode()
    flag   = (const.CLOSE_SOCKET_FLAG if close_socket else const.KEEP_SOCKET_FLAG).encode()
    
    for iteration in range(len(aes_keys)):  # encrypting by layers
        result = flag + pad_ip(route_ips[len(route_ips)-iteration-1]).encode() + result

        curr_layer_key = aes_keys[-1-iteration]
        if isinstance(curr_layer_key, crypto.Aes):
            result = curr_layer_key.encrypt(result)
    
    result = str(len(result)).zfill(5).encode() + result
    return result


def pad_ip(ip : str) -> str:
    """This function get not padded ip, for example 1.2.3.4 and return ip padded 00000001.2.3.4

    Args:
        ip (str): Ip to pad, fill with leading zeros to len 15

    Returns:
        str: Padded ip
    """

    return ip.zfill(15)