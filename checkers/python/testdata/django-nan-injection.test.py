import models
from django.http import HttpResponse
from app import get_price, deny, buy, fetch_obj


class Person(models.Model):
    first_name = models.CharField(...)
    last_name = models.CharField(...)
    birth_date = models.DateField(...)


##### True Positives #########
def test1(request):
    tid = request.POST.get("tid")

    price = get_price()

    # <expect-error>
    x = float(tid)

    if x < price:
        return deny()
    return buy()

def test2(request):
    tid = request.POST.get("tid")

    # <expect-error>
    bool(tid)

    # <expect-error>
    complex(tid)

def test3(request, something_else):
    tid = request.GET['tid']

    # <expect-error>
    float(tid)

    # <expect-error>
    bool(tid)

    # <expect-error>
    complex(tid)

def ok1(request, something_else):
    tid = request.POST.get("tid")

    obj = fetch_obj(tid)

    # <no-error>
    float(obj.num)

def ok2(request, something_else):
    tid = request.POST.get("tid")

    # <no-error>
    int(float(tid))

    # <no-error>
    float(int(tid))

    # <no-error>
    int(bool(tid))

def ok3(request):
    tid = request.POST.get("tid")

    if tid.lower() == "nan":
        raise ValueError

    # <no-error>
    num = float(tid)
    if num > get_price():
        buy()
    deny()