---
Author: 星星
Date: 2014-1-19
---

# go build 命令是如何工作的？

> 本文为转载技术翻译，原翻译地址：http://mikespook.com/2013/11/%E7%BF%BB%E8%AF%91-go-build-%E5%91%BD%E4%BB%A4%E6%98%AF%E5%A6%82%E4%BD%95%E5%B7%A5%E4%BD%9C%E7%9A%84%EF%BC%9F/

> 原
文地址：http://dave.cheney.net/2013/10/15/how-does-the-go-build-command-work

本文以 Go 的标准库为例，介绍了 Go 编译过程的工作原理。

## gc 工具链

本文将关注 gc 工具链。gc 工具链的名字来自 Go 的前端编译器 cmd/gc，这主要是为了与 gccgo 工具链进行区分。当人们讨论 Go 编译器的时候，多半是指 gc 工具链。本文不关注 gccgo 工具链。

gc 工具链是直接从 Plan 9 的工具链剥离出来的。该工具链由一个 Go 编译器、一个 C 编译器、一个汇编器和一个链接器组成。可以在 Go 代码的 src/cmd/ 子目录找到这些工具，包括在所有实现下共用的前端，和在不同处理器架构下特定的后端。后端用特定的字母标识，这也是 Plan 9 的一个传统。命令包括：

5g、6g 和 8g 是 .go 文件的对应 arm、amd64 和 386 的编译器；
5c、6c 和 8c 是 .c 文件的对应 arm、amd64 和 386 的编译器；
5a、6a 和 8a 是 .s 文件的对应 arm、amd64 和 386 的编译器；
5l、6l 和 8l 是用于上面命令产生的文件的链接器，同样对应 arm、amd64 和 386。

