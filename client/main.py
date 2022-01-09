from tor.client import tor_message
from tor.crypto import Rsa

if __name__ == '__main__':
    rsa_key_pair = Rsa()  # creating Rsa class with random keypair for all sessions
    resp = tor_message("hello", rsa_key_pair)
    print(resp)

   