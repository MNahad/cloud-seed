from flask import jsonify


def loop_check(req):
    request_json = req.get_json(silent=True)
    print("checking...")
    print(request_json)
    return jsonify(request_json)


def final(req):
    request_json = req.get_json(silent=True)
    print("end...")
    print(request_json)
    return jsonify(request_json)
