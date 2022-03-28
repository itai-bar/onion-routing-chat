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
from kivy.uix.recycleview.views import RecycleDataViewBehavior
from kivy.uix.boxlayout import BoxLayout
from kivy.uix.gridlayout import GridLayout
from kivy.properties import StringProperty, NumericProperty
import atexit
from functools import partial

import numpy

from chat_client import STATE_OFFLINE, STATE_ONLINE, STATE_BANNED, ChatClient, STATUS_FAILED
CHATS_IN_LINE = 50
DISTANCE_BETWEEN_LINES = 20
ARGUMENTS_IDX = 1
FUNC_IDX = 0
RED_COLOR = [1, 255, 255]
GREEN_COLOR = [255, 1, 255]
BLUE_COLOR = [255, 255, 1]
ADMIN_COLOR = [0, 63, 52]

# parsing time
from dateutil import parser
import datetime as dt

import threading

client = ChatClient()

def message_to_str(msg: dict) -> tuple:
    # parsing the unix time sent with the message to a readable date
    time    = parser.parse(msg['time']).strftime("%d.%m.%y %H:%M")
    sender  = msg['sender']
    content, amount_of_lines = set_new_lines(msg['content'], CHATS_IN_LINE)
    sender_color = GREEN_COLOR if sender == client.username else BLUE_COLOR

    return time, sender + ':', content, sender_color, DISTANCE_BETWEEN_LINES*amount_of_lines

def set_new_lines(content, line_size):
    return "-\n".join([content[start_of_line:start_of_line+line_size] for start_of_line in range(0, len(content), line_size)]), int(numpy.ceil(len(content)/line_size))

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

class EmptyWindow(Screen):
    def __init__(self, wm, **kw):
        self.wm = wm
        super().__init__(**kw)

class WelcomeWindow(Screen):
    def __init__(self, wm, **kw):
        self.wm = wm
        super().__init__(**kw)

    def on_enter(self, *args):
        client.load_RSA_keys()
        client.auth()
        self.wm.current = 'login'

class LoginWindow(Screen):
    username = ObjectProperty(None)
    password = ObjectProperty(None)

    def __init__(self, wm, **kw):
        self.wm = wm
        super().__init__(**kw)
    
    def btn_login(self):
        resp = client.login(self.username.text, self.password.text)
        username_tmp = self.username.text
        self.reset()

        if resp['status'] == STATUS_FAILED:
            popup('login error', resp['info'])
        else:
            client.username = username_tmp
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

class UserLabel(RecycleDataViewBehavior, BoxLayout):
    username_text = StringProperty()
    red_color = NumericProperty()
    green_color = NumericProperty()
    blue_color = NumericProperty()

class RoomMembersPopup(GridLayout):
    users = ListProperty()
    def __init__(self, wm, PopupInstance : Popup, roomName : str, **kwargs):
        super().__init__(**kwargs)
        self._roomName = roomName
        self.wm = wm
        self.PopupInstance = PopupInstance
        self.get_and_add_room_members()

    def get_and_add_room_members(self, *args):
        self.ids.banListOrMembersBttn.unbind(on_press=self.get_and_add_room_members)
        self.ids.banListOrMembersBttn.bind(on_press=self.get_ban_list)
        self.users = [] # clear last widges that was shown
        self.PopupInstance.title="Room members"
        self.ids.banListOrMembersBttn.text="Ban list"

        resp = client.get_room_members(self._roomName)
        if resp['status'] == STATUS_FAILED:
            popup('open room members list error', resp['info'])
        else:
            self._add_members_to_list(resp['onlineMembers'], STATE_ONLINE)
            self._add_members_to_list(resp['offlineMembers'], STATE_OFFLINE)
            self._set_admin_in_list(resp['adminName'])
            self.PopupInstance.title += " ---> Online members: " + f"{str(len(resp['onlineMembers']))}/{str((len(resp['offlineMembers']))+len(resp['onlineMembers']))}"

    def _set_admin_in_list(self, admin_name):
        for member in self.users:
            if member['username_text'].startswith(admin_name) and member['username_text'][len(admin_name):] in [" - Online", " - Offline"]: # check if admin and not just stars with name of admin
                member['username_text'] = "Admin:" + member['username_text']
                member['red_color'], member['green_color'], member['blue_color'] = ADMIN_COLOR
                break # only one admin, no need to continue searching

    def _add_members_to_list(self, members, state):
        if members == None:
            return
        for member in members:
            name_and_state = member + " - " + ("Online" if state == STATE_ONLINE else ("Offline" if state == STATE_OFFLINE else "Banned"))
            color = (GREEN_COLOR if state == STATE_ONLINE else ([1, 1, 1] if state == STATE_OFFLINE else RED_COLOR))
            print(name_and_state)
            self.users.append({'username_text'   : name_and_state,
                               'red_color': color[0],
                               'green_color': color[1],
                               'blue_color': color[2]})
    
    def get_ban_list(self, *args):
        self.ids.banListOrMembersBttn.unbind(on_press=self.get_ban_list)
        self.ids.banListOrMembersBttn.bind(on_press=self.get_and_add_room_members)
        self.users = [] # clear last widges that was shown
        self.PopupInstance.title="Ban list"
        self.ids.banListOrMembersBttn.text="Room members"

        resp = client.get_banned_members(self._roomName)
        if resp['status'] == STATUS_FAILED:
            popup('open ban list error', resp['info'])
        else:
            self._add_members_to_list(resp['bannedMembers'], STATE_BANNED)

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
    

