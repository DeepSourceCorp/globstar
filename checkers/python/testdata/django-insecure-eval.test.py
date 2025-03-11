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