需要注意的是，每个命令都可以在任何被支持的平台上进行编译，这也是 Go 的交叉编译能力的一种体现。你可以在[这篇文章](http://dave.cheney.net/2013/07/09/an-introduction-to-cross-compilation-with-go-1-1)里了解到更多关于交叉编译的情况。

[more]
## 构建包

构建一个 Go 包至少包含两个步骤，编译 .go 文件，然后将编译结果打包。考虑到 crypto/hmac 很小，只有一个源文件和测试文件，所以用其举例说明。使用 -x 选项以告诉 go build 打印出它所执行的每一步

	% go build -x crypto/hmac
	WORK=/tmp/go-build249279931
	mkdir -p $WORK/crypto/hmac/_obj/
	mkdir -p $WORK/crypto/
	cd /home/dfc/go/src/pkg/crypto/hmac
	/home/dfc/go/pkg/tool/linux_arm/5g -o $WORK/crypto/hmac/_obj/_go_.5 -p crypto/hmac -complete -D _/home/dfc/go/src/pkg/crypto/hmac -I $WORK ./hmac.go
	/home/dfc/go/pkg/tool/linux_arm/pack grcP $WORK $WORK/crypto/hmac.a $WORK/crypto/hmac/_obj/_go_.5

逐一了解这些步骤

	WORK=/tmp/go-build249279931
	mkdir -p $WORK/crypto/hmac/_obj/
	mkdir -p $WORK/crypto/

go build 创建了一个临时目录 /tmp/go-build249279931 并且填充一些框架性的子目录用于保存编译的结果。第二个 mkdir 可能是多余的，已经创建了 issue 6538 来跟踪这个问题。

	cd /home/dfc/go/src/pkg/crypto/hmac
	/home/dfc/go/pkg/tool/linux_arm/5g -o $WORK/crypto/hmac/_obj/_go_.5 -p crypto/hmac -complete -D _/home/dfc/go/src/pkg/crypto/hmac -I $WORK ./hmac.go

go 工具切换 crypto/hmac 的源代码目录，并且调用架构对应的 go 编译器，在本例中是 5g。实际上是没有 cd 的，当 5g 执行时 /home/dfc/go/src/pkg/crypto/hmac 作为 exec.Command.Dir 的参数传递。这意味着为了让命令行更加精简，.go 源文件可以使用对应其源代码目录的相对路径。

编译器生成唯一的一个临时文件 $WORK/crypto/hmac/_obj/_go_.5 将在最后一步中使用。

	/home/dfc/go/pkg/tool/linux_arm/pack grcP $WORK $WORK/crypto/hmac.a $WORK/crypto/hmac/_obj/_go_.5

最后一步是打包目标文件到将被链接器和编译器使用的归档文件 .a 中。

由于在包上调用了 go build ，$WORK 中的结果将在编译结束后删除。如果调用 go install -x 将会输出额外的两行

	mkdir -p /home/dfc/go/pkg/linux_arm/crypto/
	cp $WORK/crypto/hmac.a /home/dfc/go/pkg/linux_arm/crypto/hmac.a

这演示了 go build 与 install 的不同之处； build 构建，install 构建并且安装，以便用于其他构建。

## 构建更加复杂的包

你可能正在思考前面例子中的打包步骤。这里的编译器和链接器仅接受了单一的文件作为包的内容，如果包含多个目标文件，在使用它们之前，必须将其打包到一个单一的 .a 归档文件中。

cgo 是一个常见的产生超过一个中间目标文件的例子，不过它对于本文来说太过复杂了，这里用包含 .s 汇编文件的情况作为替代的例子，例如 crypto/md5。

	% go build -x crypto/md5
	WORK=/tmp/go-build870993883
	mkdir -p $WORK/crypto/md5/_obj/
	mkdir -p $WORK/crypto/
	cd /home/dfc/go/src/pkg/crypto/md5
	/home/dfc/go/pkg/tool/linux_amd64/6g -o $WORK/crypto/md5/_obj/_go_.6 -p crypto/md5 -D _/home/dfc/go/src/pkg/crypto/md5 -I $WORK ./md5.go ./md5block_decl.go
	/home/dfc/go/pkg/tool/linux_amd64/6a -I $WORK/crypto/md5/_obj/ -o $WORK/crypto/md5/_obj/md5block_amd64.6 -D GOOS_linux -D GOARCH_amd64 ./md5block_amd64.s
	/home/dfc/go/pkg/tool/linux_amd64/pack grcP $WORK $WORK/crypto/md5.a $WORK/crypto/md5/_obj/_go_.6 $WORK/crypto/md5/_obj/md5block_amd64.6

这个例子执行在 linux/amd64 主机上，6g 被调用以编译两个 .go 文件：md5.go 和 md5block_decl.go。后面这个包含一些用汇编实现的函数的声明。

这时 6a 被调用以汇编 md5block_amd64.s。选择哪个 .s 来编译的逻辑在我之前的关于条件编译的文章中进行了说明。

最后调用 pack 来打包 Go 的目标文件 _go_.6，以及汇编目标文件 md5block_amd64.6 到单一的一个归档文件中。

## 构建命令

一个 Go 命令是一个命名为 main 的包。main 包，或者说命令的编译方式与其他包一致，不过它们会在内部经过一些额外的步骤来链接成最终的可执行文件。让我们通过 cmd/gofmt 来研究一下这个过程

	% go build -x cmd/gofmt
	WORK=/tmp/go-build979246884
	mkdir -p $WORK/cmd/gofmt/_obj/
	mkdir -p $WORK/cmd/gofmt/_obj/exe/
	cd /home/dfc/go/src/cmd/gofmt
	/home/dfc/go/pkg/tool/linux_amd64/6g -o $WORK/cmd/gofmt/_obj/_go_.6 -p cmd/gofmt -complete -D _/home/dfc/go/src/cmd/gofmt -I $WORK ./doc.go ./gofmt.go ./rewrite.go ./simplify.go
	/home/dfc/go/pkg/tool/linux_amd64/pack grcP $WORK $WORK/cmd/gofmt.a $WORK/cmd/gofmt/_obj/_go_.6
	cd .
	/home/dfc/go/pkg/tool/linux_amd64/6l -o $WORK/cmd/gofmt/_obj/exe/a.out -L $WORK $WORK/cmd/gofmt.a
	cp $WORK/cmd/gofmt/_obj/exe/a.out gofmt

前面的六行应该已经熟悉了，main 包会向其他 Go 包一样编译和打包。

不同之处在于倒数第二行，调用链接器生成二进制的可执行文件。

	/home/dfc/go/pkg/tool/linux_amd64/6l -o $WORK/cmd/gofmt/_obj/exe/a.out -L $WORK $WORK/cmd/gofmt.a

最后一行复制并重命名编译后的二进制文件到其最终的位置和名称。如果使用了 go install，二进制也会被复制到 $GOPATH/bin（如果设置了 $GOBIN，则为 $GOBIN）。

## 历史原因

如果你回到足够久远的时代，回到 go tool 之前，回到 Makefiles 的时代，你可以找到 Go 编译过程的核心。这个例子来自 release.r60 的文档

	$ cat >hello.go <<EOF
	package main
	
	import "fmt"
	
	func main() {
	        fmt.Printf("hello, world\n")
	}
	EOF
	$ 6g hello.go
	$ 6l hello.6
	$ ./6.out
	hello, world

也就是这些，6g 编译了 .go 文件到 .6 目标文件，6l 链接目标文件以及 fmt （和运行时）包来生成二进制文件 6.out。

## 结束语

在本文中，我们讨论了 go build 的工作原理，并了解了 go install 对编译结果处理的不同方式。

现在你已经知道 go build 是如何工作的了，以及如何通过 -x 展示编译过程，可以尝试传递这个标识到 go test 并且观察其结果。

另外，如果已经在系统中安装了 gccgo，可以向 go build 传递 -compiler gccgo，然后使用 -x 来了解 Go 代码是如何用这个编译器进行编译的。