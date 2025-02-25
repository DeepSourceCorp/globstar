class AdminController < ApplicationController
  # Insecure: Hardcoded credentials (AVOID THIS)
  # <expect-error> http basic auth hardcoded password
  http_basic_authenticate_with name: "admin", password: "password"

  # Secure: Using environment variables (Recommended)
  http_basic_authenticate_with name: "admin", password: ENV['ADMIN_PASSWORD']

  def index
    render plain: "Welcome, Admin!"
  end
end
