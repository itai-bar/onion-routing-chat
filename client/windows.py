# login window 
# signup window
# main menu window
# create room window
# rooms window with a room join option
# chat window
# room managment for admin only

from concurrent.futures import thread
from multiprocessing import managers
from kivy.uix.screenmanager import Screen, ScreenManager
from kivy.uix.floatlayout import FloatLayout
from kivy.properties import ObjectProperty
from kivy.properties import ListProperty
from kivy.uix.popup import Popup
from kivy.uix.label import Label
from kivy.uix.button import Button
from kivy.uix.scrollview import ScrollView
from kivy.uix.recycleview import RecycleView
from kivy.uix.textinput import TextInput
from kivy.uix.gridlayout import GridLayout
import atexit
from functools import partial

from chat_client import STATE_OFFLINE, STATE_ONLINE, ChatClient, STATUS_FAILED

# parsing time
from dateutil import parser
import datetime as dt

import threading

client = ChatClient()
client.auth()

def message_to_str(msg: dict) -> str:
    # parsing the unix time sent with the message to a readable date
    t = parser.parse(msg['time']).strftime("%d.%m.%y %H:%M")
    return f"{t} | {msg['sender']} - {msg['content']}"

def exit_handler():
    ChatWindow.getting_updates = False # stop the update request loop
    client.cancel_update() # makes the server stop holding the requeqst
    client.logout()
    
# calling the exit handler at exit
atexit.register(exit_handler) 

def popup(title, text):
    pop = Popup(title=title,
                  content=Label(text=text),
                  size_hint=(None, None), size=(400, 400))
    pop.open()

class LoginWindow(Screen):
    username = ObjectProperty(None)
    password = ObjectProperty(None)

    def __init__(self, wm, **kw):
        self.wm = wm
        super().__init__(**kw)
    
    def btn_login(self):
        resp = client.login(self.username.text, self.password.text)
        self.reset()

        if resp['status'] == STATUS_FAILED:
            popup('login error', resp['info'])
        else:
            self.wm.current = 'main'
            
    def btn_goto_signup(self):
        self.reset()
        self.wm.current = 'signup'

    def reset(self):
        self.username.text = ''
        self.password.text = ''

class SignupWindow(Screen):
    username = ObjectProperty(None)
    password = ObjectProperty(None)

    def __init__(self, wm, **kw):
        self.wm = wm
        super().__init__(**kw)

    def btn_signup(self):
        resp = client.register(self.username.text, self.password.text)
        self.reset()

        if resp['status'] == STATUS_FAILED:
            popup('signup error', resp['info'])
        else:
            self.wm.current = 'login'

    def btn_goto_login(self):
        self.wm.current = 'login'

    def reset(self):
        self.username.text = ''
        self.password.text = ''

class CreateRoomPopup(GridLayout):
    def __init__(self, **kwargs):
        super().__init__(**kwargs)

    def btn_create_room(self):
        resp = client.create_room(self.ids.roomName.text, self.ids.roomPassword.text)
        self.reset()
        
        if resp['status'] == STATUS_FAILED:
            popup('create room error', resp['info'])

    def reset(self):
        self.ids.roomPassword.text = ''
        self.ids.roomName.text = ''
    
class PasswordPopup(GridLayout):
    def __init__(self, wm, PopupInstance : Popup, roomName : str, **kwargs):
        super().__init__(**kwargs)
        self._roomName = roomName
        self.wm = wm
        self.PopupInstance = PopupInstance
        
    def btn_enter_room(self):
        resp = client.join_room(self._roomName, self.ids.roomPassword.text)
        self.ids.roomPassword.text = ""

        if resp['status'] == STATUS_FAILED:
            popup('enter room error', resp['info'])
        else:
            self.PopupInstance.dismiss()
            self.wm.current = 'chat'

class RoomMembersPopup(GridLayout):
    def __init__(self, wm, PopupInstance : Popup, roomName : str, **kwargs):
        super().__init__(**kwargs)
        self._roomName = roomName
        self.wm = wm
        self.PopupInstance = PopupInstance
        self.on_enter()

    def on_enter(self):
        resp = client.get_room_members(self._roomName)
        if resp['status'] == STATUS_FAILED:
            popup('open room members list error', resp['info'])
        else:
            self._add_admin_to_list(resp['adminName'])
            self._add_members_to_list(resp['onlineMembers'], STATE_ONLINE)
            self._add_members_to_list(resp['offlineMembers'], STATE_OFFLINE)

    def _add_admin_to_list(self, admin_name):
        print("admin is:", admin_name)
        pass #TODO:add admin so it will be recognized as admin

    def _add_members_to_list(self, members, state):
        for member in members:            
            name_and_state = member + " - " + "Online" if state == STATE_ONLINE else "Offline"
            print(name_and_state)
            self.ids.roomMembers.add_widget(Label(text=name_and_state))

    def close_room(self):
        print("ask for close room")  # included notice all online members that server doesn't exists anymore. (get them out?!)
    
    def get_ban_list(self):
        print("ask for users in ban")

