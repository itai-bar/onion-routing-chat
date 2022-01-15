from cgi import test
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
        return self._aes.decrypt(self._client.send(req))
        

    def register(self, username, password):
        req = { 'username' : username, 'password' : password }
        print(self._send_req(CODE_REGISTER, req)) 
        

if __name__ == '__main__':
    tester = Tester(TorClient(Rsa(), sys.argv[1], sys.argv[2]))
    tester.auth()
    tester.register('itai', 'very secret pass')