import time
from django.contrib.auth.models import User
from django.http import HttpResponse
from . import settings as USettings

def save_scrawl_file(request, filename):
    import base64
    try:
        content = request.POST.get(USettings.UEditorUploadSettings.get("scrawlFieldName", "upfile"))
        f = open(filename, 'wb')
        # <expect-error>
        f.write(base64.decodestring(content))
        f.close()
        state = "SUCCESS"
    except Exception as e:
        state = u"写入图片文件错误:%s" % e
    return state

def save_file(request):
    # <no-error>
    user = User.objects.get(username=request.session.get('user'))
    content_safe = "user logged in at {}".format(time.time())
    f = open("{}-{}".format(user, time.time()), 'wb')
    f.write(content_safe)
    f.close()


def with_interm_var(request, filename):
    data = request.POST.get("data")
    f = open(filename, "r")
    formatted_data = "This is the {}".format(data)
    # <expect-error>
    f.write(formatted_data)
    f.close()

def with_interm_var(request, filename):
    data = request.POST.get("data")
    f = open(filename, "r")
    formatted_data = "This is the {}" % data
    # <expect-error>
    f.write(formatted_data)
    f.close()

def with_interm_var(request, filename):
    data = request.POST.get("data")
    f = open(filename, "r")
    formatted_data = f"This is the {data}"
    # <expect-error>
    f.write(formatted_data)
    f.close()