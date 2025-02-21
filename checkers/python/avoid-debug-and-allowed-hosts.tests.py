# Match 
# <expect-error>
DEBUG = True
ALLOWED_HOSTS = ["*"]
ALLOWED_HOSTS = ['*']
ALLOWED_HOSTS = ["google.com","*","facebook.com"]
ALLOWED_HOSTS = ["*","google.com","facebook.com"]
ALLOWED_HOSTS = ["google.com","facebook.com","*"]
# Not Macth
# <no-error>
ALLOWED_HOSTS = ["google.com","facebook.com"]
DEBUG = False