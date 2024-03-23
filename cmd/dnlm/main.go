package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	indexFiles = []string{
		"./ui/templates/index.layout.gohtml",
		"./ui/templates/blog_index.page.gohtml",
		"./ui/templates/head.partial.gohtml",
		"./ui/templates/footer.partial.gohtml",
	}

	indexTmpl = template.Must(template.ParseFiles(indexFiles...))

	blogFiles = []string{
		"./ui/templates/blog.layout.gohtml",
		"./ui/templates/head.partial.gohtml",
		"./ui/templates/footer.partial.gohtml",
	}

	blogTmpl = template.Must(template.ParseFiles(blogFiles...))

	notFound = []string{
		"./ui/templates/404.layout.gohtml",
		"./ui/templates/head.partial.gohtml",
		"./ui/templates/footer.partial.gohtml",
	}

	notFoundTmpl = template.Must(template.ParseFiles(notFound...))

	mapMutex     = &sync.RWMutex{}
	urlToFileMap = make(map[string]string)
)

func main() {
	a := &app{
		errorLog: log.New(os.Stderr, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile),
		infoLog:  log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime),
		config:   make(map[string]interface{}),
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	config := a.acceptArgs()

	a.initConfig(config)
	a.initURLToFileMap()
	a.watchFiles()

	if a.config != nil {
		a.infoLog.Printf("Config file has been successfully loaded")
	} else {
		a.errorLog.Panic("Config file is empty")
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./ui/static"))))
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/blog", a.handleBlogIndex)
	mux.HandleFunc("/blog/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/blog/" {
			a.handleBlog(w, r) // Handle individual blog posts
		} else {
			a.handleBlogIndex(w, r) // Handle the blog index (this is weird)
		}
	})
	mux.HandleFunc("/sitemap.xml", a.handleSitemap)
	mux.HandleFunc("/rss.xml", a.handleRSS)

	srv := &http.Server{
		Addr:     ":" + port,
		Handler:  mux,
		ErrorLog: a.errorLog,
	}

	a.infoLog.Printf("Server starting at port %s", port)
	err := srv.ListenAndServe()
	if err != nil {
		a.errorLog.Fatal("ListenAndServe: ", err)
	}
}
