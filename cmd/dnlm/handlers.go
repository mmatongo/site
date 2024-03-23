package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
)

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
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
		a.pageNotFound(w)
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

	name := a.getNameFromFilePath(filePath)

	formattedContent := fmt.Sprintf("# %s\n\n## %s\n\n%s", name, date, data)

	output := markdown.ToHTML([]byte(formattedContent), nil, nil)

	err = blogTmpl.ExecuteTemplate(w, "blog", map[string]interface{}{
		"Name":            fmt.Sprintf("%s - %s", a.config["blog_name"], name),
		"DescriptionMeta": a.config["description"],
		"Content":         template.HTML(output),
		"Path":            r.URL.Path,
		"Source":          a.config["repository"].(string) + filePath + "?plain=1",
	})

	if err != nil {
		a.errorLog.Printf("Template execution error: %v\n", err)
	}
}

func (a *app) handleBlogIndex(w http.ResponseWriter, r *http.Request) {
	posts := []blogPost{}

	mapMutex.RLock()
	for url, filePath := range urlToFileMap {
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

func (a *app) pageNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	err := notFoundTmpl.ExecuteTemplate(w, "404", map[string]interface{}{
		"DescriptionMeta": a.config["description"],
	})

	if err != nil {
		a.errorLog.Printf("Template execution error: %v\n", err)
	}
}

func (a *app) handleSitemap(w http.ResponseWriter, r *http.Request) {
	var urls []Url

	urls = append(urls, Url{
		Loc:        "https://dnlm.pw/",
		LastMod:    time.Now(),
		ChangeFreq: "daily",
		Priority:   1.0,
	})

	urls = append(urls, Url{
		Loc:        "https://dnlm.pw/blog",
		LastMod:    time.Now(),
		ChangeFreq: "weekly",
		Priority:   0.9,
	})

	mapMutex.RLock()
	for urlPath := range urlToFileMap {
		fileInfo, err := os.Stat(urlToFileMap[urlPath])
		if err != nil {
			a.errorLog.Printf("Error getting file info: %v\n", err)
			continue
		}

		urls = append(urls, Url{
			Loc:        "https://dnlm.pw/blog/" + urlPath,
			LastMod:    fileInfo.ModTime(),
			ChangeFreq: "monthly",
			Priority:   0.8,
		})

		sort.Slice(urls, func(i, j int) bool {
			return urls[i].LastMod.After(urls[j].LastMod)
		})
	}
	mapMutex.RUnlock()

	urlset := Urlset{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		Urls:  urls,
	}

	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(urlset)
}

func (a *app) handleRSS(w http.ResponseWriter, r *http.Request) {
	var items []RssItem

	mapMutex.RLock()
	for urlPath, filePath := range urlToFileMap {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		title := a.getNameFromFilePath(filePath)
		description := string(markdown.ToHTML(data, nil, nil))
		link := "https://dnlm.pw/blog/" + urlPath
		pubDate := fileInfo.ModTime().Format(time.RFC1123Z)

		items = append(items, RssItem{
			Title:       title,
			Link:        link,
			Description: description,
			PubDate:     pubDate,
			GUID:        link,
		})

		sort.Slice(items, func(i, j int) bool {
			return items[i].PubDate > items[j].PubDate
		})
	}
	mapMutex.RUnlock()

	feed := RssFeed{
		Version: "2.0",
		XMLAtom: "http://www.w3.org/2005/Atom",
		Channel: RssChannel{
			Title:       a.config["blog_name"].(string),
			Link:        "https://dnlm.pw/blog",
			Description: a.config["description"].(string),
			AtomLink:    "https://dnlm.pw/rss.xml",
			Lang:        "en-gb",
			PubDate:     time.Now().Format(time.RFC1123Z),
			CopyRight:   "MIT License",
			Items:       items,
		},
	}

	w.Header().Set("Content-Type", "application/xml; charset=UTF-8")
	err := xml.NewEncoder(w).Encode(feed)
	if err != nil {
		a.errorLog.Printf("Error encoding RSS feed: %v\n", err)
	}
}
