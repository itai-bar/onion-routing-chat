from tor.client import TorClient
from tor.crypto import Rsa

import sys


# *** this is just a tester for the real client ***

def serialize_msg(code, msg):
    return code + str(len(msg)).zfill(5) + msg

def deserialize_msg(msg):
    return msg[4:]

def auth(client : TorClient):
    msg = serialize_msg('00', client._rsa.pem_public_key.decode())
    print(f'sending {msg}')

    resp = client.send(msg)
    print(f'got: {client._rsa.decrypt(deserialize_msg(resp))}')


if __name__ == '__main__':
    client = TorClient(Rsa(), sys.argv[1], sys.argv[2])
    auth(client)

