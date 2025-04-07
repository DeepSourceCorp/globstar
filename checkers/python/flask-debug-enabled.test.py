from flask import Flask

app = Flask(__name__)

@app.route('/')
def index():
    return flask.jsonify({"response": "ok"})

def main():
    # <no-error>
    app.run()

def env():
    # <no-error>
    app.run("0.0.0.0", debug=os.environ.get("DEBUG", False))

if __name__ == "__main__":
    # <expect-error>
    app.run("0.0.0.0", debug=True)
