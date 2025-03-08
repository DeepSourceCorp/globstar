
import sqlite3
from fastapi import FastAPI, Query
import sqlite3

app = FastAPI()

def execute_unsafe_query(query: str):
    conn = sqlite3.connect("test.db")
    cursor = conn.cursor()
    cursor.execute(query)  #unsafe with user input
    result = cursor.fetchall()
    conn.commit()
    conn.close()
    return result

def better_query(query: str, params):
    conn = sqlite3.connect("test.db")
    cursor = conn.cursor()
    cursor.execute(query, params) #safe to execute with user input
    result = cursor.fetchall()
    conn.commit()
    conn.close()
    return result


@app.get("/unsafe_query/")
def unsafe_query(user_input: str):
    #f-string case
    #<expect-error>
    query = f"SELECT * FROM users WHERE name = {user_input}"
    #binary operator case
    #<expect-error>
    query2= "SELECT * FROM users WHERE name ="+ user_input
    #should not identify this as an error
    query3= "SELECT * FROM user WHERE name= ?"
    result = execute_unsafe_query(query)
    result2= execute_unsafe_query(query=query2)

    result3= better_query(query=query3, params=(user_input,))

    return {"result": result, "result2": result2, "result3": result3}
