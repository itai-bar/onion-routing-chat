from cgi import test
from tor.client import TorClient
from tor.crypto import Rsa, Aes
from chat_client import ChatClient, STATUS_SUCCESS, STATUS_FAILED

import threading
import json
import sys
import time

if __name__ == '__main__':
    tester_tal = ChatClient()
    tester_itai = ChatClient()
    tester_dan = ChatClient()

    tester_tal.auth()
    tester_itai.auth()
    tester_dan.auth()

    assert tester_tal.register('tal', 'pass1')['status'] == STATUS_SUCCESS
    assert tester_tal.login('tal', 'pass1')['status'] == STATUS_SUCCESS

    assert tester_tal.logout()['status'] == STATUS_SUCCESS
    assert tester_tal.create_room('my_room', 'room_pass')['status'] == STATUS_FAILED 

    assert tester_tal.login('tal', 'pass1')['status'] == STATUS_SUCCESS
    assert tester_tal.create_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS
    assert tester_tal.create_room('my_room', 'room_pass')['status'] == STATUS_FAILED
    assert tester_tal.delete_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS
    assert tester_tal.create_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS

    assert tester_itai.register('itai', 'long_pass1')['status'] == STATUS_SUCCESS
    assert tester_itai.login('itai', 'forgot_pass')['status'] == STATUS_FAILED
    assert tester_itai.login('itai', 'long_pass1')['status'] == STATUS_SUCCESS
    assert tester_itai.join_room('my_room', 'wrong_pass')['status'] == STATUS_FAILED
    assert tester_itai.join_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS
    assert tester_itai.delete_room('my_room', 'room_pass')['status'] == STATUS_FAILED

    assert tester_tal.kick_user('my_room', 'itai')['status'] == STATUS_SUCCESS
    assert tester_itai.join_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS

    assert tester_tal.ban_user('my_room', 'itai')['status'] == STATUS_SUCCESS
    assert tester_tal.get_banned_members('my_room')['status'] == STATUS_SUCCESS
    assert tester_itai.join_room('my_room', 'room_pass')['status'] == STATUS_FAILED
    assert tester_itai.unban_user('my_room', 'itai')['status'] == STATUS_FAILED
    assert tester_itai.join_room('my_room', 'room_pass')['status'] == STATUS_FAILED
    assert tester_tal.unban_user('my_room', 'itai')['status'] == STATUS_SUCCESS
    assert tester_itai.join_room('my_room', 'room_pass')['status'] == STATUS_SUCCESS

    assert tester_dan.login('dandan', '123')['status'] == STATUS_FAILED
    assert tester_dan.register('dandan', 'heyhey')['status'] == STATUS_SUCCESS
    assert tester_dan.login('dandan', 'heyhey')['status'] == STATUS_SUCCESS
    assert tester_dan.join_room("my_room", "room_pass")['status'] == STATUS_SUCCESS
    assert tester_itai.create_room("itai_room", "room_strong_pass")['status'] == STATUS_SUCCESS
    assert tester_dan.join_room("my_room", "room_pass")['status'] == STATUS_FAILED
    assert tester_dan.join_room("itai_room", "room_strong_pass")['status'] == STATUS_SUCCESS

    

    t = threading.Thread(target=tester_itai.get_updates, args=('my_room', ))
    t.start()

    assert tester_dan.get_rooms()['status'] == STATUS_SUCCESS
    
    msg = 'hello this is a message!!!'
    for i in range(30):
        assert tester_tal.send_message('my_room', msg + f' {i}')['status'] == STATUS_SUCCESS

    offset = 0
    tester_dan.load_messages('my_room', 3, offset)
    offset += 3
    tester_dan.load_messages('my_room', 3, offset)
    
    tester_dan.logout()
    tester_itai.logout()
    tester_tal.logout()