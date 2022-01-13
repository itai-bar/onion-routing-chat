from tor.client import TorClient
from tor.crypto import Rsa

import sys


# *** this is just a tester for the real client ***

def serialize_msg(code, msg):
    return code + str(len(msg)).zfill(5) + msg


def auth(client : TorClient):
    msg = serialize_msg('00', client._rsa.pem_public_key.decode())
    print(f'sending {msg}')

    resp = client.send(msg)
    decrypted = client._rsa.decrypt(resp)

    client.deserialize_auth_msg(decrypted)


if __name__ == '__main__':
    client = TorClient(Rsa(), sys.argv[1], sys.argv[2])
    auth(client)
