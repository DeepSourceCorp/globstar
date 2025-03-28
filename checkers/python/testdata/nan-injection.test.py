import models
from django.http import HttpResponse
from app import get_price, deny, buy, fetch_obj
import os
import flask
import hashlib
import requests


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

app = flask.Flask(__name__)

@app.route("/buy/<tid>")
def buy_thing(tid):
    price = get_price()

    # <expect-error>
    x = float(tid)

    if x < price:
        return deny()
    return buy()

@app.route("unit_1")
def unit_1():
    tid = flask.request.args.get("tid")

    # <expect-error>
    bool(tid)

    # <expect-error>
    complex(tid)

@app.route("unit_1_5")
def unit_1_5():
    tid = flask.request.args["tid"]

    # <expect-error>
    float(tid)

    # <expect-error>
    bool(tid)

    # <expect-error>
    complex(tid)

@app.route("unit_2")
def unit_2():
    tid = flask.request.args.get("tid")

    # <no-error>
    bool(int(tid))

    # <no-error>
    float(int(tid))

@app.route("unit_3")
def unit_3():
    tid = flask.request.args.get("tid")

    # <no-error>
    obj = fetch_obj(tid)

    # <no-error>
    num = float(obj.num)

@app.route("/drip")
def drip():
    # <expect-error>
    duration = float(flask.request.args.get("duration", 2))
    numbytes = min(int(flask.request.args.get("numbytes", 10)), (10 * 1024 * 1024))  # set 10MB limit
    code = int(flask.request.args.get("code", 200))

    if numbytes <= 0:
        response = Response("number of bytes must be positive", status=400)
        return response

    # <expect-error>
    delay = float(flask.request.args.get("delay", 0))
    if delay > 0:
        time.sleep(delay)

    pause = duration / numbytes

    def generate_bytes():
        for i in xrange(numbytes):
            yield b"*"
            time.sleep(pause)

    response = Response(
        generate_bytes(),
        headers={
            "Content-Type": "application/octet-stream",
            "Content-Length": str(numbytes),
        },
    )

    response.status_code = code

    return response
