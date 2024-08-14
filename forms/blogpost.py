import os
import datetime
from flask import Flask
from flask import flash
from flask import request
from flask import redirect
from flask import url_for
from flask import render_template

import conf


DEFAULT_UPLOAD_FOLDER = 'posts'


def make_post(form):
    tags = ', '.join([f'"{tag}"' for tag in form.getlist("tags")])
    timestamp = datetime.datetime.now().isoformat()
    return f'''---
title: "{form.get("title")}"
author: "{form.get("author")}"
type: ""
date: {timestamp}
subtitle: "{form.get("subtitle")}"
image: "{form.get("image")}"
tags: [{tags}]
categories: []
parties: []
campaigns: []
worlds: []
---

{form.get("content")}
'''


def create_app(upload_folder):

    if not upload_folder:
        upload_folder = DEFAULT_UPLOAD_FOLDER

    app = Flask(__name__)
    app.secret_key = conf.SECRET_KEY
    app.config['SESSION_TYPE'] =  conf.SESSION_TYPE
    app.config['UPLOAD_FOLDER'] = upload_folder

    @app.route('/', methods=['GET', 'POST'])
    def handler():
        if request.method == 'POST':
            print(request)
            print(request.form)
            post = make_post(request.form)
            print(post)
        return render_template("blog_post.html")

    return app
