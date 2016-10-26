from flask.app import Flask

app = Flask(__name__)
app.config.from_envvar('APP_CONFIG_FILE')
app.secret_key = "secret"

with app.test_request_context():
    import routes_base
    import routes_home
