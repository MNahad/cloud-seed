from flask import jsonify


def start(req):
    request_json = req.get_json(silent=True)
    print("start...")
    print(request_json)
    return jsonify({"counter": 5})


def loop_entry(req):
    request_json = req.get_json(silent=True)
    print("entering loop...")
    print(request_json)
    if request_json is None:
        request_json = {"counter": 5}
    else:
        request_json["counter"] -= 1
    return jsonify(request_json)
