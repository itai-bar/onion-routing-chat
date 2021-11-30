from Crypto.PublicKey import RSA
from Crypto.Cipher import PKCS1_OAEP
import Crypto.Cipher
import binascii


class Rsa():
    def __init__(self, public_key = None, private_key = None):
        if public_key == None or private_key == None:
            self._key_pair = RSA.generate(3072) # 3072 for 1024bits of key
            self.pem_public_key = self._key_pair.publickey().exportKey().decode('ascii')
            self.pem_private_key = self._key_pair.exportKey()
            self.public_key = self._key_pair.publickey()
        else:
            self.private_key = private_key
            self.public_key = public_key
    def encrypt(self, plaintext):
        return PKCS1_OAEP.new(self.public_key).encrypt(plaintext)
        
    def decrypt(self, ciphertext):
        return PKCS1_OAEP.new(self._key_pair).decrypt(ciphertext)

def main():
    a = Rsa()
    print("public:", a.pem_public_key)
    print("private:", a.pem_private_key)
    encrypted = input("Enter encrypted[1 2 3]:")
    encrypted = encrypted[1:]
    encrypted = encrypted[:-1]
    encrypted = bytes([int(byte) for byte in encrypted.split(" ")])
    print("decrypted:", a.decrypt(encrypted))
    

if __name__ == "__main__":
    main()