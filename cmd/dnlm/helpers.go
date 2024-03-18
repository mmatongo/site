package main

import (
	"bufio"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	birthTime, ok := a.getFileCreationTime(nativeInfo)
	if !ok {
		return time.Time{}, errors.New("file system doesn't support creation time")
	}

	return birthTime, nil
}

func (a *app) getFileCreationTime(nativeInfo *syscall.Stat_t) (time.Time, bool) {
	birthTime := a.getBirthTime(nativeInfo)
	if !birthTime.IsZero() {
		return birthTime, true
	}

	modTime := a.getModTime(nativeInfo)
	if !modTime.IsZero() {
		return modTime, false
	}

	return time.Time{}, false
}

func getTimeFromTimespec(_ *syscall.Stat_t, specField unsafe.Pointer) time.Time {
	tspec := (*syscall.Timespec)(specField)
	if tspec == nil || (tspec.Sec == 0 && tspec.Nsec == 0) {
		return time.Time{}
	}
	return time.Unix(tspec.Sec, int64(tspec.Nsec))
}

func (a *app) getNameFromFilePath(filePath string) string {
	title := cases.Title(language.Und).String(filepath.Base(filePath))
	title = strings.Replace(title, "-", " ", -1)
	title = strings.TrimSuffix(title, ".md")

	return title
}

func (a *app) estimateReadingTime(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	wordCount := 0
	wordRegex := regexp.MustCompile(`\w+`)

	for scanner.Scan() {
		line := scanner.Text()
		cleanedLine := a.removeMarkdownSyntax(line)
		words := wordRegex.FindAllString(cleanedLine, -1)
		wordCount += len(words)
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}

	const averageWordsPerMinute = 200
	totalReadingTimeSeconds := (wordCount * 60) / averageWordsPerMinute

	readingTimeMinutes := totalReadingTimeSeconds / 60
	readingTimeSeconds := totalReadingTimeSeconds % 60

	if readingTimeMinutes == 0 && wordCount > 0 {
		readingTimeMinutes = 1
	}

	return readingTimeMinutes, readingTimeSeconds, nil
}

func (a *app) removeMarkdownSyntax(text string) string {
	patterns := []string{
		`\[\S.*?\]\(.*?\)`,  // Links: [link text](url)
		`!\[\S.*?\]\(.*?\)`, // Images: ![alt text](image url)
		`__\S.*?__`,         // Bold with __text__
		`\\*\\*.*?\\*\\*`,   // Bold with **text**
		`_\S.*?_`,           // Italic with _text_
		`\\*.*?\\*`,         // Italic with *text*
		"`.*?`",             // Inline code with `code`
		`~~.*?~~`,           // Strikethrough with ~~text~~
		`<.*?>`,             // HTML tags
		`#+`,                // Headings
		`-{3,}`,             // Horizontal rules
		`-{2,}`,             // Em dashes
		`-{1,}`,             // En dashes
		`[0-9]+\..*`,        // Ordered lists
		`[*+-].*`,           // Unordered lists
		`>.*`,               // Blockquotes
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}
	return text
}
