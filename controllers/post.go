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

package controllers

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/astaxie/beego"

	"github.com/Unknwon/gcblog/models"
)

type PostController struct {
	baseController
}

func (this *PostController) Get() {
	reqPath := this.Ctx.Request.RequestURI[1:]
	if strings.HasSuffix(reqPath, ".jpg") ||
		strings.HasSuffix(reqPath, ".png") ||
		strings.HasSuffix(reqPath, ".gif") {
		f, err := os.Open(path.Join("content", reqPath))
		if err != nil {
			beego.Error(err)
			return
		}
		defer f.Close()

		io.Copy(this.Ctx.ResponseWriter, f)
		return
	}

	this.TplNames = "home.html"
	this.Data["IsSinglePost"] = true
	this.Data["RecentArchives"] = models.GetRecentPosts()
	arch := models.GetSinglePost(this.Ctx.Request.RequestURI[1:])
	if arch == nil {
		this.Redirect("/", 302)
		return
	}
	this.Data["SinglePost"] = arch
}
