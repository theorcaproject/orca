import re
import uuid

import flask
from flask import render_template
from flask import request

from orcaweb import app
from orcaweb.services import api__trainer


@app.route("/")
def route_dashboard():
    return render_template("dashboard.html", configuration=api__trainer.get__configuration__applications())


@app.route("/application/<name>", methods=["GET"])
def route_application(name):
    return render_template("application.html", application=api__trainer.get__configuration__applications_app(name))


def extract_command(string_command):
    p = re.compile('([a-z]+)\s(.*)')
    m = p.match(string_command)
    return m.group(1).strip(), m.group(2).strip()


@app.route("/application", methods=["POST"])
def route_application__add():
    app = {
        "Name": request.form.get("name"),
        "InstallCommands": [],
        "InstallFiles": [],
        "QueryStateCommand": [],
        "RemoveCommand": []

    }
    api__trainer.set__configuration__applications_app(app)
    return ""


@app.route("/application/<name>", methods=["POST"])
def route_application__set(name):
    existing_configuration = api__trainer.get__configuration__applications_app(name)
    existing_configuration["Type"] = request.form.get("Type")
    existing_configuration["Min"] = int(request.form.get("Min"))
    existing_configuration["Max"] = int(request.form.get("Max"))

    api__trainer.set__configuration__applications_app(existing_configuration)
    return flask.redirect("/application/" + name)


@app.route("/application/<appname>/files", methods=["POST"])
@app.route("/application/<appname>/files/<id>", methods=["POST"])
def route_application__set__create_file(appname, id=None):
    if not id:
        id = str(uuid.uuid4())

    name = request.form.get("name", "")
    contents = request.form.get("contents", "")
    existing_file = api__trainer.get__configuration__applications_app__installed_file(appname, id)

    # Attempt to find a file
    if not existing_file:
        existing_file = {
            "Id": id,
            "Type": "FILE_COMMAND",
            "Command": {
                "Path": "",
                "Args": ""
            }
        }

    existing_file["Command"]["Path"] = name
    existing_file["Command"]["Args"] = contents

    api__trainer.get__configuration__applications_app__installed_file_set(appname, id, existing_file)
    return flask.redirect("/application/" + appname)


@app.route("/application/<appname>/files/<id>", methods=["GET"])
def route_application__set__get_file(appname, id=None):
    existing_file = api__trainer.get__configuration__applications_app__installed_file(appname, id)
    return flask.jsonify(existing_file)


@app.route("/application/<appname>/files/<id>", methods=["DELETE"])
def route_application__set__delete_file(appname, id):
    api__trainer.get__configuration__applications_app__delete_file(appname, id)
    return ""


@app.route("/application/<appname>/healthchecks", methods=["POST"])
@app.route("/application/<appname>/healthchecks/<id>", methods=["POST"])
def route_application__set__create_healthchecks(appname, id=None):
    if not id:
        id = str(uuid.uuid4())

    path = request.form.get("path", "")
    args = request.form.get("args", "")
    existing_file = api__trainer.get__configuration__applications_app__healthchecks(appname, id)

    # Attempt to find a file
    if not existing_file:
        existing_file = {
            "Id": id,
            "Type": "EXEC_COMMAND",
            "Command": {
                "Path": "",
                "Args": ""
            }
        }

    existing_file["Command"]["Path"] = path
    existing_file["Command"]["Args"] = args

    api__trainer.get__configuration__applications_app__healthchecks_set(appname, id, existing_file)
    return flask.redirect("/application/" + appname)


@app.route("/application/<appname>/healthchecks/<id>", methods=["GET"])
def route_application__set__get_healthchecks(appname, id=None):
    existing_file = api__trainer.get__configuration__applications_app__healthchecks(appname, id)
    return flask.jsonify(existing_file)


@app.route("/application/<appname>/healthchecks/<id>", methods=["DELETE"])
def route_application__set__delete_healthchecks(appname, id):
    api__trainer.get__configuration__applications_app__delete_healthchecks(appname, id)
    return ""


@app.route("/application/<appname>/installcommands", methods=["POST"])
@app.route("/application/<appname>/installcommands/<id>", methods=["POST"])
def route_application__set__create_installcommands(appname, id=None):
    if not id:
        id = str(uuid.uuid4())

    path = request.form.get("path", "")
    args = request.form.get("args", "")
    existing_file = api__trainer.get__configuration__applications_app__installcommands(appname, id)

    # Attempt to find a file
    if not existing_file:
        existing_file = {
            "Id": id,
            "Type": "EXEC_COMMAND",
            "Command": {
                "Path": "",
                "Args": ""
            }
        }

    existing_file["Command"]["Path"] = path
    existing_file["Command"]["Args"] = args

    api__trainer.get__configuration__applications_app__installcommands_set(appname, id, existing_file)
    return flask.redirect("/application/" + appname)


@app.route("/application/<appname>/installcommands/<id>", methods=["GET"])
def route_application__set__get_installcommands(appname, id=None):
    existing_file = api__trainer.get__configuration__applications_app__installcommands(appname, id)
    return flask.jsonify(existing_file)


@app.route("/application/<appname>/installcommands/<id>", methods=["DELETE"])
def route_application__set__delete_installcommands(appname, id):
    api__trainer.get__configuration__applications_app__delete_installcommands(appname, id)
    return ""


@app.route("/application/<appname>/removecommands", methods=["POST"])
@app.route("/application/<appname>/removecommands/<id>", methods=["POST"])
def route_application__set__create_removecommands(appname, id=None):
    if not id:
        id = str(uuid.uuid4())

    path = request.form.get("path", "")
    args = request.form.get("args", "")
    existing_file = api__trainer.get__configuration__applications_app__removecommands(appname, id)

    # Attempt to find a file
    if not existing_file:
        existing_file = {
            "Id": id,
            "Type": "EXEC_COMMAND",
            "Command": {
                "Path": "",
                "Args": ""
            }
        }

    existing_file["Command"]["Path"] = path
    existing_file["Command"]["Args"] = args

    api__trainer.get__configuration__applications_app__removecommands_set(appname, id, existing_file)
    return flask.redirect("/application/" + appname)


@app.route("/application/<appname>/removecommands/<id>", methods=["GET"])
def route_application__set__get_removecommands(appname, id=None):
    existing_file = api__trainer.get__configuration__applications_app__removecommands(appname, id)
    return flask.jsonify(existing_file)


@app.route("/application/<appname>/removecommands/<id>", methods=["DELETE"])
def route_application__set__delete_removecommands(appname, id):
    api__trainer.get__configuration__applications_app__delete_removecommands(appname, id)
    return ""
