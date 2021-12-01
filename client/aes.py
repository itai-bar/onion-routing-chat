from Crypto import Random
from Crypto import Cipher
from Crypto.Cipher import AES
from base64 import b64encode, b64decode

class Aes:
    def __init__(self, key=None, key_size=32):
        if key == None:
            self._key = Random.get_random_bytes(key_size)
        else:
            self._key = key
    
    def encrypt(self, text: str) -> str:
        """encrypted a given string using aes in CFB mode

        Args:
            text (str): aes encrypted string

        Returns:
            str: encrypted and base64 encoded string
        """
        rem = len(text) % 16
        padded = str.encode(text) + (b'\0' * (16 - rem)) if rem > 0 else str.encode(text)

        iv = Random.new().read(AES.block_size)
        cipher = AES.new(self._key, AES.MODE_CFB, iv, segment_size=128)
        enc = cipher.encrypt(padded)[:len(text)]
        return b64encode(iv + enc).decode()
    

    def decrypt(self, text: str) -> str:
        """decrypts a message using aes in CFB mode and encoded with base64

        Args:
            text (str): an encrypted message

        Returns:
            str: the original message
        """
        text = b64decode(text) # the text was base64 encoded
        iv, value = text[:16], text[16:] # extracting the iv from the text
        rem = len(value) % 16
        padded = value + (b'\0' * (16 - rem)) if rem > 0 else value 
        cipher = AES.new(self._key, AES.MODE_CFB, iv, segment_size=128)
        return cipher.decrypt(padded)[:len(value)].decode()
