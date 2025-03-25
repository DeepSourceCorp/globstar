import csv
import django
import flask
import io

from data import get_data
from util import chroot

def a(request):
    stream = io.StringIO()
    writer = csv.writer(stream)
    data = get_data()
    title = request.POST.get("title")
    title_row = title + ("," * len(data[0]) - 1)
    # <expect-error>
    writer.writerow(title_row)
    writer.writerows(data)
    header = "A header {}".format(title)
    # <expect-error>
    writer.writeheader(header)
    rowReqVal = request.build_absolute_uri("row")
    row = f"A row {rowReqVal}"
    # <no-error>
    writer.writerow(row)
    stream.flush()
    stream.seek(0)
    return stream.read()

@app.route("ok")
def ok():
    with open("data.csv") as fin:
        # <no-error>
        reader = csv.reader(fin)
        lines = [line for line in reader]
    return '\n'.join(lines)



app = flask.Flask(__name__)

@app.route("a/<title>")
def a(title):
    stream = io.StringIO()
    writer = csv.writer(stream)
    data = get_data()
    title_row = title + ("," * len(data[0]) - 1)
    # <expect-error>
    writer.writerow(title_row)
    writer.writerows(data)
    stream.flush()
    stream.seek(0)
    return stream.read()

@app.route("ok")
def ok():
    with open("data.csv") as fin:
        # <no-error>
        reader = csv.reader(fin)
        lines = [line for line in reader]
    return '\n'.join(lines)
