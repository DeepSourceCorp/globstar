import os
import flask
import hashlib

def danger(request):
    url = request.GET['url']
    # <expect-error>
    os.system('wget ' + url)

def danger2(request):
    image = request.POST['image']
    # <expect-error>
    os.system("./face-recognize %s --N 24" % image)

def danger3(request):
    url = request.GET['url']
    # <expect-error>
    os.system("nslookup " + url)

def ok(request):
    url = request.GET['url']
    # <no-error>
    os.system("echo 'hello'")

app = flask.Flask(__name__)

@app.route("/route_param/<route_param>")
def route_param(route_param):
    print("blah")
    # <expect-error>
    return os.system(route_param)

@app.route("/route_param_ok/<route_param>")
def route_param_ok(route_param):
    print("blah")
    # <no-error>
    return os.system("ls -la")

@app.route("/route_param_concat/<route_param>")
def route_param_concat(route_param):
    print("blah")
    # <expect-error>
    return os.system("echo " + route_param)

@app.route("/route_param_format/<route_param>")
def route_param_format(route_param):
    print("blah")
    # <expect-error>
    return os.system("echo {}".format(route_param))

@app.route("/route_param_percent_format/<route_param>")
def route_param_percent_format(route_param):
    print("blah")
    # <expect-error>
    return os.system("echo %s" % route_param)

@app.route("/get_param_inline", methods=["GET"])
def get_param_inline():
    # <expect-error>
    os.system(flask.request.args.get("param"))

@app.route("/get_param_inline_concat", methods=["GET"])
def get_param_inline_concat():
    # <expect-error>
    os.system("echo " + flask.request.args.get("param"))

@app.route("/get_param", methods=["GET"])
def get_param():
    param = flask.request.args.get("param")
    # <expect-error>
    os.system(param)

@app.route("/get_param_concat", methods=["GET"])
def get_param_concat():
    param = flask.request.args.get("param")
    # <expect-error>
    os.system("echo " + param)

@app.route("/get_param_format", methods=["GET"])
def get_param_format():
    param = flask.request.args.get("param")
    # <expect-error>
    os.system("echo {}".format(param))

@app.route("/get_param_percent_format", methods=["GET"])
def get_param_percent_format():
    param = flask.request.args.get("param")
    # <expect-error>
    os.system("echo %s" % (param,))

@app.route("/post_param", methods=["POST"])
def post_param():
    param = flask.request.form['param']
    # <expect-error>
    os.system(param)

@app.route("/post_param_branch", methods=["POST"])
def post_param_branch():
    param = flask.request.form['param']
    if True:
        # <expect-error>
        os.system(param)

@app.route("/subexpression", methods=["POST"])
def subexpression():
    param = "{}".format(flask.request.form['param'])
    print("do things")
    # <expect-error>
    os.system(param)

@app.route("/subexpression_concat", methods=["POST"])
def subexpression_concat():
    param = "{}".format(flask.request.form['param'])
    print("do things")
    # <expect-error>
    os.system("echo " + param)

@app.route("/subexpression_format", methods=["POST"])
def subexpression_format():
    param = "{}".format(flask.request.form['param'])
    print("do things")
    # <expect-error>
    os.system("echo {}".format(param))

@app.route("/subexpression_percent_format", methods=["POST"])
def subexpression_percent_format():
    param = "{}".format(flask.request.form['param'])
    print("do things")
    # <expect-error>
    os.system("echo %s" % param)

# Real world example
@app.route('/', methods=['GET', 'POST'])
def index():
    if flask.request.method == 'GET':
        return flask.render_template('index.html')
    # check url first
    url = flask.request.form.get('url', None)
    if url != '':
        md5 = hashlib.md5(url+app.config['MD5_SALT']).hexdigest()
        fpath = join(join(app.config['MEDIA_ROOT'], 'upload'), md5+'.jpg')
        # <expect-error>
        r = os.system('wget %s -O "%s"'%(url, fpath))
        if r != 0: abort(403)
        return flask.redirect(flask.url_for('landmark', hash=md5))

@app.route("/ok")
def ok():
    # <no-error>
    os.system("This is fine")
