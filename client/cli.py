from chat_client import ChatClient

import curses
from curses import wrapper
from curses.textpad import rectangle, Textbox

import json

client = ChatClient()
client.auth()

def parse_command(cmd: str) -> str:
    if cmd.startswith('/'):
        req, args = cmd.split()[0], cmd.split()[1:]
        req = req[1:]
        
        calls_map = {   'login' :       (2, client.login), 
                        'signup' :      (2, client.register),
                        'create_room' : (2, client.create_room),
                        'delete_room' : (2, client.delete_room),
                        'join_room' :   (2, client.join_room),
                        'kick_user' :   (2, client.kick_user),
                        'ban_user' :    (2, client.ban_user),
                        'unban_user' :  (2, client.unban_user)
                    }
    
        try:
            if len(args) == calls_map[req][0]:
                return calls_map[req][1](*args)['info']
            else:
                return f'{req} has {calls_map[req][0]} args'
        except KeyError:
            return f'{req} is not a command!'

    return ''

def msg_to_str(msg):
    pass
        
def main(screen : curses.window):
    (max_lines, max_cols) = screen.getmaxyx()
    max_lines -= 1
    max_cols -= 1
    screen.clear()

    output_win = curses.newwin(max_lines-1, 100, 0, 0)
    output_win.scrollok(1)


    command_win = curses.newwin(1, 100, max_lines, 0)
    command_tb = Textbox(command_win)

    screen.refresh()

    display = 'command: '
    output_l = 0

    while True:
        command_win.clear()
        output_win.addstr(max_lines-2, 0, display) 
        output_win.refresh()
        
        command_tb.edit()
        command = command_tb.gather()
        command_win.clear()

        resp = str(parse_command(command))
        output_win.addstr(output_l, 0, resp)
        output_win.refresh()
        output_l += 1



if __name__ == '__main__':
    wrapper(main)