import traceback

from django.http import HttpResponse


def debug_view():
    error_trace = traceback.format_exc()
    # <expect-error>
    return HttpResponse(error_trace, content_type="text/plain", status=500)
