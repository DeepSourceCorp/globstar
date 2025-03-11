from django.http import HttpResponse
from django.views.decorators.csrf import csrf_exempt

# <expect-error>
@csrf_exempt
def my_view(request):
    return HttpResponse('Hello world')

import django

# <expect-error>
@django.views.decorators.csrf.csrf_exempt
def my_view2(request):
    return HttpResponse('Hello world')
