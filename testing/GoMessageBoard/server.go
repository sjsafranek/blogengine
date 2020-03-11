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
	"github.com/karlseguin/ccache"
	"github.com/sjsafranek/lemur"
	"github.com/sjsafranek/logger"
)

func Save(path string, object interface{}) error {
	file, err := os.Create(path)
	if nil != err {
		return err
	}
	defer file.Close()
	return gob.NewEncoder(file).Encode(object)
}

// Decode Gob file
func Load(path string, object interface{}) error {
	file, err := os.Open(path)
	if nil != err {
		return err
	}
	defer file.Close()
	return gob.NewDecoder(file).Decode(object)
}

type Thread struct {
	Id       string     `json:"id"`
	Author   string     `json:"author"`
	Content  string     `json:"content"`
	Time     int64      `json:"time"`
	Comments []*Thread  `json:"commments"`
	mux      sync.Mutex `json:"-"`
}

func (self *Thread) Add(thread *Thread) error {
	self.mux.Lock()
	self.Comments = append(self.Comments, thread)
	defer self.mux.Unlock()
	return nil
}

func (self *Thread) Get(threadId string) *Thread {
	for _, thread := range self.Comments {
		if threadId == thread.Id {
			return thread
		} else {
			thread = thread.Get(threadId)
			if threadId == thread.Id {
				return thread
			}
		}
	}
	return nil
}

func (self *Thread) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *Thread) Write(w io.Writer) error {
	payload, err := self.Marshal()
	if nil != err {
		return err
	}
	_, err = fmt.Fprintln(w, payload)
	return err
}

type BulletinBoard struct {
	filename string               `json:"file"`
	Root     map[string]*Thread   `json:"root"`
	mux      sync.Mutex           `json:"-"`
	cache    *ccache.LayeredCache `json:"-"`
}

func (self *BulletinBoard) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *BulletinBoard) Write(w io.Writer) error {
	payload, err := self.Marshal()
	if nil != err {
		return err
	}
	_, err = fmt.Fprintln(w, payload)
	return err
}

// func (self *BulletinBoard) get(threadId string) *Thread {
// 	item := self.cache.Get("thread", threadId)
// 	if nil != item {
// 		return item.Value().(*Thread)
// 	}
// 	return nil
// }
//
// func (self *BulletinBoard) set(thread  *Thread) {
// 	self.cache.Set("thread", thread.Id, thread, 5*time.Minute)
// }

func (self *BulletinBoard) Get(threadId string) *Thread {
	// thread := self.get(threadId)
	// if nil != thread {
	// 	return thread
	// }

	for id, list := range self.Root {
		if id == threadId {
			// self.set(list)
			return list
		}
		thread := list.Get(threadId)
		if nil != thread {
			// self.set(thread)
			return thread
		}
	}
	return nil
}

func (self *BulletinBoard) Add(thread *Thread) error {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.Root[thread.Id] = thread
	return nil
}

func (self *BulletinBoard) Has(threadId string) bool {
	self.mux.Lock()
	defer self.mux.Unlock()
	_, ok := self.Root[threadId]
	return ok
}

func (self *BulletinBoard) Save() error {
	return Save(self.filename, self)
}

func (self *BulletinBoard) Load() error {
	self.Root = make(map[string]*Thread)
	if _, err := os.Stat(self.filename); !os.IsNotExist(err) {
		return Load(self.filename, self)
	}
	return nil
}

func (self *BulletinBoard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadId := vars["threadId"]

	// create root thread
	if "" == threadId {
		if "POST" == r.Method {
			thread, err := self.createThread(r)
			if nil != err {
				panic(err)
			}
			self.Add(thread)

			err = self.Save()
			if nil != err {
				logger.Error(err)
			}

			self.handleGetList(w, r, thread)
			return
		}
		if "GET" == r.Method {
			self.Write(w)
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

	logger.Info(thread, r.Method)

	switch r.Method {
	case "GET":
		self.handleGetList(w, r, thread)
		break
	case "POST":
		self.handleAddThread(w, r, thread)
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

func (self *BulletinBoard) handleGetList(w http.ResponseWriter, r *http.Request, thread *Thread) {
	if nil != thread.Write(w) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (self *BulletinBoard) createThread(r *http.Request) (*Thread, error) {
	r.ParseForm()
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	thread := &Thread{}
	err := json.Unmarshal(result, thread)
	if nil != err {
		return thread, err
	}

	thread.Time = time.Now().Unix()
	thread.Id = uuid.New().String()
	return thread, nil
}

func (self *BulletinBoard) handleAddThread(w http.ResponseWriter, r *http.Request, parent *Thread) {
	thread, err := self.createThread(r)
	if nil != err {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = parent.Add(thread); err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	self.handleGetList(w, r, thread)
}

func New(filename string) *BulletinBoard {
	bb := &BulletinBoard{filename: filename}
	err := bb.Load()
	if nil != err {
		panic(err)
	}
	return bb
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

	var bb = New("bb.gob")

	server, _ := lemur.NewServer()
	server.AttachHandler("/bb/{threadId}", bb)
	server.AttachHandler("/bb", bb)
	// server.AttachHandler("/bb/{threadId}/comment", bb)
	server.ListenAndServe(HTTP_PORT)
}
