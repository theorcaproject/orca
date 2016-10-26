from flask import render_template

from orcaweb import app

@app.route("/")
def route_dashboard():
    return render_template("dashboard.html")