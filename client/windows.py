# login window 
# signup window
# rooms window with a room join option
# chat window
# room managment for admin only

from kivy.uix.screenmanager import Screen, ScreenManager
from kivy.properties import ObjectProperty
from kivy.uix.popup import Popup
from kivy.uix.label import Label

from chat_client import ChatClient
from chat_client import STATUS_SUCCESS, STATUS_FAILED

client = ChatClient()
client.auth()

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
        if resp['status'] == STATUS_FAILED:
            popup('login invalid', 'username or password is wrong')
            self.reset()
        else:
            self.reset()
            print('success')
            # TODO: next window
            
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
            popup('signup invalid', 'something went wrong, try to change the username')
        else:
            self.wm.current = 'login'
            # TODO: go to the next window

    def btn_goto_login(self):
        self.wm.current = 'login'

    def reset(self):
        self.username.text = ''
        self.password.text = ''
        