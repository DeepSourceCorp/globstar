require 'digest'

class HashGenerator
  # Insecure: Uses SHA-1, which is vulnerable to collision attacks.
  def generate_sha1_hash(string)
    # <expect-error> use of sha1
    Digest::SHA1.hexdigest(string)
  end

  # Secure: Uses SHA-256, a stronger and recommended hash function.
  def generate_sha256_hash(string)
    # safe hash function
    Digest::SHA256.hexdigest(string)
  end
end

hash_generator = HashGenerator.new

input = "sensitive_data"

sha1_hash = hash_generator.generate_sha1_hash(input)
puts "SHA-1 Hash (insecure): #{sha1_hash}"

sha256_hash = hash_generator.generate_sha256_hash(input)
puts "SHA-256 Hash (secure): #{sha256_hash}"
