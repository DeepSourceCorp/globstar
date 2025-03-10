# cf. https://github.com/PyCQA/bandit/blob/b78c938c0bd03d201932570f5e054261e10c5750/examples/ciphers.py

from cryptography.hazmat.primitives.ciphers import Cipher
from cryptography.hazmat.primitives.ciphers import algorithms
from cryptography.hazmat.primitives.ciphers import modes
from cryptography.hazmat.backends import default_backend
from struct import pack

# <expect-error>
cipher = Cipher(algorithms.ARC4(key), mode=None, backend=default_backend())
encryptor = cipher.encryptor()
ct = encryptor.update(b"a secret message")

# <no-error>
cipher = Cipher(algorithms.AES(key), modes.CBC(iv), backend=default_backend())
encryptor = cipher.encryptor()
ct = encryptor.update(b"a secret message") + encryptor.finalize()

# <expect-error>
cipher = Cipher(algorithms.Blowfish(key), mode=None, backend=default_backend())
encryptor = cipher.encryptor()
ct = encryptor.update(b"a secret message")

# <no-error>
cipher = Cipher(algorithms.AES(key), modes.CBC(iv), backend=default_backend())
encryptor = cipher.encryptor()
ct = encryptor.update(b"a secret message") + encryptor.finalize()

# <expect-error>
cipher = Cipher(algorithms.IDEA(key), mode=None, backend=default_backend())
encryptor = cipher.encryptor()
ct = encryptor.update(b"a secret message")

# <no-error>
cipher = Cipher(algorithms.AES(key), modes.CBC(iv), backend=default_backend())
encryptor = cipher.encryptor()
ct = encryptor.update(b"a secret message") + encryptor.finalize()