class MainWindow(Screen):
    def __init__(self, wm, **kw):
        self.wm = wm
        super().__init__(**kw)
    
    def btn_logout(self):
        client.logout()
        self.wm.current = 'login' 
    
    def btn_create_room(self):
        create_room_popup = Popup(title='Create room', 
                                content=CreateRoomPopup(), 
                                size_hint=(0.3,0.3), size=(200, 200))
        create_room_popup.open()
        
    def btn_goto_rooms(self):
        self.wm.current = 'rooms'


class RoomsWindow(Screen):
    def __init__(self, wm, **kw):
        self.wm = wm
        super().__init__(**kw)

    def on_enter(self, *args):
        self.clean_rooms()
        self.load_rooms()

    def load_rooms(self):
        resp = client.get_rooms()

        if resp['status'] == STATUS_FAILED:
            popup('rooms error', resp['info'])
        else:
            for room in (resp['rooms'] or []):
                passwordPopup = Popup(title=f"Enter {room}'s password", size_hint=(0.3,0.3), size=(200, 200))
                passwordPopup.content = PasswordPopup(self.wm, passwordPopup, room)

                roomBtn = Button(text=room, 
                                size_hint_y=None,
                                height=100, 
                                on_press=partial(self.is_user_in_room, passwordPopup))

                # adding the button to the list
                self.ids.roomsNames.add_widget(roomBtn)
    
    def is_user_in_room(self, passwordPopup:Popup, *args):
        resp = client.is_user_in_room(passwordPopup.content._roomName)
        self.manager.statedata.current_room = passwordPopup.content._roomName

        if resp['info'] == 'user in room':
            self.wm.current = 'chat'
        else:
            passwordPopup.open()

    def clean_rooms(self):
        self.ids.roomsNames.clear_widgets()
    
    def go_to_main(self):
        self.wm.current = 'main'
    
class ChatWindow(Screen):
    #TODO: add room members button functionallity
    messages = ListProperty()
    getting_updates = False # used outside of this class
    
    def __init__(self, wm, **kw):
        self.wm = wm
        self.update_thread = None
        super().__init__(**kw)            

    def update_messages(self):
        # asking the server for updates about the chat
        while ChatWindow.getting_updates:
            msgs = client.get_update(self.manager.statedata.current_room)

            # reverse messages for better user experience
            for msg in (msgs['messages'] or [])[::-1]:  
                self.messages.append({'text' : message_to_str(msg)})

    def on_enter(self, *args):
        # allowing the update thread to loop
        ChatWindow.getting_updates = True
        new_msgs = client.load_messages(self.manager.statedata.current_room, 50, 0)

        # reverse messages for better user experience
        for msg in (new_msgs['messages'] or [])[::-1]: 
            self.messages.append({'text' : message_to_str(msg)})

        # starting the update thread
        self.update_thread = threading.Thread(target=self.update_messages, daemon=True)
        self.update_thread.start()

    def go_to_rooms(self):
        self.wm.current = 'rooms'

    def leave_room(self):
        client.leave_room(self.manager.statedata.current_room)
        self.wm.current = 'rooms'
    
    def on_leave(self, *args):
        # stoping update requests
        ChatWindow.getting_updates = False
        client.cancel_update(self.manager.statedata.current_room)

        # reseting chat properties 
        self.messages = []
        self.manager.statedata.current_room = ''

    def reset(self):
        self.ids.message.text = ''
        
    def send_message(self):
        msg = self.ids.message.text

        # /command args args
        # used for admin commands
        if msg.startswith('/'):
            self.admin_command(msg)
        else:
            client.send_message(self.manager.statedata.current_room, msg)

        self.reset()
    
    def open_room_members_list(self):
        print("opening room members list")  # TODO: check why it doesn't get open on press, and it open just when pressing 'leave room'
        roomMembersPopup = Popup(title=f"Room members", size_hint=(0.3,0.3), size=(200, 200))
        roomMembersPopup.content = RoomMembersPopup(self.wm, roomMembersPopup, self.manager.statedata.current_room)
        roomMembersPopup.open()
    
    def admin_command(self, cmd: str):
        req, args = cmd.split()[0], cmd.split()[1:]
        req       = req[1:]
        
        cmd_map = { 'kick'   :  client.kick_user,
                    'ban'    :  client.ban_user,
                    'unban'  :  client.unban_user
                }

        try:
            if len(args) != 1:
                popup('command', '1 arg only!')
            else:
                resp = cmd_map[req](self.manager.statedata.current_room, args[0])
                if resp['status'] == STATUS_FAILED:
                    popup('command failed', resp['info'])
                else:
                    popup('command succeeded', resp['info'])
        except KeyError:
            popup('command error', 'command does not exists!')