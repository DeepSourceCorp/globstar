"<h1> Hello {} </h1>".format(request.POST("user"))
"<h1> Hello {} </h1>".format(request.POST["user"])
"<h1> Hello %s </h1>" % request.POST("user")
"<h1> Hello %s </h1>" % request.POST["user"]
"<h1> Hello %s </h1>" + request.POST["user"]
f"<h1> Hello {request.POST["user"]} </h1>"
f"<h1> Hello {request.POST("user")} </h1>"

def f(A):
    link = "<h1> Hello {} </h1>"
    req = request.POST["user"]
    x = link.format(req)
