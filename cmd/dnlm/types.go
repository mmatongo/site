package main

import (
	"encoding/xml"
	"log"
	"time"

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

type Urlset struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	Urls    []Url    `xml:"url"`
}

type Url struct {
	Loc        string    `xml:"loc"`
	LastMod    time.Time `xml:"lastmod"`
	ChangeFreq string    `xml:"changefreq"`
	Priority   float64   `xml:"priority"`
}

type RssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	XMLAtom string     `xml:"xmlns:atom,attr"`
	Version string     `xml:"version,attr"`
	Channel RssChannel `xml:"channel"`
}

type RssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	AtomLink    string    `xml:"atom:link"`
	Lang        string    `xml:"language"`
	PubDate     string    `xml:"pubDate"`
	CopyRight   string    `xml:"copyright"`
	Items       []RssItem `xml:"item"`
}

type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}
