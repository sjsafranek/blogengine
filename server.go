package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sjsafranek/blogengine/plugins/blogengine"
	"github.com/sjsafranek/lemur"
	"github.com/sjsafranek/logger"
)

const (
	DEFAULT_HTTP_PORT int = 8000
)

var (
	HTTP_PORT int = DEFAULT_HTTP_PORT
)

const (
	MAJOR_VERSION int    = 0
	MINOR_VERSION int    = 0
	PATCH_VERSION int    = 1
	PROJECT_NAME  string = "Server"
)

var PROJECT_FULL_NAME string = fmt.Sprintf("%v-%v.%v.%v", PROJECT_NAME, MAJOR_VERSION, MINOR_VERSION, PATCH_VERSION)

func init() {
	flag.IntVar(&HTTP_PORT, "p", DEFAULT_HTTP_PORT, "http server port")
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
	// template, _ := filepath.Abs("tmpl/page.html")
	tmpl := template.Must(template.ParseFiles("tmpl/page.html"))
	blog := &blogengine.BlogEngine{
		Directory:    directory,
		BasePath:     "/blog",
		Template:     tmpl,
		TemplateName: "post",
	}
	server.AttachHandler("/blog", blog)

	server.ListenAndServe(HTTP_PORT)
}
