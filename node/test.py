import socket

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
    print('connected to ', server_address)
    return sock

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

def serializeDataTransferingMessage(message, route_ips):
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


sock = connect_to_server('127.0.0.1', 7777)
ips = ['123.123.123.123', '43.23.0.1', '1.1.1.1']
ips = pad_ips(ips)
msg = serializeDataTransferingMessage('hello', ips)
sock.sendall(msg.encode())