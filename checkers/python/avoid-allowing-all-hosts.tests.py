# Match 
# <expect-error>
ALLOWED_HOSTS = ["*"]
# <expect-error>
ALLOWED_HOSTS = ['*']
# <expect-error>
ALLOWED_HOSTS = ["google.com","*","facebook.com"]
# <expect-error>
ALLOWED_HOSTS = ["*","google.com","facebook.com"]
# <expect-error>
ALLOWED_HOSTS = ["google.com","facebook.com","*"]

# Not Macth
# <no-error>
ALLOWED_HOSTS = ["google.com","facebook.com"]