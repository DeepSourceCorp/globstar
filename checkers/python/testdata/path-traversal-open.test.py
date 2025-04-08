import flask
import json
import re, os
from django.http import HttpResponse
from somewhere import APIView

app = flask.Flask(__name__)

@app.route("/route_param/<route_param>")
def route_param(route_paramaaa):
    print("blah")
    # <expect-error>
    return open(route_paramaaa, 'r').read()

@app.route("/route_param_ok/<route_param>")
def route_param_ok(route_param):
    print("blah")
    # <no-error>
    return open("this is safe", 'r').read()

@app.route("/route_param_with/<route_param>")
def route_param_with(route_param):
    print("blah")
    # <expect-error>
    with open(route_param, 'r') as fout:
        return fout.read()

@app.route("/route_param_with_ok/<route_param>")
def route_param_with_ok(route_param):
    print("blah")
    # <no-error>
    with open("this is safe", 'r') as fout:
        return fout.read()

@app.route("/route_param_with_concat/<route_param>")
def route_param_with_concat(route_param):
    print("blah")
    # <expect-error>
    with open(route_param + ".csv", 'r') as fout:
        return fout.read()

@app.route("/get_param", methods=["GET"])
def get_param():
    param = flask.request.args.get("param")
    # <expect-error>
    f = open(param, 'w')
    f.write("hello world")

@app.route("/get_param_inline_concat", methods=["GET"])
def get_param_inline_concat():
    # <expect-error>
    return open("echo " + flask.request.args.get("param"), 'r').read()

@app.route("/get_param_concat", methods=["GET"])
def get_param_concat():
    param = flask.request.args.get("param")
    # <expect-error>
    return open(param + ".BDSMcsv", 'r').read()

@app.route("/get_param_format", methods=["GET"])
def get_param_format():
    param = flask.request.args.get("param")
    # <expect-error>
    return open("{}.csv".format(param)).read()

@app.route("/get_param_percent_format", methods=["GET"])
def get_param_percent_format():
    param = flask.request.args.get("param")
    # <expect-error>
    return open("echo %s" % (param,), 'r').read()

@app.route("/post_param", methods=["POST"])
def post_param():
    param = flask.request.form['param']
    if True:
        # <expect-error>
        with open(param, 'r') as fin:
            data = json.load(fin)
    return data

@app.route("/post_param", methods=["POST"])
def post_param_with_inline():
    # <expect-error>
    with open(flask.request.form['param'], 'r') as fin:
        data = json.load(fin)
    return data

@app.route("/post_param", methods=["POST"])
def post_param_with_inline_concat():
    # <expect-error>
    with open(flask.request.form['param'] + '.csv', 'r') as fin:
        data = json.load(fin)
    return data

@app.route("/subexpression", methods=["POST"])
def subexpression():
    param = "{}".format(flask.request.form['param'])
    print("do things")
    # <expect-error>
    return open(param, 'r').read()

@app.route("/ok")
def ok():
    # <no-error>
    open("static/path.txt", 'r')

def unsafe(request):
    filename = request.POST.get('filename')
    contents = request.POST.get('contents')
    print("something")
    # <expect-error>
    f = open(filename, 'r')
    f.close()

def unsafe_inline(request):
    # <expect-error>
    f = open(request.GET.get('filename'))
    f.write(request.POST.get('contents'))
    f.close()

def unsafe_dict(request):
    # <expect-error>
    f = open(request.POST['filename'])
    f.close()

def unsafe_with(request):
    filename = request.POST.get("filename")
    # <expect-error>
    with open(filename, 'r') as fin:
        data = fin.read()
    return HttpResponse(data)

def safe(request):
    # <no-error>
    filename_safe = "/tmp/data.txt"
    f = open(filename_safe)
    f.write("hello")
    f.close()

# Real-world finding
def download_doc(request):
    url = request.GET.get("url")
    format_doc = url.split(".")
    if format_doc[-1] == "docx":
        file_name = str(int(time.time())) + ".docx"
    else:
        file_name = str(int(time.time())) + ".xlsx"

    def file_iterator(_file, chunk_size=512):
        while True:
            c = _file.read(chunk_size)
            if c:
                yield c
            else:
                break

    # <expect-error>
    _file = open(url, "rb")
    response = StreamingHttpResponse(file_iterator(_file))
    response["Content-Type"] = "application/octet-stream"
    response["Content-Disposition"] = "attachment;filename=\"{0}\"".format(file_name)
    return response

class GenerateUserAPI(APIView):
    def get(self, request):
        """
        download users excel
        """
        file_id = request.GET.get("file_id")
        if not file_id:
            return self.error("Invalid Parameter, file_id is required")
        if not re.match(r"^[a-zA-Z0-9]+$", file_id):
            return self.error("Illegal file_id")
        file_path = f"/tmp/{file_id}.xlsx"
        if not os.path.isfile(file_path):
            return self.error("File does not exist")
        # <expect-error>
        with open(file_path, "rb") as f:
            raw_data = f.read()
        os.remove(file_path)
        response = HttpResponse(raw_data)
        response["Content-Disposition"] = f"attachment; filename=users.xlsx"
        response["Content-Type"] = "application/xlsx"
        return response
