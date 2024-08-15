import os
from flask import Flask
from flask import flash
from flask import request
from flask import redirect
from flask import url_for
from flask import render_template
from flask import Blueprint
from werkzeug.utils import secure_filename


def new(options={}):

    print(options)
    print(options.get("submit"))
    
    blueprint = Blueprint('dynamic', __name__)

    @blueprint.route('/', methods=['GET', 'POST'])
    def handler():
        if request.method == 'POST':
            print(request)
            print(request.form)
        return render_template(
        	"dynamic.html", 
        	title=options.get("title", "dynamic"),
        	submit=options.get("submit", {"text": "Submit", "class": "btn-primary"}),
        	fields=options.get("fields", [])
        )

    return blueprint