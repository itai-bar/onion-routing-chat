from kivy.app import App
from kivy.lang import Builder
from kivy.uix.screenmanager import ScreenManager
from kivy.event import EventDispatcher
from kivy.properties import ObjectProperty

from windows import EmptyWindow, WelcomeWindow, LoginWindow, MainWindow, RoomsWindow, SignupWindow, ChatWindow

class MyState(EventDispatcher):
    current_room = ''

class WindowManager(ScreenManager):
    statedata = ObjectProperty(MyState())

class ChatApp(App):
    def __init__(self, wm, **kwargs):
        self.wm = wm
        super().__init__(**kwargs)
        

    def build(self):
        self.prevent_gui_duplication()

        return self.wm
    
    def prevent_gui_duplication(slef):
        from kivy.resources import resource_find
        filename = 'chat.kv'
        filename = resource_find(filename) or filename
        if filename in Builder.files:
            Builder.unload_file(filename)
        Builder.load_file(filename)

        
class Chat:
    def __init__(self) -> None:
        kv = Builder.load_file('chat.kv')
        self.window_manager = WindowManager()
        
        self._screens = [LoginWindow(self.window_manager, name='login'),
                         WelcomeWindow(self.window_manager, name='welcome'),
                         SignupWindow(self.window_manager, name='signup'),
                         RoomsWindow(self.window_manager, name='rooms'),
                         MainWindow(self.window_manager, name='main'),
                         ChatWindow(self.window_manager, name='chat')] 

        self.window_manager.add_widget(EmptyWindow(self.window_manager, name='empty')) # for better UE
        for screen in self._screens:
            self.window_manager.add_widget(screen)
        
        self.window_manager.current = 'welcome'


if __name__ == '__main__':
    chat = Chat()
    ChatApp(chat.window_manager).run()