import const

def serialize_tor_message(message, route_ips):
    """ Function creates protocoled message
    data transfering message:
        1  Byte     (message-code)
        15 Bytes    (padded 2nd node ip)
        15 Bytes    (padded 3rd node ip)
        15 Bytes    (padded dst ip)
        2 Bytes     (padded data size[max is 65535])
        data size   (data)

    Args:
        message (string): message to send
        nodes_ips (list(string)): ip's of 2nd node, 3rd node and destination ip.

    Returns:
        string: message suited to protocol
    """

    message += '\0'
    result = const.TRANSFER_MESSAGE_CODE + const.TRANSFER_MESSAGE_CODE.join(pad_ips(route_ips)) # Each node remove exact amount of bytes because padding
    result += str(len(message)).zfill(5)  # fill with zeros so the dst be able to read it with no problems
    result += message
    return result

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