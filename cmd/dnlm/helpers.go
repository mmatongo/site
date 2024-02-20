package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

func (a *app) initConfig() {
	config, err := os.ReadFile("config/config.yml")
	if err != nil {
		a.errorLog.Fatalf("Error reading config file: %v", err)
	}
	err = yaml.Unmarshal(config, &a.config)

	if err != nil {
		a.errorLog.Fatalf("Error unmarshalling config file: %v", err)
	}
}

func (a *app) initURLToFileMap() {
	err := filepath.WalkDir("blog", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			updateURLToFileMap(path)
		}
		return nil
	})

	if err != nil {
		a.errorLog.Fatalf("Failed to walk blog directory: %v", err)
	}
}

func (a *app) watchFiles() {
	var err error
	a.watcher, err = fsnotify.NewWatcher()

	if err != nil {
		a.errorLog.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-a.watcher.Events:
				if !ok {
					return
				}
				switch event.Op {
				case fsnotify.Write, fsnotify.Create, fsnotify.Remove, fsnotify.Rename:
					a.infoLog.Printf("%s: %s\n", event.Op, event.Name)
					updateURLToFileMap(event.Name)
				case fsnotify.Chmod:
					// Ignore CHMOD events
				}
			case err, ok := <-a.watcher.Errors:
				if !ok {
					return
				}
				a.errorLog.Println("error:", err)
			}
		}
	}()

	err = a.watcher.Add("blog")
	if err != nil {
		a.errorLog.Println(err.Error())
	}
}

func updateURLToFileMap(filePath string) {
	if filepath.Ext(filePath) != ".md" {
		return
	}

	urlPath := strings.TrimPrefix(filePath, "blog/")
	urlPath = strings.TrimSuffix(urlPath, ".md")
	urlPath = strings.ReplaceAll(urlPath, "_", "-")

	mapMutex.Lock()
	defer mapMutex.Unlock()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		delete(urlToFileMap, urlPath)
	} else {
		urlToFileMap[urlPath] = filePath
	}
}
