from cgi import test
import string
from tor.client import TorClient
from tor.crypto import Rsa, Aes

import json
import sys

COOKIE_SIZE = 15
REQ_CODE_SIZE = 2
CODE_AUTH             	   = b"00"
CODE_UPDATE           	   = b"01"
CODE_LOGIN            	   = b"02"
CODE_REGISTER         	   = b"03"
CODE_LOGOUT           	   = b"04"
CODE_CREATE_CHAT_ROOM 	   = b"05"
CODE_DELETE_CHAT_ROOM 	   = b"06"
CODE_JOIN_CHAT_ROOM   	   = b"07"
CODE_KICK_FROM_CHAT_ROOM   = b"08"
CODE_BAN_FROM_CHAT_ROOM	   = b"09"
CODE_ERR              	   = b"11"
STATUS_SUCCESS = 1
STATUS_FAILED  = 0

KEYS_FILES_NAME = 'keys.pem'

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
        resp_json = json.loads(self._aes.decrypt(self._client.send(req)).decode())
        print(resp_json)
        return resp_json
        
    def register(self, username, password) -> dict:
        req = { 'username' : username, 'password' : password }
        return self._send_req(CODE_REGISTER, req)
        
    def login(self, username, password) -> dict:
        req = { 'username' : username, 'password' : password }
        return self._send_req(CODE_LOGIN, req)
    
    def create_room(self, room_name, password) -> dict:
        req = { 'name' : room_name, 'password' : password}
        return self._send_req(CODE_CREATE_CHAT_ROOM, req)
    
    def delete_room(self, room_name, password) -> dict:
        req = { 'name' : room_name, 'password' : password}
        return self._send_req(CODE_DELETE_CHAT_ROOM, req)
    
    def join_room(self, room_name, password) -> dict:
        req = { 'name' : room_name, 'password' : password }
        return self._send_req(CODE_JOIN_CHAT_ROOM, req)

    def kick_user(self, room_name, username) -> dict:
        req = { 'name' : room_name, 'username' : username }
        return self._send_req(CODE_KICK_FROM_CHAT_ROOM, req)

    def ban_user(self, room_name, username) -> dict:
        req = { 'name' : room_name, 'username' : username }
        return self._send_req(CODE_BAN_FROM_CHAT_ROOM, req)


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
    priv_key, pub_key = load_RSA_from_file(KEYS_FILES_NAME)
    rsa_obj = Rsa(pub_key, priv_key)
    tester_tal = Tester(TorClient(rsa_obj, sys.argv[1], sys.argv[2]))
    tester_itai = Tester(TorClient(rsa_obj, sys.argv[1], sys.argv[2]))

    tester_tal.auth()
    tester_itai.auth()

    assert tester_tal.register('tal', 'pass')['status'] == STATUS_SUCCESS
    assert tester_tal.login('tal', 'pass')['status'] == STATUS_SUCCESS
    assert tester_tal.create_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS
    assert tester_tal.create_room('my_room', 'room_pass')['status'] == STATUS_FAILED
    assert tester_tal.delete_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS
    assert tester_tal.create_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS

    assert tester_itai.register('itai', 'long_pass')['status'] == STATUS_SUCCESS
    assert tester_itai.login('itai', 'forgot_pass')['status'] == STATUS_FAILED
    assert tester_itai.login('itai', 'long_pass')['status'] == STATUS_SUCCESS
    assert tester_itai.join_room('my_room', 'wrong_pass')['status'] == STATUS_FAILED
    assert tester_itai.join_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS
    assert tester_itai.delete_room('my_room', 'room_pass')['status'] == STATUS_FAILED

    assert tester_tal.kick_user('my_room', 'itai')['status'] == STATUS_SUCCESS
    assert tester_itai.join_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS

    assert tester_tal.ban_user('my_room', 'itai')['status'] == STATUS_SUCCESS
    assert tester_itai.join_room('my_room', 'room_pass')['status'] == STATUS_FAILED
    