import json
import secret_info
import mysql.connector

RemoteMysql = secret_info.RemoteMysql

mydb = mysql.connector.connect(host=RemoteMysql.host, user=RemoteMysql.user, passwd=RemoteMysql.passwd, database=RemoteMysql.database)
mydbCursor = mydb.cursor()

def lambda_handler(event, context):
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
