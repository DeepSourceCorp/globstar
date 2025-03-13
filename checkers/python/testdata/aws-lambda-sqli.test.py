import json
import secret_info
import mysql.connector
import psycopg2
import pymssql
import boto3
import pymysql
import sqlalchemy
from aws_lambda_powertools import Logger, Tracer
from aws_lambda_powertools.utilities.typing import LambdaContext

RemoteMysql = secret_info.RemoteMysql

mydb = mysql.connector.connect(host=RemoteMysql.host, user=RemoteMysql.user, passwd=RemoteMysql.passwd, database=RemoteMysql.database)
mydbCursor = mydb.cursor()

def lambda_handler1(event, context):
    publicIP=event["queryStringParameters"]["publicIP"]
    sql1 = """UPDATE `EC2ServerPublicIP` SET %s = '%s' WHERE %s = %d""" % ("publicIP",publicIP,"ID", 1)

    # <expect-error>
    mydbCursor.executemany("""UPDATE `EC2ServerPublicIP` SET %s = '%s' WHERE %s = %d""" % ("publicIP",publicIP,"ID", 1))

    # <expect-error>
    mydbCursor.execute(sql1)

    sql2 = """UPDATE `EC2ServerPublicIP` SET {column} = '{value}' WHERE {condition_column} = {condition_value}""".format(
    column="publicIP",
    value=publicIP,
    condition_column="ID",
    condition_value=1)

    # <expect-error>
    mydbCursor.execute(sql2)

    # <expect-error>
    mydbCursor.execute(f"""UPDATE `EC2ServerPublicIP` SET publicIP = '{publicIP}' WHERE ID = {1}""")

    # <no-error>
    mydbCursor.execute("UPDATE `EC2ServerPublicIP` SET %s = '%s' WHERE %s = %s", ("publicIP",publicIP,"ID", 1))
    mydb.commit()
    
    Body={
        "publicIP":publicIP
        
    }
    return {
        'statusCode': 200,
        'body': json.dumps(Body)
    }


def lambda_handler2(event, context):
    ssm = boto3.client('ssm')
    database = ssm.get_parameter(Name = 't2-db-dbname')
    user = ssm.get_parameter(Name = 't2-db-user')
    port = ssm.get_parameter(Name = 't2-db-port')
    tableName = ssm.get_parameter(Name = 't2-db-tablename')
    password = ssm.get_parameter(Name = 't2-db-password', WithDecryption = True)
    host = ssm.get_parameter(Name = 't2-db-host', WithDecryption = True)

    engine = psycopg2.connect(
    database=database['Parameter']['Value'],
    user=user['Parameter']['Value'],
    password=password['Parameter']['Value'],
    host=host['Parameter']['Value'],
    port=port['Parameter']['Value']
    )
    tableName = tableName['Parameter']['Value']

    keyphrase = event['keyphrase']
    username = event['username']
    language = event['translateTarget']

    cur = conn.cursor()
    findQuery = '''SELECT file_name FROM {tableName} WHERE '{keyphrase}' = ANY (keyphrases) AND target_language = '{language}' AND username = '{username}' '''.format(username=username, keyphrase=keyphrase, language=language, tableName = tableName)
    # <expect-error>
    cur.execute(findQuery)
    result = cur.fetchone()
    returnList = []

    # <no-error>
    cur.execute("SELECT * FROM foobar WHERE id = '%s'", username)

    if (result is None):
        returnList.append('None')
    else:
        for i in range (0,len(result)):
            returnList.append(result[i])
            
    response =  {
        'searchedFiles':returnList,
        'language' : language
        }

    engine.commit()
    engine.close()
    
    return response 


