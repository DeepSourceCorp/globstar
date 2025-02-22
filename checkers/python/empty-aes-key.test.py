from Crypto.Ciphers import AES

def bad1():
    # <expect-error>
    cipher = AES.new("", AES.MODE_CFB, iv)
    msg = iv + cipher.encrypt(b'Attack at dawn')

def ok1(key):
    # <no-error>
    cipher = AES.new(key, AES.MODE_EAX, nonce=nonce)
    plaintext = cipher.decrypt(ciphertext)
