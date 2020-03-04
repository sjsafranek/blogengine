package blogengine

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"path"
	"sort"
	// "os"

	"github.com/BurntSushi/toml"
	"github.com/sjsafranek/logger"
	"gopkg.in/russross/blackfriday.v2"
	// blackfriday "gopkg.in/russross/blackfriday.v1"
)

// PostData is generated from the body the markdown file
type PostData struct {
	Slug            string        `toml:"-"`
	ContentMarkdown string        `toml:"-"`
	ContentHTML     template.HTML `toml:"-"`
}

// PostMetaData contains the metadata found in the
// header of the markdown file.
type PostMetaData struct {
	Title        string    `toml:"Title"`
	Date         time.Time `toml:"Date"`
	Description  string    `toml:"Description"`
	Image        string    `toml:"Image"`
	Number       int64     `toml:"Number"`
	Author  string    `toml:"Author"`
	Tags         []string  `toml:"Tags"`
}

type Post struct {
	PostData
	PostMetaData
	Posts []*Post
}

// func (self *Post) Write(w io.Writer) error {
// 	payload, err := self.Marshal()
// 	if nil != err {
// 		return err
// 	}
// 	_, err = fmt.Fprintln(w, payload)
// 	return err
// }

// Parse will parse a blog file
func Parse(fname string) (*Post, error) {
	var post Post

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return &post, err
	}

	frontMatter := bytes.Split(b, []byte("---"))
	if len(frontMatter) != 2 {
		err = fmt.Errorf("incorrect frontmatter for %s", fname)
		return &post, err
	}

	err = toml.Unmarshal(frontMatter[0], &post)
	if err != nil {
		return &post, err
	}
	output := blackfriday.Run(bytes.TrimSpace(frontMatter[1]))
	post.ContentMarkdown = string(frontMatter[1])
	post.ContentHTML = template.HTML(string(output))
	_, fname = filepath.Split(fname)
	post.Slug = strings.TrimSuffix(fname, ".md")

	return &post, err
}

// GetAll returns all posts within a directory
func GetAll(directory string) ([]*Post, error) {
	var err error
	fnames, _ := ioutil.ReadDir(directory)
	posts := make([]*Post, len(fnames))
	for i, fname := range fnames {
		posts[i], err = Parse(path.Join(directory, fname.Name()))
		if err != nil {
			posts = posts[:i]
			return posts, err
		}
	}
	// sort by date
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.Before(posts[j].Date)
	})
	return posts, err
}

type BlogEngine struct {
	Directory    string
	BasePath     string
	Template     *template.Template
	TemplateName string
}

// ServeHTTP http request handler
func (self *BlogEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var post *Post
	var err error

	// check for root url
	if !strings.Contains(r.URL.String(), fmt.Sprintf("%v/", self.BasePath)) {

		posts, err := GetAll(self.Directory)
		post = &Post{Posts: posts}
		if nil != err {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {

		parts := strings.Split(r.URL.Path, self.BasePath)
		page := parts[1]
		if "/" == page {
			page = "/index"
		}

		pagePath := filepath.Join(self.Directory, page)

		if "/" == page[len(page)-1:] {
			// build post from directory
			posts, err := GetAll(pagePath)
			post = &Post{Posts: posts}
			if nil != err {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// build post from file
			filePath := fmt.Sprintf("%v.md", pagePath)
			post, err = Parse(filePath)
			if nil != err {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	}

	// render in template
	err = self.Template.ExecuteTemplate(w, self.TemplateName, post)
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
