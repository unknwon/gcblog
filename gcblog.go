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

// An open source project for Golang China blog.
package main

import (
	"github.com/astaxie/beego"

	"github.com/Unknwon/gcblog/controllers"
)

const (
	APP_VER = "0.1.0.1213"
)

func main() {
	beego.Info(beego.AppName, APP_VER)

	// Register routers.
	beego.Router("/", &controllers.HomeController{})
	beego.Router("/archives", &controllers.ArchiveController{})
	beego.Router("/:all", &controllers.PostController{})

	// Register template functions.
	beego.AddFuncMap("byte2str", byte2Str)

	beego.Run()
}

func byte2Str(data []byte) string {
	return string(data)
}
