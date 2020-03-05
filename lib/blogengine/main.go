package blogengine

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
	Title       string    `toml:"Title"`
	Date        time.Time `toml:"Date"`
	Description string    `toml:"Description"`
	Image       string    `toml:"Image"`
	Author      string    `toml:"Author"`
	Tags        []string  `toml:"Tags"`
}

type Post struct {
	PostData
	PostMetaData
	Posts []*Post
	// IsDirectory bool
}

// func (self *Post) Write(w io.Writer) error {
// 	payload, err := self.Marshal()
// 	if nil != err {
// 		return err
// 	}
// 	_, err = fmt.Fprintln(w, payload)
// 	return err
// }

type BlogEngine struct {
	Directory    string
	BasePath     string
	Template     *template.Template
	TemplateName string
}

func (self *BlogEngine) getSlug(fpath string) string {
	return strings.TrimSuffix(strings.Replace(fpath, self.Directory, self.BasePath, -1), ".md")
}

// Parse will parse a blog file
func (self *BlogEngine) Parse(fname string) (*Post, error) {
	var post Post

	post.Title = strings.TrimSuffix(filepath.Base(fname), ".md")

	// check if file or directory
	fileInfo, err := os.Stat(fname)
	if err != nil {
		return &post, err
	}

	switch mode := fileInfo.Mode(); {

	// handle directory
	case mode.IsDir():
		post.Slug = self.getSlug(fname) + "/"
		return &post, err

		// handle file post
	case mode.IsRegular():
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
		post.Slug = self.getSlug(fname)
		return &post, err
	}

	return &post, nil
}

// GetAll returns all posts within a directory
func (self *BlogEngine) GetAll(directory string) ([]*Post, error) {
	var err error
	fnames, _ := ioutil.ReadDir(directory)
	posts := make([]*Post, len(fnames))
	for i, fname := range fnames {
		posts[i], err = self.Parse(path.Join(directory, fname.Name()))
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

// ServeHTTP http request handler
func (self *BlogEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var post *Post
	var err error

	// check for root url
	if !strings.Contains(r.URL.String(), fmt.Sprintf("%v/", self.BasePath)) {

		posts, err := self.GetAll(self.Directory)
		post = &Post{Posts: posts}
		if nil != err {
			logger.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {

		// get post/page path
		parts := strings.Split(r.URL.Path, self.BasePath)
		page := parts[1]
		if "/" == page {
			page = "/index"
		}
		pagePath := filepath.Join(self.Directory, page)

		if "/" == page[len(page)-1:] {
			// build post from directory
			posts, err := self.GetAll(pagePath)
			post = &Post{Posts: posts}
			if nil != err {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// build post from file
			filePath := fmt.Sprintf("%v.md", pagePath)
			post, err = self.Parse(filePath)
			if nil != err {
				logger.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// get posts in same directory
			post.Posts, err = self.GetAll(filepath.Dir(filePath))
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
