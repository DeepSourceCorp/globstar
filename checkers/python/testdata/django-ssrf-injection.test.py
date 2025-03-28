def test_bad_1():
    from requests import get
    from django.shortcuts import render

    def send_to_redis(request):
        bucket = request.GET.get("bucket")
        # <expect-error>
        inner_response = get("http://my.redis.foo/{}".format(bucket), data=3)
        return render({"response_code": inner_response.status_code})

def test_bad_2():
    from requests import get
    from django.http import HttpResponse

    def send_to_redis(request):
        bucket = request.GET.get("bucket")
        # <expect-error>
        inner_response = get("http://my.redis.foo/{}".format(bucket), data=3)
        return HttpResponse(body = {"response_code": inner_response.status_code})

def test_bad_3():
    from requests import get
    from django.shortcuts import render

    def send_to_redis(request):
        bucket = request.GET.get("bucket")
        # <expect-error>
        inner_response = get(f"http://my.redis.foo/{bucket}", data=3)
        return render({"response_code": inner_response.status_code})

def test_bad_4():
    from requests import get
    from django.shortcuts import render

    def send_to_redis(request):
        bucket = request.headers.get("bucket")
        # <expect-error>
        inner_response = get("http://my.redis.foo/{}".format(bucket), data=3)
        return render({"response_code": inner_response.status_code})

def test_bad_5():
    from requests import get
    from django.shortcuts import render

    def send_to_redis(request):
        bucket = request.GET["bucket"]
        # <expect-error>
        inner_response = get("http://my.redis.foo/{}".format(bucket), data=3)
        return render({"response_code": inner_response.status_code})

def test_bad_6():
    from requests import get
    from django.shortcuts import render

    def send_to_redis(request):
        bucket = request.headers["bucket"]
        # <expect-error>
        inner_response = get("http://my.redis.foo/{}".format(bucket), data=3)
        return render({"response_code": inner_response.status_code})

def test_bad_7():
    from requests import get
    from django.shortcuts import render

    def send_to_redis(request):
        bucket = request.headers["bucket"]
        interm7 = "http://my.redis.foo/{}".format(bucket)
        # <expect-error>
        inner_response = get(interm7, data=3)
        return render({"response_code": inner_response.status_code})
    


import flask
import requests


app = flask.Flask(__name__)

@app.route("/route_param/<route_param>")
def route_param(broute_param):
    print("blah")
    # <expect-error>
    return requests.get(broute_param)

@app.route("/route_param_ok/<route_param>")
def route_param_ok(route_param):
    print("blah")
    # <no-error>
    return requests.get("this is safe")

@app.get("/route_param/<route_param>")
def route_param_without_decorator(route_param):
    print("blah")
    # <expect-error>
    return requests.get(route_param)

@app.route("/get_param", methods=["GET"])
def get_param():
    param = flask.request.args.get("param")
    # <expect-error>
    requests.post(param, timeout=10)

@app.route("/get_param_ok", methods=["GET"])
def get_param_ok():
    param = flask.request.args.get("param")
    # <no-error>
    requests.post("this is safe", timeout=10)

@app.route("/get_param_inline_concat", methods=["GET"])
def get_param_inline_concat():
    # <expect-error>
    requests.get(flask.request.args.get("param") + "/id")

@app.route("/get_param_concat", methods=["GET"])
def get_param_concat():
    param = flask.request.args.get("param")
    # <expect-error>
    requests.get(param + "/id")

@app.route("/get_param_format", methods=["GET"])
def get_param_format():
    param = flask.request.args.get("param")
    # <expect-error>
    requests.get("{}.csv".format(param))

@app.route("/get_param_percent_format", methods=["GET"])
def get_param_percent_format():
    param = flask.request.args.get("param")
    # <expect-error>
    requests.get("%s/id" % (param,))

@app.route("/post_param", methods=["POST"])
def post_param():
    param = flask.request.form['param']
    if True:
        # <expect-error>
        requests.get(param)

@app.route("/subexpression", methods=["POST"])
def subexpression():
    param = "{}".format(flask.request.form['param'])
    print("do things")
    # <expect-error>
    requests.post(param, data={"hello", "world"})

@app.route("/ok")
def ok():
    requests.get("https://www.google.com")

class GitlabApi(ScmApiBase):
    @cachedmethod("cache")
    @handle_errors
    @tracer_wrap
    def get_file(self, repo_name: str, commit_sha: str, file_path: str) -> str:
        api_url = (
            f"{self.base_url}/projects/{quote(repo_name, safe='')}/repository/files"
        )
        params = {"ref": commit_sha, "file_path": file_path}

        # <no-error>
        response = requests.get(api_url, headers=self.headers, params=params)
        code = response.json()["content"]
        code = code.encode("utf-8").decode("base64").decode("utf-8")
        return code                                         
