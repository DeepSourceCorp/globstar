def lambda_handler1(event, context):
    # <no-error>
    exec("x = 1; x = x + 2")

    blah1 = "import requests; r = requests.get('https://example.com')"
    # <no-error>
    exec(blah1)

    dynamic1 = "import requests; r = requests.get('{}')"
    # <expect-error>
    exec(dynamic1.format(event['url']))

    # ok:tainted-code-exec
    eval("x = 1; x = x + 2")

    blah2 = "import requests; r = requests.get('https://example.com')"
    # ok:tainted-code-exec
    eval(blah2)

    dynamic2 = "import requests; r = requests.get('{}')"
    # <expect-error>
    eval(dynamic2.format(event['url']))

    dynamic3 = "import requests; r = requests.het('%s')"
    # <expect-error>
    exec(dynamic3 % event['url'])

    dynamic4 = f"import requests; r = requests.get('{event['url']}')"
    # <expect-error>
    eval(dynamic4)


def lambda_handler2(event, context):
    # <no-error>
    exec("x = 1; x = x + 2")

    blah1 = "import requests; r = requests.get('https://example.com')"
    # <no-error>
    exec(blah1)

    dynamic1 = "import requests; r = requests.get('{}')"
    interm1 = dynamic1.format(event['url'])
    # <expect-error>
    exec(interm1)

    # <no-error>
    eval("x = 1; x = x + 2")

    blah2 = "import requests; r = requests.get('https://example.com')"
    # <no-error>
    eval(blah2)

    dynamic2 = "import requests; r = requests.get('{}')"
    interm2 = dynamic2.format(event['url'])
    # <expect-error>
    eval(interm2)

    dynamic3 = "import requests; r = requests.het('%s')"
    interm3 = dynamic3 % event['url']
    # <expect-error>
    exec(interm3)

    dynamic4 = "import requests; r = requests.get('{}')"
    eventVar = event['url']
    interm4 = dynamic4.format(eventVar)
    # <expect-error>
    exec(interm4)
