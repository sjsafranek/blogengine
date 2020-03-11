package main

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sjsafranek/lemur"
	"github.com/sjsafranek/logger"
)

// Encode to gob file
func SaveToGobFile(path string, object interface{}) error {
	file, err := os.Create(path)
	if nil != err {
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(object)
}

// Decode gob file
func LoadGobFromFile(path string, object interface{}) error {
	file, err := os.Open(path)
	if nil != err {
		return err
	}
	defer file.Close()
	return gob.NewDecoder(file).Decode(object)
}

type DataWrapper struct {
	Thread	*Thread `json:"thread,omitempty"`
	Threads []*Thread `json:"threads,omitempty"`
}

type ResponseWrapper struct {
	Status string `json:"status"`
	Data DataWrapper `json:"data"`
}

func (self *ResponseWrapper) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *ResponseWrapper) Write(w io.Writer) error {
	payload, err := self.Marshal()
	if nil != err {
		return err
	}
	_, err = fmt.Fprintln(w, payload)
	return err
}

type Thread struct {
	Id       string     `json:"id"`
	Author   string     `json:"author"`
	Content  string     `json:"content"`
	Time     int64      `json:"time"`
	Threads []*Thread  `json:"threads"`
	mux      sync.Mutex `json:"-"`
}

func (self *Thread) Add(thread *Thread) error {
	self.mux.Lock()
	self.Threads = append(self.Threads, thread)
	defer self.mux.Unlock()
	return nil
}

func (self *Thread) Get(threadId string) *Thread {
	for _, thread := range self.Threads {
		if threadId == thread.Id {
			return thread
		} else {
			thread = thread.Get(threadId)
			if nil != thread && threadId == thread.Id {
				return thread
			}
		}
	}
	return nil
}

// func (self *Thread) Marshal() (string, error) {
// 	b, err := json.Marshal(self)
// 	if nil != err {
// 		return "", err
// 	}
// 	return string(b), err
// }
//
// func (self *Thread) Write(w io.Writer) error {
// 	payload, err := self.Marshal()
// 	if nil != err {
// 		return err
// 	}
// 	_, err = fmt.Fprintln(w, payload)
// 	return err
// }

type BulletinBoard struct {
	filename string               `json:"file"`
	Threads     []*Thread   `json:"threads"`
	mux      sync.Mutex           `json:"-"`
}

// func (self *BulletinBoard) Marshal() (string, error) {
// 	b, err := json.Marshal(self)
// 	if nil != err {
// 		return "", err
// 	}
// 	return string(b), err
// }
//
// func (self *BulletinBoard) Write(w io.Writer) error {
// 	payload, err := self.Marshal()
// 	if nil != err {
// 		return err
// 	}
// 	_, err = fmt.Fprintln(w, payload)
// 	return err
// }

func (self *BulletinBoard) Get(threadId string) *Thread {
	for _, thread1 := range self.Threads {
		if thread1.Id == threadId {
			return thread1
		}
		thread2 := thread1.Get(threadId)
		if nil != thread2 {
			return thread2
		}
	}
	return nil
}

func (self *BulletinBoard) Add(thread *Thread) error {
	self.mux.Lock()
	defer self.mux.Unlock()
	// self.Threads[thread.Id] = thread
		self.Threads = append( self.Threads,  thread)
	return nil
}

// func (self *BulletinBoard) Has(threadId string) bool {
// 	self.mux.Lock()
// 	defer self.mux.Unlock()
// 	_, ok := self.Threads[threadId]
// 	return ok
// }

func (self *BulletinBoard) Save() error {
	return SaveToGobFile(self.filename, self)
}

func (self *BulletinBoard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadId := vars["threadId"]

	// create root thread
	if "" == threadId {
		if "POST" == r.Method {
			thread, err := self.createThreadFromHttpRequest(r)
			if nil != err {
				panic(err)
			}
			self.Add(thread)

			err = self.Save()
			if nil != err {
				logger.Error(err)
			}

			self.getThreadHttpHandler(w, r, thread)
			return
		}
		if "GET" == r.Method {
			wrapper := ResponseWrapper{
				Status: "ok",
				Data: DataWrapper{
					Threads: self.Threads,
				},
			}
			wrapper.Write(w)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	thread := self.Get(threadId)
	if nil == thread {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		self.getThreadHttpHandler(w, r, thread)
		break
	case "POST":
		self.postThreadHttpHandler(w, r, thread)
		err := self.Save()
		if nil != err {
			logger.Error(err)
		}
		break
	default:
		errors.New("Method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (self *BulletinBoard) createThreadFromHttpRequest(r *http.Request) (*Thread, error) {
	r.ParseForm()
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	thread := &Thread{}
	err := json.Unmarshal(result, thread)
	if nil != err {
		return thread, err
	}

	// Set these after Unmarshal to prevent overwriting
	thread.Time = time.Now().Unix()
	thread.Id = uuid.New().String()
	return thread, nil
}

func (self *BulletinBoard) getThreadHttpHandler(w http.ResponseWriter, r *http.Request, thread *Thread) {
	if nil == thread {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	wrapper := ResponseWrapper{
		Status: "ok",
		Data: DataWrapper{
			Thread: thread,
		},
	}

	err := wrapper.Write(w)
	if nil != err {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (self *BulletinBoard) postThreadHttpHandler(w http.ResponseWriter, r *http.Request, parent *Thread) {
	thread, err := self.createThreadFromHttpRequest(r)
	if nil != err {
		logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = parent.Add(thread); err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	self.getThreadHttpHandler(w, r, thread)
}

func New(filename string) (*BulletinBoard, error) {
	bb := &BulletinBoard{
		filename: filename,
	}

	_, err := os.Stat(filename)
	if !os.IsNotExist(err) {
		return bb, LoadGobFromFile(filename, bb)
	} else {
		return bb, nil
	}
	return bb, err
}

const (
	DEFAULT_HTTP_PORT int = 4444
)

var (
	HTTP_PORT int  = DEFAULT_HTTP_PORT
	DEBUG     bool = false
)

const (
	MAJOR_VERSION int    = 0
	MINOR_VERSION int    = 0
	PATCH_VERSION int    = 1
	PROJECT_NAME  string = "Bulletin Board Server"
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

	bb, err := New("bb.gob")
	if nil != err {
		panic(err)
	}

	server, _ := lemur.NewServer()
	server.AttachHandler("/bb/{threadId}", bb)
	server.AttachHandler("/bb", bb)
	server.ListenAndServe(HTTP_PORT)
}
