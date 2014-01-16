---
Author: 无闻
Date: 2014-1-16
---

# Sourcegraph 简介

> 本文为转载，原文地址：http://wuwen.org/article/14/sourcegraph-intro.html

[Sourcegraph](https://sourcegraph.com/) 号称通过分析全球的开源项目来真正地展现相关项目之间的关联。它主要提供以下两个功能：

1. 根据代码查找文档与使用用例。
2. 选择正确的库或函数来使用。

那么，如何才能开始使用呢？首页一个孤零零的搜索框对于新手来说确实显得有些寂寥。不得不承认，当下的 Sourcegraph 还只能进行针对性比较强的搜索，还未达到很高层次的模糊搜索。不过，官方给出了三个新手可以考虑使用的方面：

1. 查找您已经熟悉的库或函数。
2. 通过 Sourcegraph 的功能来学习如何使用某个库或函数。
3. 在 Sourcegraph 上比较两个库或函数，来决定哪个更好用一些。

官方给出的建议还是比较抽象的。下面，我就以 [beego](http://beego.me) 为例，展示如何开始使用 Sourcegraph 这个网站来作为 Go 语言学习和开发的辅助工具。

[more]

## 搜索库

在首页搜索框，我们输入 `beego` 后，可看到一个即时提示的下拉框列表：

![](http://wuwen.org/static/upload/201401160705289.png)

- 第一行有一本书的图标的，表示这是一个项目，即 beego 这个项目。
- 第二行有一个用户图标的，则是以 beego 为名的用户项目，在这个例子中，是一个组织的名称。
- 其它行则是与 beego 关键字有关的函数或方法。

敲击回车，则可进入完整的搜索结果列表：https://sourcegraph.com/search?q=beego

## 项目首页

进入 [beego 的项目主页](https://sourcegraph.com/github.com/astaxie/beego) 可以看到一系列数据：

#### 顶部

![](http://wuwen.org/static/upload/201401160707323.png)

- beego 的贡献者数量
- beego 的使用者数量
- 使用 beego 的项目数量
- beego 的依赖包

#### 左侧

![](http://wuwen.org/static/upload/2014011607082910.png)

beego 在 GitHub 上的 README 文件。

#### 右侧

![](http://wuwen.org/static/upload/201401160709189.png)

beego 的函数、类型和方法的列表和统计数据，可进行项目内搜索。

## 具体信息页面

现在，我们进入 `beego.Controller.GetString()` 方法的具体信息页面：https://sourcegraph.com/github.com/astaxie/beego/symbols/go/github.com/astaxie/beego/Controller:type/GetString

- 首先出现的是其它项目使用这个方法的一些示例，具有非常完善的鼠标悬浮提示和单击跳转功能。
- 接着是使用者的一些统计信息。
- 最后是该方法的定义。

当您单击任意其它函数、类型或方法的时候，也会跳转到相应的具体信息页面。

当您的鼠标悬浮在某个示例的时候，左侧会出现 `More` 按钮，单击之后即可加载相应的整个源文件进行浏览；如果您想要某个单独页面来浏览一个源文件，则可以单击文件图标右侧的 XXX.go 链接，例如：[GetString](https://sourcegraph.com/github.com/lisijie/goblog/tree/ed9bd599a1c7664738eae5bc4b8d9bae08651db3/controllers/admin/user.go)。

## 总结

当初我写 [Go Walker](http://gowalker.org) 时的初衷就是在线源码快速浏览，加上强大的直接跳转功能。很高兴 sourcegraph 正在朝着这个方向努力并且做得很好，不可谓不是一个强大的在线源码浏览工具！