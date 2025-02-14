# cf. https://github.com/PyCQA/bandit/blob/b78c938c0bd03d201932570f5e054261e10c5750/examples/crypto-md5.py

from cryptography.hazmat.primitives import hashes

# <expect-error>
hashes.SHA1()
# <no-error>
hashes.SHA256()
# <no-error>
hashes.SHA3_256()
