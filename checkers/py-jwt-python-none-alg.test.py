import jwt

def bad1():
    # <expect-error>
    encoded = jwt.encode({'some': 'payload'}, None, algorithm='none')
    return encoded

def bad2(encoded):
    # <expect-error>
    jwt.decode(encoded, None, algorithms=['none'])
    return encoded

def ok(secret_key):
    # <no-error>
    encoded = jwt.encode({'some': 'payload'}, secret_key, algorithm='HS256')
    return encoded
