import json

import requests


def get__configuration__applications():
    response = requests.get("http://localhost:5001/state/config/applications")
    return json.loads(response.text)

def get__configuration__applications_app(name):
    response = requests.get("http://localhost:5001/state/config/applications")
    for application in json.loads(response.text):
        if application["Name"] == name:
            return application


def set__configuration__applications_app(object):
    requests.post("http://localhost:5001/state/config/applications", data=json.dumps(object))


def get__configuration__applications_app__installed_file(application, id):
    application = get__configuration__applications_app(application)
    for file in application["InstallFiles"]:
        if file["Id"] == id:
            return file


def get__configuration__applications_app__installed_file_set(application, id, existing_file):
    found = False
    application = get__configuration__applications_app(application)
    ret_files = []
    for file in application["InstallFiles"]:
        if file["Id"] == id:
            file = existing_file
            found = True
        ret_files.append(file)

    if not found:
        ret_files.append(existing_file)

    application["InstallFiles"] = ret_files
    set__configuration__applications_app(application)

def get__configuration__applications_app__delete_file(application, id):
    application = get__configuration__applications_app(application)
    ret_files = []
    for file in application["InstallFiles"]:
        if file["Id"] != id:
            ret_files.append(file)

    application["InstallFiles"] = ret_files
    set__configuration__applications_app(application)



def get__configuration__applications_app__healthchecks(application, id):
    application = get__configuration__applications_app(application)
    for file in application["QueryStateCommand"]:
        if file["Id"] == id:
            return file


def get__configuration__applications_app__healthchecks_set(application, id, existing_file):
    found = False
    application = get__configuration__applications_app(application)
    ret_files = []
    for file in application["QueryStateCommand"]:
        if file["Id"] == id:
            file = existing_file
            found = True
        ret_files.append(file)

    if not found:
        ret_files.append(existing_file)

    application["QueryStateCommand"] = ret_files
    set__configuration__applications_app(application)

def get__configuration__applications_app__delete_healthchecks(application, id):
    application = get__configuration__applications_app(application)
    ret_files = []
    for file in application["QueryStateCommand"]:
        if file["Id"] != id:
            ret_files.append(file)

    application["QueryStateCommand"] = ret_files
    set__configuration__applications_app(application)



def get__configuration__applications_app__installcommands(application, id):
    application = get__configuration__applications_app(application)
    for file in application["InstallCommands"]:
        if file["Id"] == id:
            return file


def get__configuration__applications_app__installcommands_set(application, id, existing_file):
    found = False
    application = get__configuration__applications_app(application)
    ret_files = []
    for file in application["InstallCommands"]:
        if file["Id"] == id:
            file = existing_file
            found = True
        ret_files.append(file)

    if not found:
        ret_files.append(existing_file)

    application["InstallCommands"] = ret_files
    set__configuration__applications_app(application)

def get__configuration__applications_app__delete_installcommands(application, id):
    application = get__configuration__applications_app(application)
    ret_files = []
    for file in application["InstallCommands"]:
        if file["Id"] != id:
            ret_files.append(file)

    application["InstallCommands"] = ret_files
    set__configuration__applications_app(application)


def get__configuration__applications_app__removecommands(application, id):
    application = get__configuration__applications_app(application)
    for file in application["RemoveCommand"]:
        if file["Id"] == id:
            return file


def get__configuration__applications_app__removecommands_set(application, id, existing_file):
    found = False
    application = get__configuration__applications_app(application)
    ret_files = []
    for file in application["RemoveCommand"]:
        if file["Id"] == id:
            file = existing_file
            found = True
        ret_files.append(file)

    if not found:
        ret_files.append(existing_file)

    application["RemoveCommand"] = ret_files
    set__configuration__applications_app(application)

def get__configuration__applications_app__delete_removecommands(application, id):
    application = get__configuration__applications_app(application)
    ret_files = []
    for file in application["RemoveCommand"]:
        if file["Id"] != id:
            ret_files.append(file)

    application["RemoveCommand"] = ret_files
    set__configuration__applications_app(application)