import subprocess
import sys
import shlex

def handler(event, context):
  # <no-error>
  subprocess.call("echo 'hello'")

  # <no-error>
  subprocess.call(["echo", "a", ";", "rm", "-rf", "/"])

  # <expect-error>
  subprocess.call("grep -R {} .".format(event['id']), shell=True)

  # <no-error>
  subprocess.call("grep -R {} .".format(event['id']))

  cmd = event['id'].split()
  # <no-error>
  subprocess.call([cmd[0], cmd[1], "some", "args"])

  # <expect-error>
  subprocess.call([cmd[0], cmd[1], "some", "args"], shell=True)

  # <expect-error>
  subprocess.call("grep -R {} .".format(event['id']), shell=True)

  # <expect-error>
  subprocess.call("grep -R {} .".format(event['id']), shell=True, cwd="/home/user")

  python_file = f"""
      print("What is your name?")
      name = input()
      print("Hello " + {event['id']})
  """
  # <no-error>
  program = subprocess.Popen(['python2', python_file], stdin=subprocess.PIPE, text=True)

  # <expect-error>
  program = subprocess.Popen(['python2', python_file], stdin=subprocess.PIPE, text=True, shell=True)

  # <expect-error>
  program = subprocess.Popen(['python2', python_file], stdin=subprocess.PIPE, shell=True, text=True)

  # <no-error>
  program = subprocess.Popen(['python2', shlex.split(python_file)], stdin=subprocess.PIPE, shell=True, text=True)

  program.communicate(input=payload, timeout=1)
