Rails.application.configure do
   # <expect-error> openssl verify mode none
  config.action_mailer.smtp_settings = {
    openssl_verify_mode: OpenSSL::SSL::VERIFY_NONE
  }
end

Rails.application.configure do
   # <expect-error> openssl verify mode none
  config.action_mailer.smtp_settings = {
    openssl_verify_mode: "none"
  }
end

#Safe 
Rails.application.configure do
  config.action_mailer.smtp_settings = {
    openssl_verify_mode: OpenSSL::SSL::VERIFY_PEER
  }
end