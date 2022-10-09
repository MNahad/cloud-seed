from flask import jsonify

def http(req):
    return jsonify({"status": "OK"})


def http_post(req):
    if req.method !== "POST":
        return "Bad Request", 400
    else:
        return "OK", 200
