from client import Client
from tor.crypto import Rsa

from kivy.app import App
from kivy.lang import Builder
from kivy.uix.screenmanager import ScreenManager

from windows import LoginWindow

class WindowManager(ScreenManager):
    pass

kv = Builder.load_file('chat.kv')
window_manager = WindowManager()

screens = [LoginWindow(name='login')] # TODO: add every screen here
for screen in screens:
    window_manager.add_widget(screen)

class ChatApp(App):
    def build(self):
        return window_manager

window_manager.current = 'login'
client = Client()

if __name__ == '__main__':
    ChatApp().run()