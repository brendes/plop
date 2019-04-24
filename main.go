// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	blackfriday "github.com/russross/blackfriday/v2"
)

type pageMetadata struct {
	URI         string
	Title       string
	Description string
	Date        time.Time
	Body        string
	template    string
	filename    string
	rss         bool
}

var tpl *template.Template

func parseFile(filename string) (*pageMetadata, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	content := string(b)
	separator := "\n---\n"
	position := strings.Index(content, separator)
	if position <= 0 {
		fmt.Printf("No header section detected for %q\n", filename)
		return nil, nil
	}

	header := content[:position]
	page := &pageMetadata{
		Body:     content[position+len(separator):],
		filename: filename,
	}

	for _, line := range strings.Split(header, "\n") {
		switch {
		case strings.HasPrefix(line, "title: "):
			page.Title = line[7:]
		case strings.HasPrefix(line, "description: "):
			page.Description = line[13:]
		case strings.HasPrefix(line, "template: "):
			page.template = line[10:]
		case strings.HasPrefix(line, "uri: "):
			page.URI = line[5:]
		case strings.HasPrefix(line, "date: "):
			page.Date, _ = time.Parse("2006-01-02", line[6:])
		case strings.HasPrefix(line, "rss: true"):
			page.rss = true
		}
	}

	if page.Title == "" || page.template == "" || page.URI == "" {
		fmt.Printf("Missing required header parameter (title, template, uri) for %q\n", filename)
		return nil, nil
	}

	if tpl.Lookup(page.template) == nil {
		fmt.Printf("The template %q specified in %q is not defined\n", page.template, filename)
		return nil, nil
	}

	return page, nil
}

func browse(folder string) (files []string, err error) {
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func generatePage(page *pageMetadata) error {
	directory := filepath.Dir(page.URI)
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Printf("Creating directory %q\n", directory)
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Creating file %q\n", page.URI)
	file, err := os.Create(page.URI)
	if err != nil {
		return err
	}
	defer file.Close()

	return tpl.ExecuteTemplate(file, page.template, page)
}

func generateFeed(entries []pageMetadata) error {
	sort.Slice(entries, func(i, j int) bool {
		return entries[j].Date.Before(entries[i].Date)
	})

	filename := "feed.xml"
	fmt.Printf("Creating file %q\n", filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tpl.ExecuteTemplate(file, "rss", map[string]interface{}{"Entries": entries})
}

func build(srcFolder, tplFolder string) (err error) {
	tplFuncs := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
		"isodate": func(ts time.Time) string {
			return ts.Format("2006-01-02")
		},
		"strdate": func(ts time.Time) string {
			return ts.Format("January 2, 2006")
		},
		"atomdate": func(ts time.Time) string {
			return ts.Format(time.RFC3339)
		},
		"now": func() time.Time {
			return time.Now()
		},
		"xmlprolog": func() template.HTML {
			return template.HTML(`<?xml version="1.0" encoding="utf-8"?>`)
		},
		"markdown": func(input string) template.HTML {
			return template.HTML(blackfriday.Run([]byte(input)))
		},
		"cdata": func(s template.HTML) template.HTML {
			return template.HTML(`<![CDATA[` + s + `]]>`)
		},
	}

	tpl, err = template.New("main").Funcs(tplFuncs).ParseGlob(tplFolder + "/*.html")
	if err != nil {
		return err
	}

	files, err := browse(srcFolder)
	if err != nil {
		return err
	}

	var feedEntries []pageMetadata
	for _, file := range files {
		page, err := parseFile(file)
		if err != nil {
			return err
		}

		if page != nil {
			if err := generatePage(page); err != nil {
				return err
			}

			if page.rss {
				feedEntries = append(feedEntries, *page)
			}
		}
	}

	if len(feedEntries) > 0 {
		if err := generateFeed(feedEntries); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	port := flag.String("port", "3000", "Port to listen on")
	serve := flag.Bool("serve", false, "Start local HTTP server")
	srcFolder := flag.String("src", "src", "Sources folder")
	tplFolder := flag.String("tpl", "templates", "Templates folder")
	flag.Parse()

	if *serve {
		fmt.Printf("Listening on HTTP port: %s\n", *port)
		err := http.ListenAndServe("127.0.0.1:"+*port, http.FileServer(http.Dir(".")))
		if err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if err := build(*srcFolder, *tplFolder); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
