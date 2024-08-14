from flask import Flask

import conf
import blogpost
import fileupload


app = Flask(__name__)
app.secret_key = conf.SECRET_KEY
app.config['SESSION_TYPE'] =  conf.SESSION_TYPE

app.register_blueprint(blogpost.new(conf.BLOG_FOLDER), url_prefix="/create/post")
app.register_blueprint(fileupload.new(conf.FILE_FOLDER), url_prefix="/upload/file")

