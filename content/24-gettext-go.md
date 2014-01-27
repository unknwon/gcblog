---
Author: chai2010
Date: 2014-1-24
---

# Go 语言的国际化支持(资源文件翻译)

在之前的 [Go语言的国际化支持(基于gettext-go)](http://my.oschina.net/chai2010/blog/190914) 中, 讲到了如何翻译源代码中的字符串.

项目地址在: [http://code.google.com/p/gettext-go](http://code.google.com/p/gettext-go). 文档在 [godoc.org](http://godoc.org/code.google.com/p/gettext-go/gettext) 或 [gowalker.org](http://gowalker.org/code.google.com/p/gettext-go/gettext) .

根据评论的反馈(@羊半仙), 之前版本的缺少对资源文件的支持.

最近对 [gettext-go](http://code.google.com/p/gettext-go) 做了一些改进, 主要涉及以下几点:

- 支持资源文件的翻译
- 支持zip格式的翻译文件

[more]


## 资源文件的翻译

资源文件的翻译比字符串更简单:

	import (
		"code.google.com/p/gettext-go/gettext"
	)
	
	func main() {
		gettext.SetLocale("zh_CN")
		gettext.Textdomain("hello")
	
		// translate resource
		fmt.Println(string(gettext.Getdata("poems.txt")))
		// Output: ...
	}

输出的内容是 李白 诗歌 [<月下独酌>](http://code.google.com/p/gettext-go/source/browse/examples/local/zh_CN/LC_RESOURCE/hello/poems.txt):

- 简体中文版: [<月下独酌>](http://code.google.com/p/gettext-go/source/browse/examples/local/zh_CN/LC_RESOURCE/hello/poems.txt)
- 繁体中文版: [<月下独酌>](http://code.google.com/p/gettext-go/source/browse/examples/local/zh_TW/LC_RESOURCE/hello/poems.txt)
- 英文翻译版: [<月下独酌>](http://code.google.com/p/gettext-go/source/browse/examples/local/default/LC_RESOURCE/hello/poems.txt)

## zip格式的翻译文件支持

可以将翻译文件目录打包为zip格式. 如果绑定的翻译文件不是一个目录, 而是一个文件, 则会当作zip文件处理.

	import (
		"code.google.com/p/gettext-go/gettext"
	)
	
	func main() {
		gettext.SetLocale("zh_CN")
		gettext.Textdomain("hello")
	
		gettext.BindTextdomain("hello", "local.zip", nil)
	
		...
	}

Zip文件在这里: [http://gettext-go.googlecode.com/hg/examples/local.zip](http://gettext-go.googlecode.com/hg/examples/local.zip)

如果希望将zip文件嵌入到程序中, 可以先用工具将zip文件转换为`[]byte`格式的数据, 然后绑定到翻译域中:


	import (
		"code.google.com/p/gettext-go/gettext"
	)

	// 根据 local.zip 生成
	var local_zip_data = []byte{
		0x00, 0x01, 0x02, 0x03, ...
	}
	
	func main() {
		gettext.SetLocale("zh_CN")
		gettext.Textdomain("hello")
	
		gettext.BindTextdomain("hello", "embeded_local", local_zip_data)
	
		...
	}

如果 `gettext.BindTextdomain` 的第三个参数不为 `nil`, 则会将该参数传入的数据作为zip数据处理.

## 翻译目录的结构

[gettext-go](http://code.google.com/p/gettext-go) 包的翻译文件涉及以下几个概念:

- **local**: 本地采用的语言. 默认值可以由`$(LC_MESSAGES)`或`$(LANG)`环境变量指定.
- **domain** : 翻译的名字空间, 类似文本编辑器中的配色风格的概念,  `gettext.BindTextdomain` 的第一个参数.
- **domain_path**: **domain** 对应的目录路径, 也可能是zip文件路径或zip数据名,  `gettext.BindTextdomain` 的第二个参数.
- **domain_data**: **domain** 对应的zip数据(可以为空), `gettext.BindTextdomain` 的第二个参数.

不管是zip还是本地目录, 内部的目录组织结构是一致的:

	Root: "path" or "file.zip/zipBaseName"
	 +-default                 # local: $(LC_MESSAGES) or $(LANG) or "default"
	 |  +-LC_MESSAGES            # just for `gettext.Gettext`
	 |  |   +-hello.mo             # $(Root)/$(local)/LC_MESSAGES/$(domain).mo
	 |  |   \-hello.po             # $(Root)/$(local)/LC_MESSAGES/$(domain).mo
	 |  |
	 |  \-LC_RESOURCE            # just for `gettext.Getdata`
	 |      +-hello                # domain map a dir in resource translate
	 |         +-favicon.ico       # $(Root)/$(local)/LC_RESOURCE/$(domain)/$(filename)
	 |         \-poems.txt
	 |
	 \-zh_CN                   # simple chinese translate
	    +-LC_MESSAGES
	    |   +-hello.mo             # try "$(domain).mo" first
	    |   \-hello.po             # try "$(domain).po" second
	    |
	    \-LC_RESOURCE
	        +-hello
	           +-favicon.ico       # try "$(local)/$(domain)/file" first
	           \-poems.txt         # try "default/$(domain)/file" second

其中 `$(Root)/$(local)/LC_MESSAGES/` 对应 `gettext.Gettext` 的翻译字符串, `$(Root)/$(local)/LC_RESOURCE/$(domain)/?` 对应 `gettext.Getdata` 要翻译的资源文件.

翻译字符串时, 会先尝试`mo`格式的二进制翻译文件, 如果失败则会继续尝试`po`格式的原始翻译文件, 如果依然失败则返回原字符串.

翻译资源文件时, 如果资源文件缺失会继续尝试加载Local为`default`的资源文件, 如果依然失败则返回`nil`.

## 展望

目前 [gettext-go](http://code.google.com/p/gettext-go) 的运行时已经初步完备, 可以支持字符串和资源文件的翻译. 不过 [gettext-go](http://code.google.com/p/gettext-go) 的辅助工具依然不足, 特别是缺少可以从Go程序中自动提取字符串的`xgettext`工具. 下一步会考虑实现一个针对Go语言的`xgettext`工具.

在Go语言中还有一类比较特殊的字符串资源: 用于godoc显示的文档信息(类似的还有blog文章等). 如果`xgettext`工具能够支持Go语言文档的信息的提取和翻译后文档的合并, 那么文档的翻译将会方便很多.

如果有感兴趣的同学可以一起合作完善.
