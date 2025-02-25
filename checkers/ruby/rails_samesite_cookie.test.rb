# config/initializers/session_store.rb

# Insecure: SameSite set to 'None' (can lead to CSRF attacks if Secure flag is not enabled)
# <expect-error> samesite none
Rails.application.config.session_store :cookie_store, key: "_myapp_session", same_site: :none

# Insecure: Missing 'same_site' attribute (defaults to Lax but can be risky in certain cases)
# <expect-error> empty samesite
Rails.application.config.session_store :cookie_store, key: "_myapp_session", httponly: true

# Secure: SameSite set to 'Strict' (Prevents CSRF attacks by restricting cookies to first-party requests)
Rails.application.config.session_store :cookie_store, key: "_myapp_session", same_site: :strict, httponly: true, secure: Rails.env.production?
