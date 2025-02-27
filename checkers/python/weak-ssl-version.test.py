# cf. https://github.com/PyCQA/bandit/blob/b1411bfb43795d3ffd268bef17a839dee954c2b1/examples/ssl-insecure-version.py

import ssl
from pyOpenSSL import SSL

# <expect-error>
ssl.wrap_socket(ssl_version=ssl.PROTOCOL_SSLv2)
# <expect-error>
SSL.Context(method=SSL.SSLv2_METHOD)
# <expect-error>
SSL.Context(method=SSL.SSLv23_METHOD)

# <no-error>
ssl.wrap_socket(ssl_version=ssl.PROTOCOL_TLSv1_2)

# <expect-error>
some_other_method(ssl_version=ssl.PROTOCOL_SSLv2)
# <expect-error>
some_other_method(method=SSL.SSLv2_METHOD)
# <expect-error>
some_other_method(method=SSL.SSLv23_METHOD)

# <expect-error>
ssl.wrap_socket(ssl_version=ssl.PROTOCOL_SSLv3)
# <expect-error>
ssl.wrap_socket(ssl_version=ssl.PROTOCOL_TLSv1)
# <expect-error>
SSL.Context(method=SSL.SSLv3_METHOD)
# <expect-error>
SSL.Context(method=SSL.TLSv1_METHOD)

# <expect-error>
some_other_method(ssl_version=ssl.PROTOCOL_SSLv3)
# <expect-error>
some_other_method(ssl_version=ssl.PROTOCOL_TLSv1)
# <expect-error>
some_other_method(method=SSL.SSLv3_METHOD)
# <expect-error>
some_other_method(method=SSL.TLSv1_METHOD)

ssl.wrap_socket()

# <expect-error>
def open_ssl_socket(version=ssl.PROTOCOL_SSLv2):
    pass

# <expect-error>
def open_ssl_socket(version=SSL.SSLv2_METHOD):
    pass

# <expect-error>
def open_ssl_socket(version=SSL.SSLv23_METHOD):
    pass

# <expect-error>
def open_ssl_socket(version=SSL.TLSv1_1_METHOD):
    pass
