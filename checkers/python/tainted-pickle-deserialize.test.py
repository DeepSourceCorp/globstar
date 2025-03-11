import _pickle
import cPickle
from dill import loads
import shelve


def lambda_handler(event, context):

  # <expect-error>
  _pickle.load(event['exploit_code'])

  # <expect-error>
  obj = cPickle.loads(f"foobar{event['exploit_code']}")

  # <expect-error>
  loads(event['exploit_code'])(123)

  # <expect-error>
  with shelve.open(f"/tmp/path/{event['object_path']}") as db:
    db['eggs'] = 'eggs'

  # <no-error>
  _pickle.loads('hardcoded code')

  # <no-error>
  code = '/file/path'
  cPickle.load(code)

  # <no-error>
  name = 'foobar'
  shelve.open(f"/tmp/path/{name}")
