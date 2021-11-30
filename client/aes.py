from Crypto import Random
from Crypto.Cipher import AES

KEY_SIZE = 32

class Aes:
    def __init__(self, key = None):
        if key == None:
            self._key = Random.get_random_bytes(KEY_SIZE)
        else:
            self._key = key

    def encrypt(self, buf):
        buf = self.pad(buf, AES.block_size)
        # a new random iv in every encryption
        iv = Random.new().read(AES.block_size) 
        cipher = AES.new(self._key, AES.MODE_CBC, iv)
        return iv + cipher.encrypt(buf.encode())
 
    def decrypt(self, buf):
        # getting the iv out
        iv = buf[:AES.block_size] 
        buf = buf[AES.block_size:]

        cipher = AES.new(self._key, AES.MODE_CBC, iv)
        return self.unpad(cipher.decrypt(buf)).decode()

    def pad(self, buf, size): 
        return buf + (size - len(buf) % size) * chr(size - len(buf) % size)

    def unpad(self, buf):
        return buf[:-ord(buf[len(buf)-1:])]
