# config/initializers/session_store.rb
 # <expect-error> httpOnly cookie false
Rails.application.config.session_store :cookie_store, key: "_secure_app_session",

  # Insecure: `httponly: false` allows JavaScript access (AVOID THIS)
  httponly: false, secure: Rails.env.production?

# Secure Alternative: `httponly: true` (Recommended)
# This prevents JavaScript from accessing session cookies.
Rails.application.config.session_store :cookie_store, key: "_secure_app_session",
  # Safe httponly cookie true
  httponly: true, secure: Rails.env.production?
