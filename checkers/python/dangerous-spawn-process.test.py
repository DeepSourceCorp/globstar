import os
import shlex
from somewhere import something

def handler(event, context):
    # <no-error>
    os.spawnlp(os.P_WAIT, "ls")

    # <no-error>
    os.spawnlpe(os.P_WAIT, "ls")

    # <no-error>
    os.spawnv(os.P_WAIT, "/bin/ls")

    # <no-error>
    os.spawnve(os.P_WAIT, "/bin/ls", ["-a"], os.environ)

    # <expect-error>
    os.spawnlp(os.P_WAIT, event['cmd'])

    # <expect-error>
    os.spawnlpe(os.P_WAIT, event['cmd'])

    # <expect-error>
    os.spawnv(os.P_WAIT, f"foo-{event['cmd']}")

    # <expect-error>
    os.spawnv(os.P_WAIT, "foo-%s" % event['cmd'])

    # <expect-error>
    os.spawnv(os.P_WAIT, "foo-{}".format(event['cmd']))

    # <expect-error>
    os.spawnve(os.P_WAIT, event['cmd'], ["-a"], os.environ)

    # <expect-error>
    os.spawnve(os.P_WAIT, "/bin/bash", ["-c", f"ls -la {event['cmd']}"], os.environ)

    # <expect-error>
    os.spawnl(os.P_WAIT, "/bin/bash", "-c", f"ls -la {event['cmd']}")

    # <expect-error>
    os.spawnl(os.P_WAIT, "/bin/bash", "-c", event['cmd'])
