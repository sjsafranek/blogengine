package blogengine

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
	"net/http"

	"github.com/BurntSushi/toml"
	"gopkg.in/russross/blackfriday.v2"
	"github.com/sjsafranek/logger"
)

// Post is the basic post
type Post struct {
	// set in frontmatter
	// meta data
	Title       string
	Date        time.Time
	Description string
	Image       string
	// stuff for the title
	Number       string // 001
	QuoteAuthor  string // Theloninus Monk
	QuoteContent string // What! This is a piano?

	// generated from file
	Slug            string // generated from filename
	ContentMarkdown string
	ContentHTML     template.HTML
}

// Parse will parse a blog file
func Parse(fname string) (post Post, err error) {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return
	}

	frontMatter := bytes.Split(b, []byte("---"))
	if len(frontMatter) != 2 {
		err = fmt.Errorf("incorrect frontmatter for %s", fname)
		return
	}

	err = toml.Unmarshal(frontMatter[0], &post)
	if err != nil {
		return
	}
	output := blackfriday.Run(bytes.TrimSpace(frontMatter[1]))
	post.ContentMarkdown = string(frontMatter[1])
	post.ContentHTML = template.HTML(string(output))
	_, fname = filepath.Split(fname)
	post.Slug = strings.TrimSuffix(fname, ".md")
	return
}

type BlogEngine struct {
	Directory string
	BasePath string
	Template *template.Template
	TemplateName string
}

func (self *BlogEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, self.BasePath)
	page := parts[1]
	if "/" == page {
		page = "/index"
	}

	pagePath := filepath.Join(self.Directory, page)
	pagePath = fmt.Sprintf("%v.md", pagePath)

	post, err := Parse(pagePath)
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = self.Template.ExecuteTemplate(w, self.TemplateName, post)
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
