import string
from tor.client import TorClient
from tor.crypto import Rsa, Aes

import json
import sys

COOKIE_SIZE = 15
REQ_CODE_SIZE = 2
CODE_AUTH     = b"00"
CODE_UPDATE   = b"01"
CODE_LOGIN    = b"02"
CODE_REGISTER = b"03"
CODE_LOGOUT   = b"04"
CODE_MSG      = b"05"

# *** this is just a tester for the server ***

class Tester:
    def __init__(self, client: TorClient):
        self._client = client
        
    def auth(self):
        msg = CODE_AUTH + self._client._rsa.pem_public_key

        resp = self._client.send(msg)
        decrypted = self._client._rsa.decrypt(resp)

        try:
            self._cookie, self._aes =  decrypted[:COOKIE_SIZE], Aes(decrypted[COOKIE_SIZE:])
        except IndexError as e:
            print('Error: auth msg is invalid:', e)
            sys.exit(1)
    
    def _send_req(self, code : bytes, data : dict):
        req = code + self._cookie + self._aes.encrypt(json.dumps(data).encode())
        return self._aes.decrypt(self._client.send(req)).decode()
        
    def register(self, username, password):
        req = { 'username' : username, 'password' : password }
        resp = self._send_req(CODE_REGISTER, req)
        print(f'{req} : {resp}')
        
    def login(self, username, password):
        req = { 'username' : username, 'password' : password }
        resp = self._send_req(CODE_LOGIN, req)
        print(f'{req} : {resp}')

def load_RSA_from_file(path_to_keys):
    with open(path_to_keys, 'rb') as in_file:
        all_data = in_file.read().split(b"\n\n")
        return all_data[0], all_data[1] # 0 is private key 1 is public key
        

def write_RSA_to_file(path_to_keys : str, keys : Rsa):
    with open(path_to_keys, 'wb') as out_file:
        out_file.write(keys.pem_private_key)
        out_file.write(b'\n\n')
        out_file.write(keys.pem_public_key)

if __name__ == '__main__':
    keys_file_name = 'keys.pem'
    priv_key, pub_key = load_RSA_from_file(keys_file_name)
    rsa_obj = Rsa(pub_key, priv_key)
    tester = Tester(TorClient(rsa_obj, sys.argv[1], sys.argv[2]))

    tester.auth()
    tester.register('itai', 'pass')

    tester.login('itai', 'pass')
    tester.login('ita', 'pass')
    tester.login('itai', 'wrongpass')