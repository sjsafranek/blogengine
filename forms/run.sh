#flask --app 'blogpost:create_app(None)' run --reload --host=0.0.0.0 --port 1234 --debug

#flask --app 'fileupload:create_app(None)' run --reload --host=0.0.0.0 --port 1234 --debug

flask --app 'app:create_app("config.json")' run --reload --host=0.0.0.0 --port 1234 --debug