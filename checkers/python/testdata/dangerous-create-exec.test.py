import asyncio

class AsyncEventLoop:
    def __enter__(self):
        self.loop = asyncio.new_event_loop()
        asyncio.set_event_loop(self.loop)
        return self.loop

    def __exit__(self, *args):
        self.loop.close()


def handler_func1(event, context):
    args = event['cmds']
    program = args[0]
    with AsyncEventLoop() as loop:
        # <expect-error>
        proc = loop.run_until_complete(asyncio.subprocess.create_subprocess_exec(program, *args))
        loop.run_until_complete(proc.communicate())

def handler_func2(event, context):
    program = "bash"
    with AsyncEventLoop() as loop:
        # <expect-error>
        proc = loop.run_until_complete(asyncio.subprocess.create_subprocess_exec(program, event['cmd']))
        loop.run_until_complete(proc.communicate())


def handler_func3(event, context):
    program = "sh"
    eventcmd = event['cmd']
    with AsyncEventLoop() as loop:
        # <expect-error>
        proc = loop.run_until_complete(asyncio.subprocess.create_subprocess_exec(program, "bash", "-c", eventcmd))

def handler_func4(event, context):
    program = "sh"
    with AsyncEventLoop() as loop:
        # <expect-error>
        proc = loop.run_until_complete(asyncio.subprocess.create_subprocess_exec(program, "bash", "-c", event['cmd']))

def handler_func5(event, context):
    program = "sh"
    args = event['cmds']
    with AsyncEventLoop() as loop:
        # <expect-error>
        proc = loop.run_until_complete(asyncio.subprocess.create_subprocess_exec(program, ["bash", "-c", args]))

def handler_func6(event, context):
    program = "sh"
    with AsyncEventLoop() as loop:
        # <expect-error>
        proc = loop.run_until_complete(asyncio.subprocess.create_subprocess_exec(program, ["bash", "-c", event['cmds']]))

def ok_handler1(event, context):
    program = "echo"
    loop = asyncio.new_event_loop()
    # <no-error>
    proc = loop.run_until_complete(asyncio.subprocess.create_subprocess_exec(program, [program, "123"]))
    loop.run_until_complete(proc.communicate())

def ok_handler2(event, context):
    program = "sudo"
    ok_cmd2 = "cat"
    with AsyncEventLoop() as loop:
        # <no-error>
        proc = loop.run_until_complete(asyncio.subprocess.create_subprocess_exec(program, "bash", "-c", ok_cmd2))

def ok_handler3(event, context):
    program = "sudo"
    ok_cmd3 = "rm -rf /"
    with AsyncEventLoop() as loop:
        # <no-error>
        proc = loop.run_until_complete(subprocess.create_subprocess_exec(program, ok_cmd3))