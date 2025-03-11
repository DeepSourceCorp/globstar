require 'digest'

class HashGenerator
  # Insecure: Uses MD5, which is vulnerable to collision attacks.
  def generate_md5_hash(string)
    puts "Warning: MD5 is insecure and should not be used for cryptographic purposes."
    # <expect-error> use of md5
    Digest::MD5.hexdigest(string)
  end

  # Secure: Uses SHA-256, a recommended hash function for cryptographic use.
  def generate_sha256_hash(string)
    # safe hash function
    Digest::SHA256.hexdigest(string)
  end
end

hash_generator = HashGenerator.new

input = "sensitive_data"

md5_hash = hash_generator.generate_md5_hash(input)
puts "MD5 Hash (insecure): #{md5_hash}"

sha256_hash = hash_generator.generate_sha256_hash(input)
puts "SHA-256 Hash (secure): #{sha256_hash}"
