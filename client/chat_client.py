from tor.client import TorClient
from tor.crypto import Rsa, Aes

import sys
import json

COOKIE_SIZE = 15
REQ_CODE_SIZE = 2
CODE_AUTH                 = b"00"
CODE_UPDATE               = b"01"
CODE_LOGIN                = b"02"
CODE_REGISTER             = b"03"
CODE_LOGOUT               = b"04"
CODE_CREATE_CHAT_ROOM     = b"05"
CODE_DELETE_CHAT_ROOM     = b"06"
CODE_JOIN_CHAT_ROOM       = b"07"
CODE_KICK_FROM_CHAT_ROOM  = b"08"
CODE_BAN_FROM_CHAT_ROOM   = b"09"
CODE_UNBAN_FROM_CHAT_ROOM = b"10"
CODE_SEND_MESSAGE         = b"11"
CODE_LOAD_MESSAGES        = b"12"
CODE_GET_ROOMS            = b"13"
CODE_IS_USER_IN_ROOM      = b"14"
CODE_CANCEL_UPDATE        = b"15"
CODE_ERR                  = b"99"
STATUS_SUCCESS = 1
STATUS_FAILED  = 0

KEYS_FILES_NAME = 'keys.pem'

class ChatClient:
    def __init__(self):
        priv_key, pub_key = load_RSA_from_file(KEYS_FILES_NAME)
        rsa_obj = Rsa(pub_key, priv_key)

        self._client = TorClient(rsa_obj,
                                 sys.argv[1], sys.argv[2])

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

    def logout(self) -> dict:
        return self._send_req(CODE_LOGOUT, None)
    
    def create_room(self, room_name, password) -> dict:
        req = { 'roomName' : room_name, 'password' : password}
        return self._send_req(CODE_CREATE_CHAT_ROOM, req)
    
    def delete_room(self, room_name, password) -> dict:
        req = { 'roomName' : room_name, 'password' : password}
        return self._send_req(CODE_DELETE_CHAT_ROOM, req)
    
    def join_room(self, room_name, password) -> dict:
        req = { 'roomName' : room_name, 'password' : password }
        return self._send_req(CODE_JOIN_CHAT_ROOM, req)

    def kick_user(self, room_name, username) -> dict:
        req = { 'roomName' : room_name, 'username' : username }
        return self._send_req(CODE_KICK_FROM_CHAT_ROOM, req)

    def ban_user(self, room_name, username) -> dict:
        req = { 'roomName' : room_name, 'username' : username }
        return self._send_req(CODE_BAN_FROM_CHAT_ROOM, req)

    def unban_user(self, room_name, username) -> dict:
        req = { 'roomName' : room_name, 'username' : username }
        return self._send_req(CODE_UNBAN_FROM_CHAT_ROOM, req)
    
    def send_message(self, room_name, content) -> dict:
        req = { 'roomName' : room_name, 'content': content }
        return self._send_req(CODE_SEND_MESSAGE, req)
    
    def load_messages(self, room_name, amount, offset) -> dict:
        req = { 'roomName' : room_name, 'amount' : amount, 'offset' : offset }
        return self._send_req(CODE_LOAD_MESSAGES, req)

    def get_rooms(self) -> dict:
        return self._send_req(CODE_GET_ROOMS, None)
    
    def is_user_in_room(self, room_name) -> dict:
        req = {'roomName' : room_name}
        return self._send_req(CODE_IS_USER_IN_ROOM, req)
    
    def get_update(self, room_name) -> dict:
        req = { 'roomName' : room_name }
        return self._send_req(CODE_UPDATE, req)

    def get_updates(self, room_name):
        while True:
            req = { 'roomName' : room_name }
            self._send_req(CODE_UPDATE, req)
    
    def cancel_update(self):
        return self._send_req(CODE_CANCEL_UPDATE, None)

            
#TODO: to make it faster remove in master branch
def load_RSA_from_file(path_to_keys):
    with open(path_to_keys, 'rb') as in_file:
        all_data = in_file.read().split(b"\n\n")
        return all_data[0], all_data[1] # 0 is private key 1 is public key
        

def write_RSA_to_file(path_to_keys : str, keys : Rsa):
    with open(path_to_keys, 'wb') as out_file:
        out_file.write(keys.pem_private_key)
        out_file.write(b'\n\n')
        out_file.write(keys.pem_public_key)
