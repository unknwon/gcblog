// Copyright 2013 gcblog authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package models

import (
	"bytes"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Unknwon/com"
	"github.com/astaxie/beego"
	"github.com/gorilla/feeds"
	"github.com/howeyc/fsnotify"
	"github.com/slene/blackfriday"
)

type archive struct {
	Name, Title  string
	Author, Date string
	Content      []byte
	Excerpt      []byte
}

var (
	archives  []*archive
	works     []*archive
	eventTime = make(map[string]int64)
)

var Feed = &feeds.Feed{
	Title:       "Golang 中国博客",
	Link:        &feeds.Link{Href: "http://blog.go-china.org"},
	Description: "Golang 中国博客",
	Author:      &feeds.Author{"无闻", "joe2010xtmf@163.com"},
	Created:     time.Now(),
	Copyright:   "Copyright © 2013-2014 Golang 中国",
}

func loadArchiveNames() {
	archNames := strings.Split(beego.AppConfig.String("archives"), "|")
	archives = make([]*archive, 0, len(archNames))
	Feed.Items = make([]*feeds.Item, 0, len(archNames))
	for _, name := range archNames {
		arch := getFile(path.Join("content", name))
		if arch == nil {
			continue
		}
		arch.Name = name
		archives = append(archives, arch)
		Feed.Items = append(Feed.Items, &feeds.Item{
			Title:   arch.Title,
			Link:    &feeds.Link{Href: "http://blog.go-china.org/" + arch.Name},
			Author:  &feeds.Author{Name: arch.Author},
			Created: time.Now(),
		})
	}

	workNames := strings.Split(beego.AppConfig.String("works"), "|")
	works = make([]*archive, 0, len(workNames))
	for _, name := range workNames {
		if len(name) == 0 {
			continue
		}
		work := getFile(path.Join("work", name))
		if work == nil {
			continue
		}
		work.Name = name
		works = append(works, work)
	}
}

var blogs []blog

type blog struct {
	Name, Link, Intro string
}

func loadRecommendBlogs() {
	infos := strings.Split(beego.AppConfig.String("recommend_blogs"), "|||")
	blogs = make([]blog, 0, len(infos))
	for _, info := range infos {
		ins := strings.Split(info, "|")
		if len(ins) != 3 {
			continue
		}
		blogs = append(blogs, blog{ins[0], ins[1], ins[2]})
	}
}

// checkTMPFile returns true if the event was for TMP files.
func checkTMPFile(name string) bool {
	if strings.HasSuffix(strings.ToLower(name), ".tmp") {
		return true
	}
	return false
}

// getFileModTime retuens unix timestamp of `os.File.ModTime` by given path.
func getFileModTime(path string) int64 {
	path = strings.Replace(path, "\\", "/", -1)
	f, err := os.Open(path)
	if err != nil {
		beego.Error("Fail to open file:", err)
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		beego.Error("Fail to get file information:", err)
		return time.Now().Unix()
	}

	return fi.ModTime().Unix()
}

func init() {
	loadArchiveNames()
	loadRecommendBlogs()

	// Watch changes.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		beego.Error("NewWatcher:", err)
		return
	}

	go func() {
		for {
			select {
			case e := <-watcher.Event:

				// Skip TMP files for Sublime Text.
				if checkTMPFile(e.Name) {
					continue
				}

				mt := getFileModTime(e.Name)
				if t := eventTime[e.Name]; mt == t {
					continue
				}

				eventTime[e.Name] = mt
				beego.Info("Changes detected")
				beego.ParseConfig()
				loadArchiveNames()
				loadRecommendBlogs()
			case err := <-watcher.Error:
				beego.Error("Watcher error:", err)
			}
		}
	}()

	err = watcher.Watch("content")
	if err != nil {
		beego.Error("Watch path:", err)
		return
	}
	err = watcher.Watch("work")
	if err != nil {
		beego.Error("Watch path:", err)
		return
	}
	err = watcher.Watch("conf")
	if err != nil {
		beego.Error("Watch path:", err)
		return
	}
}

func markdown(raw []byte) []byte {
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_XHTML
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	htmlFlags |= blackfriday.HTML_GITHUB_BLOCKCODE
	htmlFlags |= blackfriday.HTML_OMIT_CONTENTS
	htmlFlags |= blackfriday.HTML_COMPLETE_PAGE
	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	// set up the parser
	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_HARD_LINE_BREAK
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK

	body := blackfriday.Markdown(raw, renderer, extensions)
	return body
}

func getFile(filePath string) *archive {
	df := &archive{}
	p, err := com.ReadFile(filePath + ".md")
	if err != nil {
		beego.Error("models.getFile -> ", err)
		return nil
	}

	s := string(p)
	// Parse author and date.
	if i := strings.Index(s, "---\n"); i > -1 {
		if j := strings.Index(s[i+4:], "---\n"); j > -1 {
			infos := strings.Split(s[i+4:j+i+4], "\n")
			for k := range infos {
				z := strings.Index(infos[k], ":")
				if z == -1 {
					continue
				}

				switch k {
				case 0: // Author.
					df.Author = strings.TrimSpace(infos[k][z+1:])
				case 1: // Date.
					df.Date = strings.TrimSpace(infos[k][z+1:])
				}
			}
			s = s[j+i+8:]
			for {
				if !strings.HasPrefix(s, "\n") {
					break
				}
				s = s[1:]
			}
		}
	}

	// Parse and render.
	i := strings.Index(s, "\n")
	if i > -1 {
		// Has title.
		df.Title = strings.TrimSpace(
			strings.Replace(s[:i+1], "#", "", -1))
		df.Content = []byte(strings.TrimSpace(s[i+2:]))
	} else {
		df.Content = p
	}

	df.Content = markdown(df.Content)

	// add "click to read more" feature
	more := []byte("\n<p>[more]</p>\n")
	if i := bytes.Index(df.Content, more); i != -1 {
		df.Excerpt = df.Content[:i]
		for k := 0; k < len(more); k++ {
			df.Content[i+k] = ' '
		}
	}

	return df
}

func GetAllPosts() []*archive {
	return archives
}

func GetRecentPosts(num int) []*archive {
	if len(archives) >= num {
		return archives[:num]
	}
	return archives
}

func GetSinglePost(name string) *archive {
	for _, arch := range archives {
		if name == arch.Name {
			return arch
		}
	}
	return nil
}

func GetAllWorks() []*archive {
	return works
}

func GetSingleWork(name string) *archive {
	for _, work := range works {
		if name == work.Name {
			return work
		}
	}
	return nil
}

func GetBlogs() []blog {
	return blogs
}
