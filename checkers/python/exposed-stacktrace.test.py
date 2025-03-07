import logging
import traceback

from django.http import HttpResponse

LOGGER = logging.getLogger(__name__)


# <expect-error>
def debug_view(request):
    error_trace = traceback.format_exc()
    return HttpResponse(error_trace, content_type="text/plain", status=500)


# <no-error>
def save_view(request):
    tb = traceback.format_stack()
    LOGGER.error(tb)
    return HttpResponse("Internal Server Error", content_type="text/plain", status=500)
