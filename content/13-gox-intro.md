---
Author: skyblue
Date: 2014-1-13
---

# gox 入门教程

先说下交叉编译是什么？
交叉编译也就是你可以在 linux 上编译出可以在 windows 上运行的程序，在 32 位系统编译出 64 位系统运行的程序。

gox 就是方便你使用 golang 的交叉编译的工具。

[more]

### 安装 gox

首先你的机器上需要装有 golang。配置好了 GOROOT,GOPATH 这两个环境变量。
我机器上的配置是这个样子（仅供参考）

	export GOROOT=$HOME/go
	export GOPATH=$HOME/goproj
	export GOBIN=$GOPATH/bin
	export PATH=$PATH:$GOBIN

安装 gox 其实很简单（只需要 2 步）。

	go get github.com/mitchellh/gox
	
之后命令行输入 `gox -h`，应该会用输出的，不然你要检查下 PATH 变量设置的是否正确。

下一步需要编译出其他平台需要的库。这步有一点慢，要有点耐心。

	gox -build-toolchain

输出大概是这个样子

	The toolchain build can't be parallelized because 	compiling a single
	Go source directory can only be done for one platform at a time. Therefore,
	the toolchain for each platform will be built one at a time.

	--> Toolchain: darwin/386
	--> Toolchain: darwin/amd64
	...


当这一步完成时，gox 已经可以开始能用了。

### 使用 gox（简单入门）

下面我们来体验一下 gox 的强大。
需要注意的是 gox 没法指定一个文件进行编译的。

为方便起见，我们先到到 `$GOPATH/src` 下，建立一个 hello 文件夹。随便写个 hello.go 程序。比如

	package main
	func main() {
		println("hello world")
	}

进入到程序目录中，直接运行 gox。程序会一口气生成 17 个文件。横跨 windows,linux,mac,freebsd,netbsd 五大操作系统。以及 3 种了下的处理器(386、amd64、arm)
关于处理器的介绍可以看看这个 <http://www.361way.com/cpuinfo/1510.html>
arm 类型的处理器，在手机上用的比较多。

### 使用 gox（指定生成的平台(OS)和处理器(ARCH)）

很多的选项其实 `gox -h` 的帮助都可以查的很清楚。

如果我们想生成 linux 和 windows 上的程序，只要通过一下命令：

	gox -os "windows linux" -arch amd64

目录下你就能看到生成出来的两个程序
	
	hello_linux_amd64
	hello_windows_amd64.exe

也可以这样用,效果与刚才的命令等价

	gox -osarch "windows/amd64 linux/amd64"
	
### 进阶

还可以继续学习的东西 [goxc](https://github.com/laher/goxc), 该工具封装了 gox 提供了更为强大的功能。

### 可能遇到的问题

1. 交叉编译暂时还不支持 CGO（估计再过 1 年也不会支持）

### 相关资源

* [跨平台编译 go 程序](http://studygolang.com/topics/21)
* [如何使用 golang 的交叉编译](http://www.cnblogs.com/ghj1976/archive/2013/04/19/3030703.html)
* [golang 在线编译](http://build.golangtc.com)
