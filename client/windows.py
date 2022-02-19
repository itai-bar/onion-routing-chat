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
            self.wm.current = 'rooms'
            
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

class PasswordPopup(GridLayout):
    def __init__(self, roomName, **kwargs):
        super().__init__(**kwargs)
        self._roomName = roomName
        
    def btn_enter_room(self):
        resp = client.join_room(self._roomName, self.ids.roomPassword.text)
        if resp['status'] == STATUS_FAILED:
            popup('enter room error', resp['info'])
        else:
            #self.wm.current = 'login'
            #GOTO: go to room
            pass


class RoomsWindow(Screen):
    def __init__(self, wm, **kw):
        self.wm = wm
        super().__init__(**kw)
        
        for room in range(1,4):
            roomName = "fake" + str(room)
            show = PasswordPopup(roomName)
            passwordPopup = Popup(title="Enter " + roomName + "'s password", content=show, size_hint=(0.3,0.3), size=(200, 200))
            roomBtn = Button(text=roomName, size_hint_y=None,height=100, on_press=lambda a:passwordPopup.open())
            self.ids.roomsNames.add_widget(roomBtn)

    def on_enter(self, *args):
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
                    passwordPopup = Popup(title="Enter " + room + "'s password", content=show, size_hint=(0.3,0.3), size=(200, 200))
                    roomBtn = Button(text=room, size_hint_y=None,height=100, on_press=lambda a:passwordPopup.open())
                    self.ids.roomsNames.add_widget(roomBtn)
            
