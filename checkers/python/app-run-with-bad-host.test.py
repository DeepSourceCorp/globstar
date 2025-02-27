# Ref: https://owasp.org/Top10/A01_2021-Broken_Access_Control [OWASP A01:2021]

# <expect-error>
app.run(host="0.0.0.0")

# <expect-error>
app.run("0.0.0.0")

# OK
foo.run("0.0.0.0")
