import Flask

app = Flask(__name__)

def hello():
  app.run()

# <expect-error>
app.run()

# <expect-error>
app.run(debug=True)

if __name__ == '__main__':
    app.run()
