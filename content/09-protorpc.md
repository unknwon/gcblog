---
Author: chai2010
Date: 2014-1-9
---

# Go 语言的 RPC 介绍

## 标准库的 RPC

RPC 是远程调用的简称, 简单的说就是要调用本地函数一样调用服务器的函数.

Go 语言的标准库已经提供了 RPC 框架和不同的 RPC 实现.

下面是一个服务器的例子:

[more]

	type Echo int
	
	func (t *Arith) Hi(args string, reply *string) error {
		*reply = "echo:" + args
		return nil
	}
	
	func main() {
		rpc.Register(new(Echo))
		rpc.HandleHTTP()
		l, e := net.Listen("tcp", ":1234")
		if e != nil {
			log.Fatal("listen error:", e)
		}
		http.Serve(l, nil)
	}

其中 `rpc.Register` 用于注册 RPC 服务, 默认的名字是对象的类型名字(这里是 `Echo`). 如果需要指定特殊的名字, 可以用 `rpc.RegisterName` 进行注册.

被注册对象的类型所有满足以下规则的方法会被导出到 RPC 服务接口:

	func (t *T) MethodName(argType T1, replyType *T2) error

被注册对应至少要有一个方法满足这个特征, 否则可能会注册失败.

然后 `rpc.HandleHTTP` 用于指定 RPC 的传输协议, 这里是采用 http 协议作为 RPC 调用的载体. 用户也可以用 `rpc.ServeConn` 接口, 定制自己的传输协议. 

客户端可以这样调用 `Echo.Hi` 接口:

	func main() {
		client, err := rpc.DialHTTP("tcp", "127.0.0.1:1234")
		if err != nil {
			log.Fatal("dialing:", err)
		}
	
		var args = "hello rpc"
		var reply string
		err = client.Call("Echo.Hi", args, &reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		fmt.Printf("Arith: %d*%d=%d\n", args.A, args.B, reply)
	}

客户端先用 `rpc.DialHTTP` 和 RPC 服务器进行一个链接(协议必须匹配).

然后通过返回的 `client` 对象进行远程函数调用. 函数的名字是由
`client.Call` 第一个参数指定(是一个字符串).

基于 HTTP 的 RPC 调用一般是在调整时使用, 默认可以通过浏览 `"127.0.0.1:1234/debug/rpc"` 页面查看 RPC 的统计信息.

## 基于 JSON 的 RPC 调用

在上面的 RPC 例子中, 我们采用了默认的 HTTP 协议作为 RPC 调用的传输载体.

除了传输协议, 还有可以指定一个 RPC 编码协议, 用于编码/节目 RPC 调用的函数参数和返回值. RPC 调用不指定编码协议时, 默认采用 Go 语言特有的 `gob` 编码协议.

因为, 其他语言一般都不支持 Go 语言的 `gob` 协议, 因此如果需要跨语言 RPC 调用就需要
采用通用的编码协议.

Go 的标准库还提供了一个 `"net/rpc/jsonrpc"` 包, 用于提供基于 JSON 编码的 RPC 支持.

服务器部分只需要用 `rpc.ServeCodec` 指定 JSON 编码协议就可以了:


	func main() {
		lis, err := net.Listen("tcp", ":1234")
		if err != nil {
			return err
		}
		defer lis.Close()
	
		srv := rpc.NewServer()
		if err := srv.RegisterName("Echo", new(Echo)); err != nil {
			return err
		}
	
		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Fatalf("lis.Accept(): %v\n", err)
			}
			go srv.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}

客户端部分值需要用 `jsonrpc.Dial` 代替 `rpc.Dial` 就可以了:

	func main() {
		client, err := jsonrpc.DialHTTP("tcp", "127.0.0.1:1234")
		if err != nil {
			log.Fatal("dialing:", err)
		}
		...
	}


如果需要在其他语言中使用 `jsonrpc` 和 Go 语言进行通讯, 需要封装一个和 `jsonrpc`
匹配的库.

