import os

def main():
    # These should be flagged
    # <expect-error>
    connect(host="example.com", token="hqd#18ey283y28wdbbcwbd1ueh1ue2h")
    # <expect-error>
    set_password(password="A3b$c8d!eF9gHiJkLmNoPqRsTuVwXyZ")
    # <expect-error>
    configure(key="AKIATESTKEYTESTKEYTESTKEYTEST", secret="TEST/SECRET/KEY/EXAMPLE/1234567890")

    # These should NOT be flagged
    # <no-error>
    set_password(password="password123")  # Low entropy
    # <no-error>
    configure(username="test_user", value=42)  # Not a sensitive argument

    # Should not flag non-string values
    # <no-error>
    set_token(token=os.getenv("API_TOKEN"))
    
    # Should not flag commented out code
    # <no-error>
    # connect(host="example.com", token="commented_out_secret")

    # Edge cases
    # <no-error>
    empty_string(arg="")  # Empty string
    # <no-error>
    numeric_value(key=12345)  # Not a string 