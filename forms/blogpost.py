import os
import datetime
from flask import Flask
from flask import flash
from flask import request
from flask import redirect
from flask import url_for
from flask import render_template
from flask import Blueprint


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


def new(options={}):
    upload_folder = options.get("upload_folder", DEFAULT_UPLOAD_FOLDER)

    blogpost = Blueprint('blogpost', __name__)

    @blogpost.route('/', methods=['GET', 'POST'])
    def handler():
        if request.method == 'POST':
            print(request)
            print(request.form)
            post = make_post(request.form)
            print(post)
        return render_template("blog_post.html")

    return blogpost
