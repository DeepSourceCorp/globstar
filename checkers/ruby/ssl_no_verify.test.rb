require 'net/http'
require 'openssl'

# Insecure: Disabling SSL certificate verification (Vulnerable to MITM attacks)
def insecure_ssl_request(url)
  uri = URI.parse(url)
  http = Net::HTTP.new(uri.host, uri.port)
  http.use_ssl = true
  
  # <expect-error> ssl no verify (Disables SSL verification)
  http.verify_mode = OpenSSL::SSL::VERIFY_NONE

  request = Net::HTTP::Get.new(uri.request_uri)
  response = http.request(request)
  response.body
end

# Secure: Enabling SSL certificate verification (Prevents MITM attacks)
def secure_ssl_request(url)
  uri = URI.parse(url)
  http = Net::HTTP.new(uri.host, uri.port)
  http.use_ssl = true

  # Secure verification of SSL certificates
  http.verify_mode = OpenSSL::SSL::VERIFY_PEER

  request = Net::HTTP::Get.new(uri.request_uri)
  response = http.request(request)
  response.body
end