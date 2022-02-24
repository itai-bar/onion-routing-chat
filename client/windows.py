# login window 
# signup window
# main menu window
# create room window
# rooms window with a room join option
# chat window
# room managment for admin only

from kivy.uix.screenmanager import Screen, ScreenManager
from kivy.uix.floatlayout import FloatLayout
from kivy.properties import ObjectProperty
from kivy.uix.popup import Popup
from kivy.uix.label import Label
from kivy.uix.button import Button
from kivy.uix.scrollview import ScrollView
from kivy.uix.recycleview import RecycleView
from kivy.factory import Factory
from kivy.uix.textinput import TextInput
from kivy.uix.gridlayout import GridLayout
import atexit
from functools import partial

from chat_client import ChatClient
from chat_client import STATUS_SUCCESS, STATUS_FAILED

client = ChatClient()
client.auth()

def exit_handler():
    client.logout()
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
    def __init__(self, roomName, **kwargs):
        super().__init__(**kwargs)
        self._roomName = roomName
        
    def btn_enter_room(self):
        resp = client.join_room(self._roomName, self.ids.roomPassword.text)
        self.ids.roomPassword.text = ""

        if resp['status'] == STATUS_FAILED:
            popup('enter room error', resp['info'])
        else:
            #self.wm.current = 'login'
            #TODO: go to room
            pass

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
        self.set_fake_rooms(7)
        self.load_rooms()

    def load_rooms(self):
        resp = client.get_rooms()
        if resp['status'] == STATUS_FAILED:
            popup('rooms error', resp['info'])
        else:
            rooms = resp['rooms']
            if rooms:
                for room in rooms:
                    show = PasswordPopup(room)
                    passwordPopup = Popup(title=f"Enter {room}'s password", content=show, size_hint=(0.3,0.3), size=(200, 200))
                    roomBtn = Button(text=room, size_hint_y=None,height=100, on_press=partial(self.is_user_in_room, passwordPopup))
                    self.ids.roomsNames.add_widget(roomBtn)
    
    def is_user_in_room(self, passwordPopup:Popup, *args):
        resp = client.is_user_in_room(passwordPopup.content._roomName)
        if resp['info'] == 'user in room':
            print("user already in room")
            pass #pass user to room
        else:
            passwordPopup.open()

    
    def clean_rooms(self):
        self.ids.roomsNames.clear_widgets()
    
    def go_to_main(self):
        self.wm.current = 'main'
    
    def set_fake_rooms(self, amount_of_fakes):
        for room in range(1, amount_of_fakes+1):
            room_name = "fake" + str(room)
            show = PasswordPopup(room_name)
            password_popup = Popup(title=f"Enter {room_name}'s password", content=show, size_hint=(0.3,0.3), size=(200, 200))
            roomBtn = Button(text=room_name, size_hint_y=None,height=100, on_press=password_popup.open)
            self.ids.roomsNames.add_widget(roomBtn)
