from textwrap import dedent
import base64
import os

def unsafe(request):
    code = request.POST.get('code')
    print("something")
    # <expect-error>
    eval(code)

def unsafe_inline(request):
    # <expect-error>
    eval(request.GET.get('data'))

def unsafe_dict(request):
    # <expect-error>
    eval(request.POST['data'])

def unsafe_var_dict(request):
    data_dict = request.POST['data']
    # <expect-error>
    eval(data_dict)

def safe(request):
    # ok: user-eval
    code_safe = """
    print('hello')
    """
    eval(dedent(code_safe))

def unsafe_percent_formatting(request):
    message = request.POST.get('message')
    print("do stuff here")
    code_percent_fmt = """
    print(%s)
    """ % message
    # <expect-error>
    eval(code_percent_fmt)

def unsafe_inline(request):
    # <expect-error>
    eval("print(%s)" % request.GET.get('message'))

def unsafe_dict(request):
    # <expect-error>
    eval("print(%s)" % request.POST['message'])

def safe(request):
    code = """
    print('hello')
    """
    # <no-error>
    eval(dedent(code))

def fmt_unsafe(request):
    message = request.POST.get('message')
    print("do stuff here")
    code_fmt_call = """
    print({})
    """.format(message)
    # <expect-error>
    eval(code_fmt_call)

def fmt_unsafe_inline(request):
    # <expect-error>
    eval("print({})".format(request.GET.get('message')))

def fmt_unsafe_dict(request):
    # <expect-error>
    eval("print({}, {})".format(request.POST['message'], "pwned"))

def fmt_safe(request):
    code = """
    print('hello')
    """
    # <no-error>
    eval(dedent(code))

def fstr_unsafe(request):
    message = request.POST.get('message')
    print("do stuff here")
    code_unsafe_fstring = f"""
    print({message})
    """
    # <expect-error>
    eval(code_unsafe_fstring)

def fstr_unsafe_inline(request):
    # <expect-error>
    eval(f"print({request.GET.get('message')})")

def fstr_unsafe_dict(request):
    # <expect-error>
    eval(f"print({request.POST['message']})")

def fstr_safe(request):
    var = "hello"
    code = f"""
    print('{var}')
    """
    # <no-error>
    eval(dedent(code))


def unsafe(request):
    message = request.POST.get('message')
    print("do stuff here")
    code_binary_op = """
    print(%s)
    """ % message
    # <expect-error>
    exec(code_binary_op)

def unsafe_inline(request):
    # <expect-error>
    exec("print(%s)" % request.GET.get('message'))

def unsafe_dict(request):
    # <expect-error>
    exec("print(%s)" % request.POST['message'])

def safe(request):
    # ok: user-exec-format-string
    code = """
    print('hello')
    """
    exec(dedent(code))

def fmt_unsafe(request):
    message = request.POST.get('message')
    print("do stuff here")
    code_format_call = """
    print({})
    """.format(message)
    # <expect-error>
    exec(code_format_call)

def fmt_unsafe_inline(request):
    # <expect-error>
    exec("print({})".format(request.GET.get('message')))

def fmt_unsafe_dict(request):
    # <expect-error>
    exec("print({}, {})".format(request.POST['message'], "pwned"))

def fmt_safe(request):
    # ok: user-exec-format-string
    code = """
    print('hello')
    """
    exec(dedent(code))

def code_execution(request):
    data = ''
    msg = ''
    first_name = ''
    if request.method == 'POST':

        # Clear out a previous success to reset the exercise
        try:
            os.unlink('p0wned.txt')
        except:
            pass

        
        first_name = request.POST.get('first_name', '')

        try:    # Try it the Python 3 way...
            # <expect-error>
            exec(base64.decodestring(bytes(first_name, 'ascii')))
        except TypeError:
            try:    # Try it the Python 2 way...
                # <expect-error>
                exec(base64.decodestring(first_name))
            except:
                pass
        except:
            pass


def danger(request):
    url = request.GET['url']
    # <expect-error>
    os.system('wget ' + url)

def danger2(request):
    image = request.POST['image']
    # <expect-error>
    os.system("./face-recognize %s --N 24" % image)

def danger3(request):
    url = request.GET['url']
    # <expect-error>
    os.system("nslookup " + url)

def ok(request):
    # <no-error>
    url = request.GET['url']
    os.system("echo 'hello'")
