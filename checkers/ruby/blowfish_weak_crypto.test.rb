require 'crypt/blowfish'
require 'openssl'
require 'base64'

# Blowfish Encryption
def encrypt_blowfish(key, data)
  # <expect-error> weak blowfish crypto
  blowfish = Crypt::Blowfish.new(key)
  blowfish.encrypt_string(data)
end

def decrypt_blowfish(key, encrypted_data)
  # <expect-error> weak blowfish crypto
  blowfish = Crypt::Blowfish.new(key)
  blowfish.decrypt_string(encrypted_data)
end

# AES Encryption (Recommended)
def encrypt_aes(key, data)
  # Safe AES encryption
  cipher = OpenSSL::Cipher.new('aes-256-gcm')
  cipher.encrypt
  iv = cipher.random_iv
  cipher.key = key

  encrypted_data = cipher.update(data) + cipher.final
  auth_tag = cipher.auth_tag

  { encrypted: Base64.encode64(encrypted_data), iv: Base64.encode64(iv), tag: Base64.encode64(auth_tag) }
end

def decrypt_aes(key, encrypted_data, iv, tag)
  # Safe AES decryption
  decipher = OpenSSL::Cipher.new('aes-256-gcm')
  decipher.decrypt
  decipher.iv = Base64.decode64(iv)
  decipher.key = key
  decipher.auth_tag = Base64.decode64(tag)

  decipher.update(Base64.decode64(encrypted_data)) + decipher.final
end

key = "thisisaverysecurekey123456789012"  # AES requires a 32-byte key for AES-256
data = "Sensitive Data"

# Blowfish Encryption & Decryption
encrypted_blowfish = encrypt_blowfish(key, data)
decrypted_blowfish = decrypt_blowfish(key, encrypted_blowfish)

# AES Encryption & Decryption
encrypted_aes = encrypt_aes(key, data)
decrypted_aes = decrypt_aes(key, encrypted_aes[:encrypted], encrypted_aes[:iv], encrypted_aes[:tag])

# Output results
puts "Blowfish Encrypted Data: #{encrypted_blowfish}"
puts "Blowfish Decrypted Data: #{decrypted_blowfish}"
puts "AES Encrypted Data: #{encrypted_aes[:encrypted]}"
puts "AES Decrypted Data: #{decrypted_aes}"
