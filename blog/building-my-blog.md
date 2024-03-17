# Building my blog

A few weeks ago I decided to rebuild my website. initially, I had been using a highly customised next.js blog starter but it never really felt like it was mine. In my initial design, I was aiming for something along the lines of [this](https://muan.co) but I never really got to that point. I was always too busy to work on it and when I did, I was always too tired to make any meaningful progress. So I decided to start from scratch and build something that I could be proud of.

> “How much you can learn when you fail determines how far you will go into achieving your goals.”
>
> ― Roy Bennett

I started by looking at a few designs that fit the aesthetic I was going for and I found [this](https://suzenfylke.com/). I liked the minimalistic design and the overall flow of it. It was also bare enough that I could add my personal touch to it.

> “You can't plan for everything or you never get started in the first place.”
>
> ― Jim Butcher, [Changes](https://www.goodreads.com/work/quotes/6778696-changes)

When I was starting I had a few goals in mind. I wanted to build something that was fast, minimalistic and easy to use. I also wanted to build something that I could easily extend and add new features to as well as something that I could easily maintain. It being statically generated was also a big plus for me. But most importantly, I needed it to be dynamic and not just a static site that I could add new content to it without having to rebuild the entire site and without using a DB.

This blog is organised like this:

```sh
blog/
├── cmd/
│   └── dnlm/
├── config/
└── ui/
    ├── static/
    └── templates/
```


### The all seeing watcher

```go
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
```
Here's how I achieved this. I use the fsnotify package to watch for changes and rebuild the paths to the blog posts. The posts are built dynamically from makrdown in the blog directory so when a new post is added or an existing one is modified, we do the equivalent of a `go run main.go` to rebuild the paths to the posts. This is done by calling the `updateURLToFileMap` function:

```go
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
```

This is by far not the most efficient wat to do this and the code makes alot of very unhealthy assumptions but it works for now.
Basically, we get rid of the blog/ prefix and the .md suffix and replace all underscores with hyphens. We then add the path to the map if it exists and remove it if it doesn't.

The repository for the blog is open source and can be found [here](https://github.com/mmatongo/site). It can be found on the v2 branch. I'm still working on it and I'm open to any suggestions or contributions.
