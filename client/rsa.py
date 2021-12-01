from Crypto.PublicKey import RSA
from Crypto.Cipher import PKCS1_OAEP
from Crypto.Hash import SHA256


KEY_SIZE = 4096

class Rsa():
    def __init__(self, given_pem_public_key=None, given_pem_private_key=None):
        """C'tor that generates RSA key in case that one of the arguments is None
        and creates RSA key in case that given public and private pems of keys

        Args:
            given_pem_public_key (bytes, optional): public key(PEM format). Defaults to None.
            given_pem_private_key (bytes, optional): private key(PEM format). Defaults to None.
        """
        if given_pem_public_key == None or given_pem_private_key == None:
            self._key_pair = RSA.generate(KEY_SIZE)
            self.pem_public_key = self._key_pair.publickey().exportKey()
            self.pem_private_key = self._key_pair.exportKey()
            self.public_key = self._key_pair.publickey()
        else:
            self.pem_private_key = given_pem_private_key
            self.pem_public_key = given_pem_public_key
            self.public_key = RSA.import_key(given_pem_public_key)
            self._key_pair = RSA.import_key(given_pem_private_key)

    def encrypt(self, plaintext):
        """function encrypts plaintext with initalized public key(self.public_key)

        Args:
            plaintext (bytes): text to be encrypted

        Returns:
            bytes: encrypted data in bytes
        """
        return PKCS1_OAEP.new(self.public_key, SHA256.new()).encrypt(plaintext)
        
    def decrypt(self, ciphertext):
        """function decrypts ciphertext with initalized private key(self._key_pair)

        Args:
            ciphertext (bytes): cipher to be decrypted

        Returns:
            bytes: decrypted data in bytes
        """
        return PKCS1_OAEP.new(self._key_pair, SHA256.new()).decrypt(ciphertext)
