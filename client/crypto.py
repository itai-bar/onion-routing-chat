from Crypto import Random
from Crypto.Cipher import AES
from Crypto.PublicKey import RSA
from Crypto.Cipher import PKCS1_OAEP
from Crypto.Hash import SHA256
import const

class Aes:
    def __init__(self, key : bytes=None, key_size : int=32):
        if key == None:
            self._key = Random.get_random_bytes(key_size)
        else:
            self._key = key
    
    def encrypt(self, text: bytes) -> bytes:
        """Encrypts a given text using aes in CFB mode

        Args:
            text (bytes): aes encrypted text

        Returns:
            bytes: encrypted text
        """
        rem = len(text) % 16
        padded = text + (b'\0' * (16 - rem)) if rem > 0 else text

        iv = Random.new().read(AES.block_size)
        cipher = AES.new(self._key, AES.MODE_CFB, iv, segment_size=128)
        enc = cipher.encrypt(padded)[:len(text)]
        return iv + enc
    

    def decrypt(self, text: bytes) -> bytes:
        """Decrypts a message using aes in CFB mode

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



class Rsa():
    def __init__(self, given_pem_public_key : bytes=None, given_pem_private_key : bytes=None):
        """C'tor that generates RSA key in case that one of the arguments is None
        and creates RSA key in case that given public and private pems of keys

        Args:
            given_pem_public_key (bytes, optional): public key(PEM format). Defaults to None.
            given_pem_private_key (bytes, optional): private key(PEM format). Defaults to None.
        """
        if given_pem_public_key == None or given_pem_private_key == None:
            self._key_pair = RSA.generate(const.KEY_SIZE)
            self.pem_public_key = self._key_pair.publickey().exportKey()
            self.pem_private_key = self._key_pair.exportKey()
            self.public_key = self._key_pair.publickey()
        else:
            self.pem_private_key = given_pem_private_key
            self.pem_public_key = given_pem_public_key
            self.public_key = RSA.import_key(given_pem_public_key)
            self._key_pair = RSA.import_key(given_pem_private_key)

    def encrypt(self, plaintext : bytes) -> bytes:
        """Function encrypts plaintext with initalized public key(self.public_key)

        Args:
            plaintext (bytes): text to be encrypted

        Returns:
            bytes: encrypted data in bytes
        """
        return PKCS1_OAEP.new(self.public_key, SHA256.new()).encrypt(plaintext)
        
    def decrypt(self, ciphertext : bytes) -> bytes:
        """Function decrypts ciphertext with initalized private key(self._key_pair)

        Args:
            ciphertext (bytes): cipher to be decrypted

        Returns:
            bytes: decrypted data in bytes
        """
        return PKCS1_OAEP.new(self._key_pair, SHA256.new()).decrypt(ciphertext)


def decrypt_by_order(ciphertext : bytes, aes_keys_for_decryption : list) -> bytes:
    """Function decrypted ciphertext by order of given list of aes-keys

    Args:
        ciphertext (bytes): encrypted text
        aes_keys_for_decryption (list): keys for decryption(ordered)

    Returns:
        bytes: decrypted ciphertext as plaintext
    """
    decrypted = ciphertext  # first time initialization
    for key in aes_keys_for_decryption:
        if isinstance(key, Aes):
            decrypted = key.decrypt(decrypted)
    return decrypted