class MessageLabel(RecycleDataViewBehavior, BoxLayout):
    time_text    = StringProperty()
    sender_text  = StringProperty()
    sender_text_len = NumericProperty()
    content_text = StringProperty()
    red_color = NumericProperty()
    green_color = NumericProperty()
    blue_color = NumericProperty()
    height_by_lines = NumericProperty()

    
class ChatWindow(Screen):
    messages = ListProperty()
    room_name_text = StringProperty()
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
                time, sender, content, color, height_by_lines = message_to_str(msg) 
                self.messages.append({'time_text'   : time,
                                      'sender_text' : sender,
                                      'content_text': content,
                                      'red_color': color[0],
                                      'green_color': color[1],
                                      'blue_color': color[2],
                                      'height_by_lines': height_by_lines})

    def on_enter(self, *args):
        # showing the name of the room to the user
        self.room_name_text = self.manager.statedata.current_room
        # allowing the update thread to loop
        ChatWindow.getting_updates = True
        new_msgs = client.load_messages(self.manager.statedata.current_room, 50, 0)

        # reverse messages for better user experience
        for msg in (new_msgs['messages'] or [])[::-1]: 
            time, sender, content, color, height_by_lines = message_to_str(msg) 
            self.messages.append({'time_text'   : time,
                                  'sender_text' : sender,
                                  'sender_text_len': len(sender),
                                  'content_text': content,
                                  'red_color': color[0],
                                  'green_color': color[1],
                                  'blue_color': color[2],
                                  'height_by_lines': height_by_lines})

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
        self.ids.message.focus = True
    
    def open_room_members_list(self):
        roomMembersPopup = Popup(size_hint=(None,None), size=(400, 300))
        roomMembersPopup.content = RoomMembersPopup(self.wm, roomMembersPopup, self.manager.statedata.current_room)
        roomMembersPopup.open()
    
    def admin_command(self, cmd: str):
        req, args = cmd.split()[0], cmd.split()[1:]
        req       = req[1:]
        
        cmd_map = { 'kick'     :  [client.kick_user, '[username]'],
                    'ban'      :  [client.ban_user, '[username]'],
                    'unban'    :  [client.unban_user, '[username]'],
                    'delete'   :  [client.delete_room, '[room password]']
                }

        try:
            if req == 'help':
                self.print_commands(cmd_map)
                return
            else:
                resp = cmd_map[req][FUNC_IDX](self.manager.statedata.current_room, args[0])
            if resp['status'] == STATUS_FAILED:
                popup('command failed', resp['info'])
            else:
                popup('command succeeded', resp['info'])
        except KeyError:
            popup('command error', 'command does not exists!')
        except IndexError:
            popup('command error', f'Invalid arguments\nusage:/{req} {cmd_map[req][ARGUMENTS_IDX]}')

    def print_commands(self, cmd_map:dict):
        commands_with_usage = []
        for command, value in cmd_map.items():
            commands_with_usage.append(f"{command} {value[ARGUMENTS_IDX]}")
        popup('commands', '\n'.join(commands_with_usage))