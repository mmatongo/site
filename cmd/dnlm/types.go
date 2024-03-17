package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

type app struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	config   map[string]interface{}
	watcher  *fsnotify.Watcher
}

type blogPost struct {
	Title string
	URL   string
	Time  string
}
