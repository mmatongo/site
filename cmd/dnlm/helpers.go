package main

import (
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

func (a *app) initConfig(configPath string) {
	config, err := os.ReadFile(configPath)
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
			a.updateURLToFileMap(path)
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
					a.updateURLToFileMap(event.Name)
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

func (a *app) updateURLToFileMap(filePath string) {
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

func (a *app) acceptArgs() string {
	var config string
	flag.StringVar(&config, "config", "", "Path to your custom config.")
	flag.Parse()

	if config == "" {
		a.infoLog.Println("No config path provided, defaulting to config/config.yml")
		config = "config/config.yml"
	} else {
		a.infoLog.Printf("Config path %s received, parsing yaml file for values.", config)
	}

	return config
}

func (a *app) getCreationDate(info os.FileInfo) (time.Time, error) {
	if info == nil {
		return time.Time{}, errors.New("file information is nil")
	}

	nativeInfo, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}, errors.New("failed to get native file information")
	}

	birthTime, ok := getFileCreationTime(nativeInfo)
	if !ok {
		return time.Time{}, errors.New("file system doesn't support creation time")
	}

	return birthTime, nil
}

func getFileCreationTime(nativeInfo *syscall.Stat_t) (time.Time, bool) {
	birthTime := getBirthTime(nativeInfo)
	if !birthTime.IsZero() {
		return birthTime, true
	}

	modTime := getModTime(nativeInfo)
	if !modTime.IsZero() {
		return modTime, false
	}

	return time.Time{}, false
}

func getBirthTime(nativeInfo *syscall.Stat_t) time.Time {
	return getTimeFromTimespec(nativeInfo, unsafe.Pointer(&nativeInfo.Ctimespec))
}

func getModTime(nativeInfo *syscall.Stat_t) time.Time {
	return getTimeFromTimespec(nativeInfo, unsafe.Pointer(&nativeInfo.Mtimespec))
}

func getTimeFromTimespec(_ *syscall.Stat_t, specField unsafe.Pointer) time.Time {
	tspec := (*syscall.Timespec)(specField)
	if tspec == nil || (tspec.Sec == 0 && tspec.Nsec == 0) {
		return time.Time{}
	}
	return time.Unix(tspec.Sec, int64(tspec.Nsec))
}
