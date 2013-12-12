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
	"path"
	"strings"

	"github.com/Unknwon/com"
	"github.com/astaxie/beego"
	"github.com/slene/blackfriday"
)

type archive struct {
	Name    string
	Title   string
	Content []byte
}

var archives []*archive

func init() {
	archNames := strings.Split(beego.AppConfig.String("archives"), "|")
	archives = make([]*archive, 0, len(archNames))
	for _, name := range archNames {
		arch := getFile(path.Join("content", name))
		if arch == nil {
			continue
		}
		arch.Name = name
		archives = append(archives, arch)
	}
}

func markdown(raw []byte) (out []byte) {
	return blackfriday.MarkdownCommon(raw)
}

func getFile(filePath string) *archive {
	df := &archive{}
	p, err := com.ReadFile(filePath + ".md")
	if err != nil {
		beego.Error("models.getFile -> ", err)
		return nil
	}

	// Parse and render.
	s := string(p)
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
	return df
}

func GetRecentPosts() []*archive {
	if len(archives) >= 10 {
		return archives[:10]
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
