from django.db.models.expressions import RawSQL
from django.http import HttpResponse

##### RawSQL() True Positives #########
def get_user_age(request):
  user_name = request.get('user_name')
  # <expect-error>
  user_age = RawSQL('SELECT user_age FROM myapp_person where user_name = %s' % user_name)
  html = "<html><body>User Age %s.</body></html>" % user_age
  return HttpResponse(html)

def get_user_age(request):
  user_name = request.get('user_name')
  # <expect-error>
  user_age = RawSQL(f'SELECT user_age FROM myapp_person where user_name = {user_name}')
  html = "<html><body>User Age %s.</body></html>" % user_age
  return HttpResponse(html)

def get_user_age(request):
  user_name = request.get('user_name')
  # <expect-error>
  user_age = RawSQL('SELECT user_age FROM myapp_person where user_name = %s'.format(user_name))
  html = "<html><body>User Age %s.</body></html>" % user_age
  return HttpResponse(html)

def get_users(request):
  client_id = request.headers.get('client_id')
  # <expect-error>
  users = RawSQL('SELECT * FROM myapp_person where client_id = %s'.format(client_id))
  html = "<html><body>Users %s.</body></html>" % users
  return HttpResponse(html)

def get_users(request):
  client_id = request.headers.get('client_id')
  # <expect-error>
  users = RawSQL(f'SELECT * FROM myapp_person where client_id = {client_id}')
  html = "<html><body>Users %s.</body></html>" % users
  return HttpResponse(html)

def get_users(request):
  client_id = request.headers.get('client_id')
  interm = f'SELECT * FROM myapp_person where client_id = {client_id}'
  # <expect-error>
  users = RawSQL(interm)
  html = "<html><body>Users %s.</body></html>" % users
  return HttpResponse(html)

##### raw() True Negatives #########
def get_users(request):
  client_id = request.headers.get('client_id')
  # using param list is ok
  users = RawSQL('SELECT * FROM myapp_person where client_id = %s', (client_id,))
  html = "<html><body>Users %s.</body></html>" % users
  return HttpResponse(html)

##### True Positives #########
def fetch_name_0(request):
  with connection.cursor() as cursor:
      # <expect-error>
      cursor.execute(f"SELECT foo FROM bar WHERE baz = {request.data.get('baz')}")
      # <expect-error>
      cursor.execute("SELECT foo FROM bar WHERE baz = %s" % request.data.get('baz'))
      # <expect-error>
      cursor.execute("SELECT foo FROM bar WHERE baz = %s".format(request.data.get('baz')))
      row = cursor.fetchone()
  return row

def fetch_name_1(request):
  baz = request.data.get("baz")
  with connection.cursor() as cursor:
      # <expect-error>
      cursor.execute(f"UPDATE bar SET foo = 1 WHERE baz = {baz}")
      # <expect-error>
      cursor.execute(f"SELECT foo FROM bar WHERE baz = {baz}")
      row = cursor.fetchone()
  return row

def fetch_name_2(request):
  baz = request.data.get("baz")
  with connection.cursor() as cursor:
      # <expect-error>
      cursor.execute("SELECT foo FROM bar WHERE baz = %s" % baz)
      row = cursor.fetchone()
  return row

def fetch_name_3(request):
  baz = request.data.get("baz")
  with connection.cursor() as cursor:
      # <expect-error>
      cursor.execute("SELECT foo FROM bar WHERE baz = %s".format(baz))
      row = cursor.fetchone()
  return row

def upload(request, project_id):

    if request.method == 'POST':

        proj = Project.objects.get(pk=project_id)
        form = ProjectFileForm(request.POST, request.FILES)

        if form.is_valid():
            # Dependent on feature in develop
            name = request.POST.get('name', False)
            upload_path = store_uploaded_file(name, request.FILES['file'])

            other_name = "{}".format(name)
            curs = connection.cursor()
            # <expect-error>
            curs.execute(
                "insert into taskManager_file ('name','path','project_id') values ('%s','%s',%s)" %
                (other_name, upload_path, project_id))


##### True Negatives #########
def fetch_name_4(request):
  # using param list is ok
  baz = request.data.get("baz")
  with connection.cursor() as cursor:
      cursor.execute("UPDATE bar SET foo = 1 WHERE baz = %s", [baz])
      cursor.execute("SELECT foo FROM bar WHERE baz = %s", [baz])
      row = cursor.fetchone()

  return row

class Person(models.Model):
    first_name = models.CharField(...)
    last_name = models.CharField(...)
    birth_date = models.DateField(...)

##### raw() True Positives #########
def get_user_age(request):
  user_name = request.get('user_name')
  # <expect-error>
  user_age = Person.objects.raw('SELECT user_age FROM myapp_person where user_name = %s' % user_name)
  html = "<html><body>User Age %s.</body></html>" % user_age
  return HttpResponse(html)

def get_user_age(request):
  user_name = request.get('user_name')
  # <expect-error>
  user_age = Person.objects.raw(f"SELECT user_age FROM myapp_person where user_name = {user_name}")
  html = "<html><body>User Age %s.</body></html>" % user_age
  return HttpResponse(html)

def get_user_age(request):
  user_name = request.get('user_name')
  # <expect-error>
  user_age = Person.objects.raw('SELECT user_age FROM myapp_person where user_name = %s'.format(user_name))
  html = "<html><body>User Age %s.</body></html>" % user_age
  return HttpResponse(html)

def get_users(request):
  client_id = request.headers.get('client_id')
  # <expect-error>
  users = Person.objects.raw('SELECT * FROM myapp_person where client_id = %s' % client_id)
  html = "<html><body>Users %s.</body></html>" % users
  return HttpResponse(html)

def get_users(request):
  client_id = request.headers.get('client_id')
  intermVar = f'SELECT * FROM myapp_person where client_id = {client_id}'
  # <expect-error>
  users = Person.objects.raw(intermVar)
  html = "<html><body>Users %s.</body></html>" % users
  return HttpResponse(html)

##### raw() True Negatives #########
def get_user_age(request):
  user_name = request.get('user_name')
  # django queryset is good
  user_age = Person.objects.filter(user_name=user_name).first()
  html = "<html><body>User Age %s.</body></html>" % user_age
  return HttpResponse(html)

def get_users(request):
  client_id = request.headers.get('client_id')
  # using param list is ok
  users = Person.objects.raw('SELECT * FROM myapp_person where client_id = %s', (client_id,))
  html = "<html><body>Users %s.</body></html>" % users
  return HttpResponse(html)
