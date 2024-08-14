import os
from flask import Flask
from flask import flash
from flask import request
from flask import redirect
from flask import url_for
from flask import render_template
from flask import Blueprint
from werkzeug.utils import secure_filename

import conf


DEFAULT_UPLOAD_FOLDER = 'uploads'
ALLOWED_EXTENSIONS = {'txt', 'pdf', 'png', 'jpg', 'jpeg', 'gif'}


fileupload = Blueprint('fileupload', __name__)


def allowed_file(filename):
    return '.' in filename and \
           filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS


def new(upload_folder):

    if not upload_folder:
        upload_folder = DEFAULT_UPLOAD_FOLDER
    
    blogpost = Blueprint('fileupload', __name__)

    @blogpost.route('/', methods=['GET', 'POST'])
    def handler():
        if request.method == 'POST':
            # check if the post request has the file part
            if 'file' not in request.files:
                print('No file part')
                flash('No file part')
                return redirect(request.url)
            file = request.files['file']
            # If the user does not select a file, the browser submits an
            # empty file without a filename.
            if file.filename == '':
                print('No selected file')
                flash('No selected file')
                return redirect(request.url)
            if file and allowed_file(file.filename):
                print('downloading file')
                filename = secure_filename(file.filename)
                file.save(os.path.join(upload_folder, filename))
                flash("File uploaded")
        return render_template("upload_file.html")

    return blogpost
