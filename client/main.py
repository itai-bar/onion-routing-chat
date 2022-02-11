from kivy.app import App
from kivy.lang import Builder
from kivy.uix.screenmanager import ScreenManager

from windows import LoginWindow, RoomsWindow, SignupWindow

class WindowManager(ScreenManager):
    pass

class ChatApp(App):
    def __init__(self, wm, **kwargs):
        self.wm = wm
        super().__init__(**kwargs)

    def build(self):
        return self.wm

class Chat:
    def __init__(self) -> None:
        kv = Builder.load_file('chat.kv')
        self.window_manager = WindowManager()

        self._screens = [LoginWindow(self.window_manager, name='login'),
                         SignupWindow(self.window_manager, name='signup'),
                         RoomsWindow(self.window_manager, name='rooms')] 
        for screen in self._screens:
            self.window_manager.add_widget(screen)
        
        self.window_manager.current = 'login'
        #self.window_manager.current = 'rooms'

if __name__ == '__main__':
    chat = Chat()
    ChatApp(chat.window_manager).run()