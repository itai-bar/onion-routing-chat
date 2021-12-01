from Crypto import Random
from Crypto.Cipher import AES

class Aes:
    def __init__(self, key=None, key_size=32):
        if key == None:
            self._key = Random.get_random_bytes(key_size)
        else:
            self._key = key
    
    def encrypt(self, text: bytes) -> bytes:
        """encrypts a given text using aes in CFB mode

        Args:
            text (bytes): aes encrypted text

        Returns:
            bytes: encrypted text
        """
        rem = len(text) % 16
        padded = str.encode(text) + (b'\0' * (16 - rem)) if rem > 0 else str.encode(text)

        iv = Random.new().read(AES.block_size)
        cipher = AES.new(self._key, AES.MODE_CFB, iv, segment_size=128)
        enc = cipher.encrypt(padded)[:len(text)]
        return iv + enc
    

    def decrypt(self, text: bytes) -> bytes:
        """decrypts a message using aes in CFB mode

        Args:
            text (bytes): an encrypted message

        Returns:
            bytes: the original message
        """
        iv, value = text[:16], text[16:] # extracting the iv from the text
        rem = len(value) % 16
        padded = value + (b'\0' * (16 - rem)) if rem > 0 else value 
        cipher = AES.new(self._key, AES.MODE_CFB, iv, segment_size=128)
        return cipher.decrypt(padded)[:len(value)]