def lambda_handler3(event, context):
    current_user = event['user_id']
    secret_dict = get_secret_dict()

    port = str(secret_dict['port']) if 'port' in secret_dict else '1433'
    dbname = secret_dict['dbname'] if 'dbname' in secret_dict else 'master'
    conn = pymssql.connect(server=secret_dict['host'],
                            user=secret_dict['username'],
                            password=secret_dict['password'],
                            database=dbname,
                            port=port,
                            login_timeout=5,
                            as_dict=True)
    cursor = conn.cursor(as_dict=True)

    query = "SELECT roleprin.name FROM sys.database_role_members rolemems "\
            "JOIN sys.database_principals roleprin ON roleprin.principal_id = rolemems.role_principal_id "\
            "JOIN sys.database_principals userprin ON userprin.principal_id = rolemems.member_principal_id "\
            "WHERE userprin.name = '%s'" % current_user
    # <expect-error>
    cursor.execute(query)

    # <no-error>
    cursor.execute("SELECT * FROM user WHERE id ='%s'", current_user)
    return {
        'statusCode': 200,
        'body': 'ok'
    }

def lambda_handler(event, context):
    status_code = 400
    try:
        user_id = event['requestContext']['identity']['cognitoIdentityId']
        sql = '''
                SELECT
                id
                ,userId
                ,stationId
                ,stationName
                ,duration
                ,price
                ,createdDate
                FROM
                rideTransactions
                WHERE
                userId = "{}"
                ORDER BY
                createdDate DESC;
                '''.format(user_id)

        conn_info = connection_info(DB_CREDS)
        conn = pymysql.connect(host=conn_info['host'], user=conn_info['username'], password=conn_info['password'], database=conn_info['dbname'], connect_timeout=30, cursorclass=pymysql.cursors.DictCursor)
        with conn.cursor() as cur:
            # <expect-error>
            cur.execute(sql)
            rows = cur.fetchall()

            # <no-error>
            cur.execute('SELECT * FROM foobar')
            rows2 = cur.fetchall()
        conn.close()
        output = [{'id': c['id'], 'userId': c['userId'], 'stationId': c['stationId'], 'stationName': c['stationName'], 'duration': c['duration'], 'price': float(c['price']), 'createdDate': c['createdDate'].isoformat()} for c in rows]
        status_code = 200

    except Exception as e:
        print('ERROR: ', e)
        output = '{}'.format(e)

    response = {
    'statusCode': status_code,
    'headers': {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Credentials': True,
        'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent',
        'Access-Control-Allow-Methods': 'GET,POST,PUT,DELETE,OPTIONS,HEAD,PATCH',
        'Content-Type': 'application/json'
    },
    'body': json.dumps(output)
    }

    print('[INFO] Query response: {}'.format(json.dumps(response)))

    return response
a = "some_string"
SERVICE = "lambda-connection-pooling-demo"
logger = Logger(service=SERVICE)
tracer = Tracer(service=SERVICE)

class LambdaProxyIntegrationResponse(t.TypedDict, total=False):
    statusCode: int
    body: str
    headers: t.Dict[str, t.Any]

DB_USER_SECRET_NAME = os.environ.get("DB_USER_SECRET_NAME")
DB_HOST = os.environ.get("DB_HOST")

secrentsmanager = boto3.client(service_name='secretsmanager')
get_secret_value_response = secrentsmanager.get_secret_value(SecretId=DB_USER_SECRET_NAME)
secret = json.loads(get_secret_value_response["SecretString"])
db_user = secret["username"]
db_password = secret["password"]
db_host = DB_HOST or secret["host"]
db_port = secret["port"]

url = f"mysql+pymysql://{db_user}:{db_password}@{db_host}:{db_port}/"
engine = sqlalchemy.create_engine(
    url,
    connect_args={
        "ssl": {
            "ssl_ca": "./AmazonRootCA1.pem",
        }
    },
    pool_recycle=50,
)

def handler(event, context):
    logger.debug("connecting to db...")
    with engine.connect() as connection:
        try:
            # <expect-error>
            connection.execute(f"SELECT * FROM foobar WHERE id = '{event['id']}'")
            # <no-error>
            connection.execute("SELECT * FROM foobar WHERE id = '?'", event['id'])
        except Exception as e:
            logger.error("An error occured:")
            print(e)
            return {
                "statusCode": 200,
                "body": json.dumps({
                    "state": "ERROR",
                    "message": f"response from '{context.log_stream_name}'"
                })
            }
    return {
        "statusCode": 200,
        "body": json.dumps({
            "state": "SUCCESS",
            "message": f"response from '{context.log_stream_name}'"
        })
    }