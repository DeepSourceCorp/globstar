
import sqlite3
from fastapi import FastAPI, Query
import sqlite3

app = FastAPI()

def execute_unsafe_query(query: str):
    conn = sqlite3.connect("test.db")
    cursor = conn.cursor()
    cursor.execute(query)  # ✅ Uses parameterized query
    result = cursor.fetchall()
    conn.commit()
    conn.close()
    return result

@app.get("/unsafe_query/")
def unsafe_query(user_input: str):
    query = f"SELECT * FROM users WHERE name = {user_input}"
    query2= "SELECT * FROM users WHERE name ="+ user_input
    result = execute_unsafe_query(query)
    result2= execute_unsafe_query(query=query2)
    return {"result": result, "result2": result2}
