# config/application.rb

require_relative 'boot'
require 'rails/all'

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.

Bundler.require(*Rails.groups)

module SslExampleApp
  class Application < Rails::Application
    config.load_defaults 7.0

    # Example methods to demonstrate force_ssl settings

    def bad_ssl
      # <expect-error> force-ssl false - This is unsafe and should not be used in production.
      config.force_ssl = false
      Rails.logger.warn("SSL is disabled. This is insecure and should be avoided in production environments.")
    end

    def ok_ssl
      # safe force-ssl true - Enforces SSL, redirecting all HTTP requests to HTTPS.
      config.force_ssl = true
      Rails.logger.info("SSL is enabled. All HTTP requests will be redirected to HTTPS.")
    end
  end
end
