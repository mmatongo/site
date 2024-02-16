package main

import (
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gomarkdown/markdown"
)

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	err := indexTmpl.ExecuteTemplate(w, "index", map[string]interface{}{
		"Name":            a.config["name"],
		"Profession":      a.config["profession"],
		"DescriptionBody": a.config["description"],
		"DescriptionMeta": a.config["description"],
	})

	if err != nil {
		a.errorLog.Printf("Template execution error: %v\n", err)
	}
}

func (a *app) handleBlog(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/blog/")
	mapMutex.RLock()
	filePath, ok := urlToFileMap[path]
	mapMutex.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	data, err := os.ReadFile(filePath)

	if err != nil {
		a.errorLog.Printf("Error reading file: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	output := markdown.ToHTML(data, nil, nil)
	err = blogTmpl.ExecuteTemplate(w, "blog", map[string]interface{}{
		"Name":            a.config["blog_name"],
		"DescriptionMeta": a.config["description"],
		"Content":         template.HTML(output),
	})

	if err != nil {
		a.errorLog.Printf("Template execution error: %v\n", err)
	}
}

func (a *app) handleBlogIndex(w http.ResponseWriter, r *http.Request) {
	posts := []blogPost{}

	mapMutex.RLock()
	for url, filePath := range urlToFileMap {
		title := strings.Replace(strings.TrimPrefix(filePath, "blog/"), "-", " ", -1)
		// though golang.org/x/text/case can be used, it's not worth the dependency
		title = strings.Title(strings.TrimSuffix(title, ".md"))
		posts = append(posts, blogPost{
			Title: title,
			URL:   url,
		})
	}
	mapMutex.RUnlock()

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Title < posts[j].Title
	})

	err := indexTmpl.ExecuteTemplate(w, "blogIndex", map[string]interface{}{
		"Posts":    posts,
		"BlogName": a.config["blog_name"],
		"Name":     a.config["blog_name"],
	})

	if err != nil {
		a.errorLog.Printf("Template execution error: %v\n", err)
	}
}
