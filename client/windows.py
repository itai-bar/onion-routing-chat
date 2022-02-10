# login window 
# signup window
# rooms window with a room join option
# chat window
# room managment for admin only

from kivy.uix.screenmanager import Screen
from kivy.properties import ObjectProperty
from kivy.uix.popup import Popup
from kivy.uix.label import Label

class LoginWindow(Screen):
    username = ObjectProperty(None)
    password = ObjectProperty(None)
    
    def btn_login(self):
        pass

    def btn_goto_signup(self):
        pass

    def reset(self):
        pass