require 'openssl'
require 'base64'
require 'bcrypt'

def generate_dsa_key
  # <expect-error> Use of DSA key (not recommended)
  dsa = OpenSSL::PKey::DSA.new(1024)
  dsa
end

# Safe - Hashing a password securely using bcrypt
def hash_password(password)
  BCrypt::Password.create(password) # Automatically generates a salt
end