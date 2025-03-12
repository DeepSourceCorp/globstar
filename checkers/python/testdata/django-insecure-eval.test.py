from textwrap import dedent

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
