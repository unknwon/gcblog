---
Author: chai2010
Date: 2014-1-7
---

# Go语言的国际化支持(基于 gettext-go)


## hello, world!

假设有以下的程序, 输出: "Hello, world!".

	package main
	
	import (
		"fmt"
	)
	
	func main() {
		fmt.Println("Hello, world!")
	}

现在要让改程序支持不同语言的用户, 然后以本地语言输出相同意思的信息.
这就是很多程序面临的国际化问题.


## Go 语言的国际化思路

程序的国际化一般涉及到编码和翻译两个概念. 其中编码一般采用 UTF8 编码标准, Go 语言已经完美支持. 而目前常见翻译技术是 Qt 的 `tr` 函数和 GNU gettext 提供的 `gettext` 函数, 另外微软的 `MFC` 也有自己的多国语言支持方式.

Go 语言目前还没有标准的多国语言翻译方式. 不过笔者已经初步将 `gettext` 的运行时环境移植到了 Go 语言(采用纯 Go 实现, 无其他依赖).
Go 语言版的 `gettext` 名字为 `gettext-go`,
项目地址在: [http://code.google.com/p/gettext-go](http://code.google.com/p/gettext-go).

[gettext-go](http://code.google.com/p/gettext-go) 同时也借鉴了 Qt 的翻译上下文特性. 在 GNU gettext 的 `po` 和 `mo` 翻译文件中都是含有 `msgctxt` 上下文信息的, 但是 C/C++ 的翻译接口函数并没有上下文的参数, 因此 传统的 `gettext` 函数没有设置上下文的参数.


可以去 [godoc.org](http://godoc.org/code.google.com/p/gettext-go/gettext) 或 [gowalker.org](http://gowalker.org/code.google.com/p/gettext-go/gettext) 查看 
[gettext-go](http://code.google.com/p/gettext-go) 的文档.


## Go 语言的多国语言支持

基于 [gettext-go](http://code.google.com/p/gettext-go) , 我们可以很容易给 Go 程序增加多国语言的支持:

	package main
	
	import (
		"fmt"
	
		"code.google.com/p/gettext-go/gettext"
	)
	
	func main() {
		gettext.BindTextdomain("hello", "local")
		gettext.Textdomain("hello")
	
		fmt.Println(gettext.Gettext("Hello, world!"))
	}

其中 `gettext.BindTextdomain` 是绑定翻译的空间, 其中 `"hello"` 是对应翻译一类信息的翻译, `"local"` 为翻译文件的所在路径(这里当前目录下的"local")子目录.

按照 GNU gettext 的习惯, 简体中文对应的翻译文件为 `"local/zh_CN/LC_MESSAGES/hello.mo"`. 不同语言的命名有一个国际规范, 比如繁体中文对应 `"zh_TW"`, 美国英文对应 `"en_US"` 等等. 但是 [gettext-go](http://code.google.com/p/gettext-go) 对名字并没有特殊的要求.

`gettext.BindTextdomain` 可以绑定多个翻译空间, 但是同一个时刻只能使用一个翻译空间.

这里我们使用 `gettext.Textdomain` 指定当前的翻译空间为 `"hello"` .

运行新的程序程序, 发现输出还是: "Hello, world!".

这是因为缺少翻译文件...

## 生成翻译文件

未来, [gettext-go](http://code.google.com/p/gettext-go) 会开发一个  GNU gettext 工具集 中 的 `xgettext` 类似工具, 用于从程序中提取要翻译的字符串.

不过目前, 我们只能手工支持翻译文件了(还好这个例子只有一个字符串需要翻译).

创建 `"local/zh_CN/LC_MESSAGES/hello.po"` 文件, 内容如下:

	msgid ""
	msgstr ""
	
	msgctxt "main.main"
	msgid "Hello, world!"
	msgstr "你好, 世界!"

保存为**UTF8**编码格式.

然后用 GNU gettext 工具集中的 `msgfmt` 命令将 `hello.po` 文件编译为 `hello.mo` 文件:

	msgfmt -o hello.mo hello.po

如果是 Windows 用户, 可以下载 [poedit](http://www.poedit.net/) 翻译工具.
然后用 [poedit](http://www.poedit.net/) 打开 `hello.po` 文件,
点击保存后会自动生成 `hello.mo` 文件(也是 [poedit](http://www.poedit.net/) 的bin目录下自带的`msgfmt` 命令生成的).

重新运行新的程序程序, **还是输出: "Hello, world!" ?**


## 本地的语言环境

在上一节, 我们已经制作了简体中文的翻译文件 `"local/zh_CN/LC_MESSAGES/hello.mo"`, 然后输出依然是英文.

这是因为 [gettext-go](http://code.google.com/p/gettext-go) 翻译时不仅要依赖
对应语言的翻译文件, 还需要知道要范围为哪种语言(和网上翻译类似, 需要知道翻译的目标语言).

如果没有指定翻译语言, [gettext-go](http://code.google.com/p/gettext-go) 会尝试获取本地的默认语言环境, 主要是通过检查 `$(LC_MESSAGES)` 和 `$(LANG)` 两个环境变量. 如果两个环境变量都没有设置, 那么默认是不进行翻译的.

我们设置环境变量后重新运行程序(Windows):

	set LANG=zh_CN
	go run hello.go

这里时候应该可以输出中文了.

## 动态切换语言

如果不想使用默认的本地语言环境, 也可以用 `gettext.SetLocale` 接口设置本地语言环境.

	func main() {
		gettext.SetLocale("zh_CN")
		gettext.BindTextdomain("hello", "local")
		gettext.Textdomain("hello")
	
		fmt.Println(gettext.Gettext("Hello, world!"))
	}

这样可以可以需要采用合适的语言翻译文件.

## 翻译的上下文

Go 语言版的 [gettext-go](http://code.google.com/p/gettext-go) 的每个 `gettext.Gettext` 调用都有一个隐含的上下文信息(如果想自己指定上下文可以使用 `gettext.PGettext`).

默认的上下文为包含 `gettext.Gettext` 调用的函数名称, 比如:

- 如果是 main 包的全局函数初始化调用, 则为 `main.init`
- 如果是 main 包的 init 函数调用, 则为 `main.init`
- 如果是 main 包的 main 函数调用, 则为 `main.main`
- 如果是 main 包中的闭包调用, 则为 `main.func`
- 如果是非 main 包的函数, 则还需要包含包的完全路径名

上下文对应 Go 的运行时调用者名称, 具体实现在这里: [caller.go](https://code.google.com/p/gettext-go/source/browse/gettext/caller.go) .

## 练习题

1. 给前面的程序增加 繁体/日文/韩文/克林贡语 等语言的支持
2. 增加一个 `-local` 参数, 用于设置本地语言
3. 提交改进建议或其他反馈意见
