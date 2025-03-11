import os
from cryptography.hazmat import backends
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.hazmat.primitives.asymmetric import dsa
from cryptography.hazmat.primitives.asymmetric import rsa


# <no-error>
ec.generate_private_key(curve=ec.SECP256K1,
                         backend=backends.default_backend())

# <no-error>
ec.generate_private_key(ec.SECP256K1,
                         backends.default_backend())

# <no-error>
ec.generate_private_key(curve=os.getenv("EC_CURVE"),
                         backend=backends.default_backend())

# <expect-error>
ec.generate_private_key(curve=ec.SECP192R1,
                         backend=backends.default_backend())

# <expect-error>
ec.generate_private_key(ec.SECT163K1,
                         backends.default_backend())

# <no-error>
dsa.generate_private_key(key_size=2048,
                         backend=backends.default_backend())

# <no-error>
dsa.generate_private_key(2048,
                         backend=backends.default_backend())

# <expect-error>
dsa.generate_private_key(key_size=1024,
                         backend=backends.default_backend())

# <expect-error>
dsa.generate_private_key(1024,
                         backend=backends.default_backend())

rsa.generate_private_key(public_exponent=65537,
# <no-error>
                         key_size=2048,
                         backend=backends.default_backend())

rsa.generate_private_key(65537,
# <no-error>
                         2048,
                         backends.default_backend())

rsa.generate_private_key(public_exponent=65537,
# <no-error>
                         key_size=os.getenv("KEY_SIZE"),
                         backend=backends.default_backend())

rsa.generate_private_key(65537,
# <no-error>
                         2048,
                         backends.default_backend())

# <expect-error>
rsa.generate_private_key(public_exponent=65537,
                         key_size=1024,
                         backend=backends.default_backend())

# <expect-error>
rsa.generate_private_key(65537,
                         1024,
                         backends.default_backend())
