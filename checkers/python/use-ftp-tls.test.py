import ftplib
import ssl

def bad():
    # <expect-error>
    ftpc = ftplib.FTP("example.com", "user", "pass")

def ok():
    # <no-error>
    ftpc = ftplib.FTP_TLS("example.com", "user", "pass", context=ssl.create_default_context())