关于 `jsonrpc` 的实现细节这里就不展开讲了, 感兴趣的话可以参考这篇文章: [JSON-RPC: a tale of interfaces](http://blog.golang.org/json-rpc-tale-of-interfaces).

## 基于 Protobuf 的 RPC 调用

[Protobuf](http://code.google.com/p/protobuf/;) 是 Google 公司开发的编码协议. 它的优势是编码后的数据体积比较小(并不是压缩算法), 比较适合用于命令的传输编码. 

Protobuf 官方团队提供 Java/C++/Python 几个语言的支持, Go语言的版本由
Go团队提供支持, 其他语言由第三方支持.

Protobuf 的语言规范中可以定义 RPC 接口. 但是在 Go 语言和 C++ 版本的 Protobuf 中都
没有生成 RPC 的实现.

不过笔者在 Go 语言的版本 Protobuf 基础上开发了 RPC 的实现 [protorpc](https://code.google.com/p/protorpc/), 同时提供的 `protoc-gen-go`
命令可以生成相应的 RPC 代码.

要使用 [protorpc](https://code.google.com/p/protorpc/), 需要先在 proto 文件定义接口(`arith.pb/arith.proto`):

	package arith;

	// go use cc_generic_services option
	option cc_generic_services = true;

	message ArithRequest {
		optional int32 a = 1;
		optional int32 b = 2;
	}

	message ArithResponse {
		optional int32 val = 1;
		optional int32 quo = 2;
		optional int32 rem = 3;
	}

	service ArithService {
		rpc multiply (ArithRequest) returns (ArithResponse);
		rpc divide (ArithRequest) returns (ArithResponse);
	}

[protorpc](https://code.google.com/p/protorpc/) 使用 `cc_generic_services` 选择控制是否输出 RPC 代码. 因此, 需要设置 `cc_generic_services` 为 `true`.

然后下载 [protoc-2.5.0-win32.zip](https://code.google.com/p/protobuf/downloads/list), 解压后可以得到一个 `protoc.exe` 的编译命令.

然后使用下面的命令获取 [protorpc](https://code.google.com/p/protorpc/) 和对应的 `protoc-gen-go` 插件.

	go get code.google.com/p/protorpc
	go get code.google.com/p/protorpc/protoc-gen-go

需要确保 `protoc.exe` 和 `protoc-gen-go.exe` 都在 `$PATH` 中.
然后运行以下命令将前面的接口文件转换为 Go 代码:

	cd arith.pb && protoc --go_out=. arith.proto

新生成的文件为 `arith.pb/arith.pb.go`.

下面是基于 Protobuf-RPC 的服务器:

	package main

	import (
		"errors"

		"code.google.com/p/goprotobuf/proto"

		"./arith.pb"
	)

	type Arith int

	func (t *Arith) Multiply(args *arith.ArithRequest, reply *arith.ArithResponse) error {
		reply.Val = proto.Int32(args.GetA() * args.GetB())
		return nil
	}

	func (t *Arith) Divide(args *arith.ArithRequest, reply *arith.ArithResponse) error {
		if args.GetB() == 0 {
			return errors.New("divide by zero")
		}
		reply.Quo = proto.Int32(args.GetA() / args.GetB())
		reply.Rem = proto.Int32(args.GetA() % args.GetB())
		return nil
	}

	func main() {
		arith.ListenAndServeArithService("tcp", ":1984", new(Arith))
	}

其中导入的 `"./arith.pb"` 的名字为 `arith`, 在 `arith.pb/arith.proto` 文件中定义(这 2 个可能不同名, 导入时要小心).

`arith.ArithRequest` 和 `arith.ArithResponse` 是RPC接口的输入和输出参数,
也是在在 `arith.pb/arith.proto` 文件中定义的.

同时生成的还有一个 `arith.ListenAndServeArithService` 函数, 用于启动 RPC 服务.
改函数的第三个参数是 RPC 的服务对象, 必须要满足 `arith.EchoService` 接口的定义.

客户端的使用也很简单, 只要一个 `arith.DialArithService` 就可以链接了:

	stub, client, err := arith.DialArithService("tcp", "127.0.0.1:1984")
	if err != nil {
		log.Fatal(`arith.DialArithService("tcp", "127.0.0.1:1984"):`, err)
	}
	defer client.Close()

`arith.DialArithService` 返回了一个 `stub` 对象, 改对象已经绑定了 RPC 的各种方法, 可以直接调用(不需要用字符串指定方法名字):

	var args ArithRequest
	var reply ArithResponse

	args.A = proto.Int32(7)
	args.B = proto.Int32(8)
	if err = stub.Multiply(&args, &reply); err != nil {
		log.Fatal("arith error:", err)
	}
	fmt.Printf("Arith: %d*%d=%d", args.GetA(), args.GetB(), reply.GetVal())

相比标准的RPC的库, [protorpc](https://code.google.com/p/protorpc/) 由以下几个有点:

1. 采用标准的 Protobuf 协议, 便于和其他语言交互
2. 自带的 `protoc-gen-go` 插件可以生成 RPC 的代码, 简化使用
3. 服务器注册和调用客户端都是具体类型而不是字符串和 `interface{}`, 这样可以由编译器保证安全
4. 底层采用了 `snappy` 压缩传输的数据, 提高效率

不足之处是使用流程比标准 RPC 要繁复(需要将 proto 转换为 Go 代码).

## C++ 调用 Go 提供的 Protobuf-RPC 服务

[protorpc](https://code.google.com/p/protorpc/) 同时也提供了 C++ 语言的实现.

C++ 版本的安装如下:

1. `hg clone https://code.google.com/p/protorpc.cxx/`
2. `cd protorpc.cxx`
3. build with cmake

下面是 C++ 的客户端链接 Go 语言版本的 服务器:

	#include "./service.pb/arith.pb.h"
	
	#include <google/protobuf/rpc/rpc_server.h>
	#include <google/protobuf/rpc/rpc_client.h>
	
	int main() {
	  ::google::protobuf::rpc::Client client("127.0.0.1", 1234);
	
	  service::ArithService::Stub arithStub(&client);
	
	  ::service::ArithRequest arithArgs;
	  ::service::ArithResponse arithReply;
	  ::google::protobuf::rpc::Error err;
	
	  // EchoService.mul
	  arithArgs.set_a(3);
	  arithArgs.set_b(4);
	  err = arithStub.multiply(&arithArgs, &arithReply);
	  if(!err.IsNil()) {
	    fprintf(stderr, "arithStub.multiply: %s\n", err.String().c_str());
	    return -1;
	  }
	  if(arithReply.c() != 12) {
	    fprintf(stderr, "arithStub.multiply: expected = %d, got = %d\n", 12, arithReply.c());
	    return -1;
	  }
	
	  printf("Done.\n");
	  return 0;
	}

C++ 完成的服务器和客户端的例子请参考: [rpcserver.cc](http://code.google.com/p/protorpc/source/browse/tests/rpctest/rpcserver.cc?repo=cxx)
和 [rpcclient.cc](http://code.google.com/p/protorpc/source/browse/tests/rpctest/rpcclient.cc?repo=cxx)

## 总结

Go 语言的 RPC 客户端是一个使用简单, 而且功能强大的RPC库. 基于标准的 RPC 库我们可以方便的定制自己的 RPC 实现(传输协议和串行化协议都可以定制).

不过在开发 [protorpc](https://code.google.com/p/protorpc/) 的过程中也发现了 `net/rpc` 包的一些不足之处:

- 内置的 `HTTP` 协议的 RPC 的串行化协议和传输协议耦合过于紧密, 用户扩展的协议无法支持内置的 `HTTP` 传输协议(因为 `rpc.Server` 和 `rpc.Client` 接口缺陷导致的问题)
- `rpc.Server` 只能注册 `rpc.ServerCodec`, 而不能注册工厂函数. 而 `jsonrpc.NewServerCodec` 需要依赖先建立链接(`conn` 参数), 这样导致了 `HTTP` 协议只能支持内置的 `gob` 协议
- `rpc.Client` 的问题和 `rpc.Server` 类似

因为 Go1 需要保证 API 的兼容性, 因此上述的问题只能希望在未来的 Go2 能得到改善.


