import json
from flask import Flask

import blogpost
import fileupload
import dynamic


DEFAULT_SECRET_KEY = "skeleton"
DEFAULT_SESSION_TYPE = "filesystem"


def load_config(filename):
	with open(filename) as fh:
		return json.load(fh)


def create_app(config_file = 'config.json'):

	config = load_config(config_file)

	app = Flask(__name__)
	app.secret_key = config.get("secret_key", DEFAULT_SECRET_KEY)
	app.config['SESSION_TYPE'] = config.get("session_type", DEFAULT_SESSION_TYPE)

	forms = config.get("forms", {})
	for endpoint in forms:
		form = forms[endpoint]
		options = form.get("options", {})

		blueprint = None
		if "blogpost" == form["type"]:
			blueprint = blogpost.new(options)
		elif "fileupload" == form["type"]:
			blueprint = fileupload.new(options)
		elif "dynamic" == form["type"]:
			blueprint = dynamic.new(options)

		app.register_blueprint(blueprint, url_prefix=endpoint)

		return app

