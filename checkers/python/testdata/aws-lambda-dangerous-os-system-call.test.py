import os

def handler(event, context):
    # ok: dangerous-system-call
    os.system("ls -al")

    # ok: dangerous-system-call
    os.popen("cat contents.txt")

    # <expect-error>
    os.system(f"ls -la {event['dir']}")

    eventVar1 = event['cmd']
    eventFlag = event['flag']
    cmdstr = "ls -la {} {}"

    # <expect-error>
    os.popen2("sudo rm -rf {}".format(eventVar1))

    # <expect-error>
    os.system(cmdstr.format(eventFlag, eventVar1))

    intermVar = cmdstr % (eventVar1, eventFlag)

    # <expect-error>
    os.popen(intermVar)
