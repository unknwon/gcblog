---
Author: 喻恒春
Date: 2014-1-8
---

# Martini 极好的 Go Web 框架

已知的其他框架看到的是传统 OOP 的影子, 到处充蚀 Class 风格的 OOP 方法. 而我们知道 GoLang 中是没有 Class 的. 笔者也曾努力用 Go 的风格做 WEB 开发, 总感到力不从心. 写出的代码不能完全称之为框架, 到更像一个拷贝源码使用的应用. 要达到灵活需要修改源码. 直到看到了 [Martini](https://github.com/codegangsta/martini). 纯 GoLang 风格的框架出现了.

## 核心 Injector

[Injector](https://github.com/codegangsta/inject) 是 [Martini](https://github.com/codegangsta/martini) 的核心. 其代码非常简洁. 功能仅仅是通过反射包, 对函数进行参数类型自动匹配进行调用. 
笔者曾经为完成类似的功能写了 [typeless](https://github.com/gohub/typeless), 这是一个繁琐的高成本的实验品. Injector 把事情简单化了, Injector 假设函数的参数都具有不同的类型. 在 WEB 开发中的 HandlerFunc 通常都具有这样的形式. 因此通过反射包可以对参数进行自动的匹配并调用 HandlerFunc, 当然事先要把所有可能使用到的参数 Map/MapTo 给 Injector 对象, 这很容易而且是可以预见的.

## 简洁的路由设计

Martini 的路由 [router.go](https://github.com/codegangsta/martini/blob/master/router.go) 写的非常简洁实用, 可见作者使用正则的功力非常深厚. 
举例:

	/wap/:category/pow/**/:id

匹配: 

	/wap/Golang/pow/Path1/ToPathN/Foo”\
	
## 灵活的中间件

这里的中间件泛指应用需求中的流程控制, 预处理, 过滤器, 捕获 panic, 日志等等.这些依然是 Injector 在背后提供动力.

举例源代码中的 ClassicMartini ,当然你可以按需求模仿一个.

```go
func Classic() *ClassicMartini {
    r := NewRouter()
    m := New() // 
    m.Use(Logger()) // 启用日志
    m.Use(Recovery()) // 捕获 panic
    m.Use(Static("public")) // 静态文件
    m.Action(r.Handle) // 最后一个其实是执行了默认的路由机制
    return &ClassicMartini{m, r}
}
```

其中 Recovery() 和 func (r *routeContext) run() 配合的方法非常值得读一下, 一个简单的计数器就完成了.

[martini-contrib](https://github.com/codegangsta/martini-contrib) 中的 web 包展示了如何通过 Martini.Use 接口进行中间件的设计. 加入新的 HandlerFunc 参数类型就是通过 Map 完成的.

假如我们要完成对 Response 流程的控制, 达不到某个条件中断 Use 和 Action 设置的Handler. 那可以简单的通过 Use 加入自己的 Recovery 和判断 Handler 实现

```go
func Recovery() Handler {
    return func(res http.ResponseWriter, c Context, logger *log.Logger) {
        defer func() {
            if err := recover(); err != nil {
                s, ok := err.(string) // 示意用 string,你可以定义类型
                if ok && s == "dont ServerError" {
                        return
                }
                res.WriteHeader(http.StatusInternalServerError)
                logger.Printf("PANIC: %s\n%s", err, debug.Stack())
            }
        }()
        c.Next()
    }
}
func MyHandler() Handler {
    return func(req *http.Request) {
        if something {
           panic("dont ServerError")
        }
    }
}
m.Use(Recovery())
m.Use(MyHandler())
```

甚至

```go
// 支持 Handlers 返回值输出
m.Get("/", func() string {
  return "hello world" // HTTP 200 : "hello world"
})
// 向 handlers 传递数据库对象
db := &MyDatabase{}
m := martini.Classic()
m.Map(db) // the service will be available to all handlers as *MyDatabase
// ...
m.Run()
```

变化和自由度非常高. 正如 Martini 介绍的:

> Martini is a powerful package for quickly writing modular web applications/services in Golang.

martini 是一个极好的 web 框架，令人惊讶的简单，难以想象的高效。

**是的, 这才是真正的 GoLang 风格.**