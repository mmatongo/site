package main

import (
	"fmt"
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

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		a.errorLog.Printf("Error getting file info: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	creationDate, err := a.getCreationDate(fileInfo)

	if err != nil {
		a.errorLog.Printf("Error getting creation date: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	date := creationDate.Format("January 2, 2006 3:04 PM")

	formattedContent := fmt.Sprintf("### Published: %s\n\n%s", date, data)

	output := markdown.ToHTML([]byte(formattedContent), nil, nil)

	name := a.getNameFromFilePath(filePath)

	err = blogTmpl.ExecuteTemplate(w, "blog", map[string]interface{}{
		"Name":            fmt.Sprintf("%s - %s", a.config["blog_name"], name),
		"DescriptionMeta": a.config["description"],
		"Content":         template.HTML(output),
		"Path":            r.URL.Path,
	})

	if err != nil {
		a.errorLog.Printf("Template execution error: %v\n", err)
	}
}

func (a *app) handleBlogIndex(w http.ResponseWriter, r *http.Request) {
	posts := []blogPost{}

	mapMutex.RLock()
	for url, filePath := range urlToFileMap {

		// though golang.org/x/text/case can be used, it's not worth the dependency
		title := a.getNameFromFilePath(filePath)
		minutes, seconds, err := a.estimateReadingTime(filePath)
		if err != nil {
			a.errorLog.Printf("Error estimating reading time: %v\n", err)
		}
		timeToRead := fmt.Sprintf("%d.%d min", minutes, seconds)

		posts = append(posts, blogPost{
			Title: title,
			URL:   url,
			Time:  timeToRead,
		})
	}
	mapMutex.RUnlock()

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Title < posts[j].Title
	})

	err := indexTmpl.ExecuteTemplate(w, "blogIndex", map[string]interface{}{
		"Posts":      posts,
		"BlogName":   a.config["blog_name"],
		"Profession": a.config["profession"],
		"Name":       a.config["blog_name"],
	})

	if err != nil {
		a.errorLog.Printf("Template execution error: %v\n", err)
	}
}
