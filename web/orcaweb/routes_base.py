from orcaweb import app


@app.route("/check")
def check():
    return "OK"