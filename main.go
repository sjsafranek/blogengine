package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/schollz/httpfileserver"
	"github.com/sjsafranek/blogengine/lib/blogengine"
	"github.com/sjsafranek/lemur"
	"github.com/sjsafranek/logger"
)

const (
	DEFAULT_HTTP_PORT int = 8000
)

var (
	HTTP_PORT int  = DEFAULT_HTTP_PORT
	DEBUG     bool = false
)

const (
	MAJOR_VERSION int    = 0
	MINOR_VERSION int    = 1
	PATCH_VERSION int    = 3
	PROJECT_NAME  string = "Wiki Blog Server"
)

var PROJECT_FULL_NAME string = fmt.Sprintf("%v-%v.%v.%v", PROJECT_NAME, MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION)

func init() {
	flag.IntVar(&HTTP_PORT, "p", DEFAULT_HTTP_PORT, "http server port")
	flag.BoolVar(&DEBUG, "d", false, "Debug mode")
	flag.Parse()
}

func main() {

	logger.Debug(PROJECT_FULL_NAME)
	logger.Debug("GOOS: ", runtime.GOOS)
	logger.Debug("CPUS: ", runtime.NumCPU())
	logger.Debug("PID: ", os.Getpid())
	logger.Debug("Go Version: ", runtime.Version())
	logger.Debug("Go Arch: ", runtime.GOARCH)
	logger.Debug("Go Compiler: ", runtime.Compiler)
	logger.Debug("NumGoroutine: ", runtime.NumGoroutine())

	server, _ := lemur.NewServer()
	server.AttachFileServer("/static/", "static")

	directory, _ := filepath.Abs("content")

	var postTmpl *template.Template
	if !DEBUG {
		postTmpl = template.Must(template.ParseFiles("tmpl/page.html", "tmpl/header.html", "tmpl/footer.html"))
	}

	blog := &blogengine.BlogEngine{
		Directory:    directory,
		BasePath:     "/blog",
		// Template:     tmpl,
		// TemplateName: "post",

		Handler: func(w http.ResponseWriter, post *blogengine.Post) {
			var tmpl *template.Template
			if nil != postTmpl {
				tmpl = postTmpl
			} else {
				tmpl = template.Must(template.ParseFiles("tmpl/page.html", "tmpl/header.html", "tmpl/footer.html"))
			}
			err := tmpl.ExecuteTemplate(w, "page", post)
			if nil != err {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		},

	}
	server.AttachHandler("/blog", blog)

	// static files
	if DEBUG {
		server.AttachHandler("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	} else {
		server.AttachHandler("/static/", httpfileserver.New("/static/", "static"))
	}

	server.ListenAndServe(HTTP_PORT)
}
