---
Author: 无闻
Date: 2014-1-1
---

# 使用 Beego 与 Mgo 开发的示例程序

> 本文为技术翻译，原文地址（需翻墙）：http://www.goinggo.net/2013/12/sample-web-application-using-beego-and.html

## 简介

当我发现 beego 框架时感觉非常激动。我只用了大约 4 个小时就将一个现有的 Web 应用程序移植到了该框架上并做了一些端对端测试的调用扩展。我想要与你分享这个基于 beego 的站点。

我构建了一个具有以下功能的示例 Web 应用程序：

1. 实现了 2 个通过 [mgo](http://labix.org/mgo) 驱动拉取 MongoDB 数据的 Web 调用。
2. 使用 [envconfig](https://github.com/kelseyhightower/envconfig) 配置环境变量作为参数。
3. 通过 [goconvey](http://smartystreets.github.io/goconvey/) 书写测试用例。
4. 结合我的 [logging](https://github.com/goinggo/tracelog) 包。

这个示例的代码可以在 GoingGo 帐户下的 GitHub 仓库中找到：https://github.com/goinggo/beego-mgo。

你可以拉取并运行它。它使用了我在 MongoLab 创建的公开 MongoDB 数据库。你会需要安装 [git](https://help.github.com/articles/set-up-git) 和 [bazaar](http://bazaar.canonical.com/en/) 以保证能够使用 go get 来将它安装到你系统中。

	go get github.com/goinggo/beego-mgo
	
使用 zscripts 文件夹里的脚步可以快速运行或测试 Web 应用程序。

## Web 应用程序代码结构

让我们来看一下项目结构和不同文件夹的功能：

- app：包含 business、model 和 service 层。
- app/models：用于 business 和 service 层的数据结构。
- app/services：用于为不同服务提供基本函数。可以是针对数据库或 Web 调用的函数。
- app/business：被 controller 和处理 business 规则的函数所调用。多种服务的组合调用可用于实现更大层面上的功能。
- controllers：URL 或 Web API 调用的切入点。控制器会直接调用 business 层的函数来直接处理请求。
- routes：用于 URL 和控制器代码的映射。
- static：用于存放脚本、CSS 和图片等静态资源。
- test：可使用 go test 运行的测试用例。
- utilities：用于支持 Web 应用程序的代码。数据库操作和 panic 处理的样例与抽象化代码。
- views：用于存放模板文件。
- zscripts：帮助更方便构建、运行和测试 Web 应用程序的脚本。

## 控制器、模型和服务

这些层的代码组成了 Web 应用程序。我的框架背后的理念是尽可能的抽象化代码为样板。这就要求我实现一个 base 控制器包和 base 服务包。

### base 控制器包

base 控制器包由所有控制器所要求的默认抽象的控制器行为来组成：

```go
type (
    BaseController struct {
        beego.Controller
        services.Service
    }
)

func (this *BaseController) Prepare() {
    this.UserId = "unknown"
    tracelog.TRACE(this.UserId, "Before",
        "UserId[%s] Path[%s]", this.UserId, this.Ctx.Request.URL.Path)

    var err error
    this.MongoSession, err = mongo.CopyMonotonicSession(this.UserId)
    if err != nil {
        tracelog.ERRORf(err, this.UserId, "Before", this.Ctx.Request.URL.Path)
        this.ServeError(err)
    }
}

func (this *BaseController) Finish() {
    defer func() {
        if this.MongoSession != nil {
            mongo.CloseSession(this.UserId, this.MongoSession)
            this.MongoSession = nil
        }
    }()

    tracelog.COMPLETEDf(this.UserId, "Finish", this.Ctx.Request.URL.Path)
}
```

一个名为 `BaseController` 的新类型直接由类型 `beego.Controller` 和 `services.Service` 嵌入而组成。这样做就可以直接获得这两个类型的所有字段和方法给 `BaseController`，并直接操作这些字段和方法。

### 服务包

服务包用于样板化所有服务所要求的代码：

```go
type (
    Service struct {
        MongoSession *mgo.Session
        UserId string
    }
)

func (this *Service) DBAction(databaseName string, collectionName string,
                              mongoCall mongo.MongoCall) (err error) {
    return mongo.Execute(this.UserId, this.MongoSession,
                         databaseName, collectionName, mongoCall)
}
```

在 `Service` 类型中，包含了 Mongo 会话和用户 ID。函数 `DBAction` 提供了运行 MongoDB 命令和查询的抽象层。

## 实现一个 Web 调用

基于 base 类型和样板函数，我们现在可以实现一个 Web 调用了：

### Buoy Controller

类型 `BuoyController` 实现了两个结合 `BaseController` 和 business 层 Web API 调用。我们主要关注名为 `Station` 的 Web 调用。

```go
type BuoyController struct {
    bc.BaseController
}

func (this *BuoyController) Station() {
    buoyBusiness.Station(&this.BaseController, this.GetString(":stationId"))
}
```

类型 `BuoyController` 直接由单独的 `BaseController` 组成。通过这种组合方式，`BuoyController` 自动获得了已经自定义实现的 `Prepare` 和 `Finish` 方法以及所有 `beego.Controller` 的字段。

函数 `Station` 通过下面路由分发。`stationId` 作为 URL 的最后一项。当这个路由被请求时，一个 `BuoyController` 对象会被实例化并执行函数 `Station`。

	beego.Router("/station/:stationId", &controllers.BuoyController{}, "get:Station")
	
函数 `Station` 会在底层调用 business 层的代码来处理请求。

### Buoy Business

Buoy Business 包实现了 `BuoyController` 的 business 层。让我看看在 `BuoyController` 中的函数 `Station` 所调用的 business 层的代码是怎样的：

```go
func Station(controller *bc.BaseController, stationId string) {
    defer bc.CatchPanic(controller, "Station")

    tracelog.STARTEDf(controller.UserId, "Station", "StationId[%s]", stationId)

    buoyStation, err := buoyService.FindStation(&controller.Service, stationId)
    if err != nil {
        ServeError(err)
        return
    }

    controller.Data["json"] = &buoyStation
    controller.ServeJson()

    tracelog.COMPLETED(controller.UserId, "Station")
}
```

你可以看到 business 层的函数 `Station` 处理了整个请求的逻辑。这个函数还使用了Buoy Service 来处理与 MongoDB 的交互。

### Buoy Service

Buoy Service 包实现了与 MongoDB 的交互。让我们来看下被 business 层的函数 `Station` 所调用的函数 `FindStation` 的代码：

```go
func FindStation(service *services.Service, stationId string) (buoyStation *buoyModels.BuoyStation, err error) {
    defer helper.CatchPanic(&err, service.UserId, "FindStation")

    tracelog.STARTED(service.UserId, "FindStation")

    queryMap := bson.M{"station_id": stationId}
    tracelog.TRACE(service.UserId, "FindStation",
        "Query : %s", mongo.ToString(queryMap))

    buoyStation = &buoyModels.BuoyStation{}
    err = service.DBAction(Config.Database, "buoy_stations",
        func(collection *mgo.Collection) error {
            return collection.Find(queryMap).One(buoyStation)
        })

    if err != nil {
        tracelog.COMPLETED_ERROR(err, service.UserId, "FindStation")
        return buoyStation, err
    }

    tracelog.COMPLETED(service.UserId, "FindStation")
    return buoyStation, err
}
```

函数 `FindStation` 用于准备通过 `DBAction` 函数进行与 MongoDB 的查询与执行操作。

## 端到端测试

我们现在有一个 URL 可以路由到一个控制器以及相应的 business 与 service 层的逻辑，它可以通过以下方式测试：

```go
func TestStation(t *testing.T) {
    r, _ := http.NewRequest("GET", "/station/42002", nil)
    w := httptest.NewRecorder()
    beego.BeeApp.Handlers.ServeHTTP(w, r)

    err := struct {
        Error string
    }{}
    json.Unmarshal(w.Body.Bytes(), &err)

    Convey("Subject: Test Station Endpoint\n", t, func() {
        Convey("Status Code Should Be 200", func() {
            So(w.Code, ShouldEqual, 200)
        })
        Convey("The Result Should Not Be Empty", func() {
            So(w.Body.Len(), ShouldBeGreaterThan, 0)
        })
        Convey("The Should Be No Error In The Result", func() {
            So(len(err.Error), ShouldEqual, 0)
        })
    })
}
```

测试创建了一个某个路由的虚拟调用。这种做法很赞，因为我们不需要真正地去启动一个 Web 应用来测试代码。使用 goconvey 我们能够创建非常优雅的输出且易于阅读。

以下是一个测试失败的样例：

```
Subject: Test Station Endpoint

  Status Code Should Be 200 ✘
  The Result Should Not Be Empty ✔
  The Should Be No Error In The Result ✘

Failures:

* /Users/bill/Spaces/Go/Projects/src/github.com/goinggo/beego-mgo/test/endpoints/buoyEndpoints_test.go 
Line 35:
Expected: '200'
Actual: '400'
(Should be equal)

* /Users/bill/Spaces/Go/Projects/src/github.com/goinggo/beego-mgo/test/endpoints/buoyEndpoints_test.go 
Line 37:
Expected: '0'
Actual: '9'
(Should be equal)


3 assertions thus far

--- FAIL: TestStation-8 (0.03 seconds)
```

以下是一个测试成功的样例：

```
Subject: Test Station Endpoint

  Status Code Should Be 200 ✔
  The Result Should Not Be Empty ✔
  The Should Be No Error In The Result ✔

3 assertions thus far

--- PASS: TestStation-8 (0.05 seconds)
```

## 结论

花点时间下载这个项目随便看看。我尽最大努力让你能够清晰地看到我所想要展示给你的重点。beego 框架让你能够非常轻松地实现抽象和样板代码，遵照 go 的风格集成 go 测试、运行及部署。