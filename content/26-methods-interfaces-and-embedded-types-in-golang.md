---
Author: AriesDevil
Date: 2014-05-05
---

# Go 语言中的方法，接口和嵌入类型

> 本文为转载技术翻译，原翻译地址：http://se77en.cc/2014/05/05/methods-interfaces-and-embedded-types-in-golang/

> 原
文地址：http://www.goinggo.net/2014/05/methods-interfaces-and-embedded-types.html

## 概述

在 Go 语言中，如果一个结构体和一个嵌入字段同时实现了相同的接口会发生什么呢？我们猜一下，可能有两个问题：

* 编译器会因为我们同时有两个接口实现而报错吗？
* 如果编译器接受这样的定义，那么当接口调用时编译器要怎么确定该使用哪个实现？

在写了一些测试代码并认真深入的读了一下标准之后，我发现了一些有意思的东西，而且觉得很有必要分享出来，那么让我们先从 Go 语言中的方法开始说起。

[more]

## 方法

Go 语言中同时有函数和方法。一个方法就是一个包含了[接受者](http://golang.org/ref/spec#Method_declarations)的函数，接受者可以是[命名类型](http://golang.org/ref/spec#Types)或者[结构体](http://golang.org/ref/spec#Struct_types)类型的一个值或者是一个指针。所有给定类型的方法属于该类型的方法集。

下面定义一个结构体类型和该类型的一个方法：

```
type User struct {
  Name  string
  Email string
}

func (u User) Notify() error
```

首先我们定义了一个叫做 `User` 的结构体类型，然后定义了一个该类型的方法叫做 `Notify`，该方法的接受者是一个 `User` 类型的值。要调用 `Notify` 方法我们需要一个 `User` 类型的值或者指针：

```
// User 类型的值可以调用接受者是值的方法
damon := User{"AriesDevil", "ariesdevil@xxoo.com"}
damon.Notify()

// User 类型的指针同样可以调用接受者是值的方法
alimon := &User{"A-limon", "alimon@ooxx.com"}
alimon.Notify()
```

在这个例子中当我们使用指针时，Go [调整](http://golang.org/ref/spec#Calls)和解引用指针使得调用可以被执行。***注意***，当接受者不是一个指针时，该方法操作对应接受者的值的副本(意思就是即使你使用了指针调用函数，但是函数的接受者是值类型，所以函数内部操作还是对副本的操作，而不是指针操作，参见：[http://play.golang.org/p/DBhWU0p1Pv](http://play.golang.org/p/DBhWU0p1Pv))。

我们可以修改 `Notify` 方法，让它的接受者使用指针类型：

```
func (u *User) Notify() error
```

再来一次之前的调用(***注意***：当接受者是指针时，即使用值类型调用那么函数内部也是对指针的操作，参见：[http://play.golang.org/p/SYBb4xPfPh](http://play.golang.org/p/SYBb4xPfPh))：

```
// User 类型的值可以调用接受者是指针的方法
damon := User{"AriesDevil", "ariesdevil@xxoo.com"}
damon.Notify()

// User 类型的指针同样可以调用接受者是指针的方法
alimon := &User{"A-limon", "alimon@ooxx.com"}
alimon.Notify()
```

如果你不清楚到底什么时候该使用值，什么时候该使用指针作为接受者，你可以去看一下[这篇介绍](http://se77en.cc/2014/05/04/choose-whether-to-use-a-value-or-pointer-receiver-on-methods/)。这篇文章同时还包含了社区约定的接受者该如何命名。

## 接口

Go 语言中的[接口](http://golang.org/doc/effective_go.html#interfaces)很特别，而且提供了难以置信的一系列灵活性和抽象性。它们指定一个特定类型的值和指针表现为特定的方式。从语言角度看，接口是一种类型，它指定一个[方法集](http://golang.org/ref/spec#Method_sets)，所有方法为[接口类型](http://golang.org/ref/spec#Interface_types)就被认为是该接口。

下面定义一个接口：

```
type Notifier interface {
  Notify() error
}
```

我们定义了一个叫做 `Notifier` 的接口并包含一个 `Notify` 方法。当一个接口只包含一个方法时，按照 Go 语言的[约定](http://golang.org/doc/effective_go.html#interface-names)命名该接口时添加 `-er` 后缀。这个约定很有用，特别是接口和方法具有相同名字和意义的时候。

我们可以在接口中定义尽可能多的方法，不过在 Go 语言标准库中，你很难找到一个接口包含两个以上的方法。

## 实现接口

当涉及到我们该怎么让我们的类型实现接口时，Go 语言是特别的一个。Go 语言不需要我们显式的实现类型的接口。如果一个接口里的所有方法都被我们的类型实现了，那么我们就说该类型实现了该接口。

让我们继续之前的例子，定义一个函数来接受任意一个实现了接口 `Notifier` 的类型的值或者指针：

```
func SendNotification(notify Notifier) error {
  return notify.Notify()
}
```

`SendNotification` 函数调用 `Notify` 方法，这个方法被传入函数的一个值或者指针实现。这样一来一个函数就可以被用来执行任意一个实现了该接口的值或者指针的指定的行为。

用我们的 `User` 类型来实现该接口并且传入一个 `User` 类型的值来调用 `SendNotification` 方法：

```
func (u *User) Notify() error {
  log.Printf("User: Sending User Email To %s<%s>\n",
      u.Name,
      u.Email)
  return nil
}

func main() {
  user := User{
    Name:  "AriesDevil",
    Email: "ariesdevil@xxoo.com",
  }

  SendNotification(user)
}

// Output:
cannot use user (type User) as type Notifier in function argument:
User does not implement Notifier (Notify method has pointer receiver)
```

详细代码：[http://play.golang.org/p/KG8-Qb7gqM](http://play.golang.org/p/KG8-Qb7gqM)

为什么编译器不考虑我们的值是实现该接口的类型？接口的调用规则是建立在这些方法的接受者和接口如何被调用的基础上。下面的是语言规范里定义的规则，这些规则用来说明是否我们一个类型的值或者指针[实现了](http://golang.org/ref/spec#Method_sets)该接口：

* 类型 `*T` 的可调用方法集包含接受者为 `*T` 或 `T` 的所有方法集

这条规则说的是如果我们用来调用特定接口方法的接口变量是一个指针类型，那么方法的接受者可以是值类型也可以是指针类型。显然我们的例子不符合该规则，因为我们传入 `SendNotification` 函数的接口变量是一个值类型。

* 类型 `T` 的可调用方法集包含接受者为 `T` 的所有方法

这条规则说的是如果我们用来调用特定接口方法的接口变量是一个值类型，那么方法的接受者必须也是值类型该方法才可以被调用。显然我们的例子也不符合这条规则，因为我们 `Notify` 方法的接受者是一个指针类型。

语言规范里只有这两条规则，我通过这两条规则得出了符合我们例子的规则：

* 类型 `T` 的可调用方法集不包含接受者为 `*T` 的方法

我们碰巧赶上了我推断出的这条规则，所以编译器会报错。`Notify` 方法使用指针类型作为接受者而我们却通过值类型来调用该方法。解决办法也很简单，我们只需要传入 `User` 值的地址到 `SendNotification` 函数就好了：

```
func main() {
  user := &User{
    Name:  "AriesDevil",
    Email: "ariesdevil@xxoo.com",
  }

  SendNotification(user)
}

// Output:
User: Sending User Email To AriesDevil<ariesdevil@xxoo.com>
```

详细代码：[http://play.golang.org/p/kEKzyTfLjA](http://play.golang.org/p/kEKzyTfLjA)

## 嵌入类型

[结构体类型](http://golang.org/ref/spec#Struct_types)可以包含匿名或者嵌入字段。也叫做嵌入一个类型。当我们嵌入一个类型到结构体中时，该类型的名字充当了嵌入字段的字段名。

下面定义一个新的类型然后把我们的 `User` 类型嵌入进去：

```
type Admin struct {
  User
  Level  string
}
```

我们定义了一个新类型 `Admin` 然后把 `User` 类型嵌入进去，注意这个不叫继承而叫组合。 `User` 类型跟 `Admin` 类型没有关系。

我们来改变一下 `main` 函数，创建一个 `Admin` 类型的变量并把变量的地址传入 `SendNotification` 函数中：

```
func main() {
  admin := &Admin{
    User: User{
      Name:  "AriesDevil",
      Email: "ariesdevil@xxoo.com",
    },
    Level: "master",
  }

  SendNotification(admin)
}

// Output
User: Sending User Email To AriesDevil<ariesdevil@xxoo.com>
```

详细代码：[http://play.golang.org/p/ivzzzk78TC](http://play.golang.org/p/ivzzzk78TC)

事实证明，我们可以 `Admin` 类型的一个指针来调用 `SendNotification` 函数。现在 `Admin` 类型也通过来自嵌入的 `User` 类型的***方法提升***实现了该接口。

如果 `Admin` 类型包含了 `User` 类型的字段和方法，那么它们在结构体中的关系是怎么样的呢？

> 当我们[嵌入](http://golang.org/doc/effective_go.html#embedding)一个类型，这个类型的方法就变成了外部类型的方法，但是当它被调用时，方法的接受者是内部类型(嵌入类型)，而非外部类型。-- Effective Go

因此嵌入类型的名字充当着字段名，同时嵌入类型作为内部类型存在，我们可以使用下面的调用方法：

```
admin.User.Notify()

// Output
User: Sending User Email To AriesDevil<ariesdevil@xxoo.com>
```

详细代码：[http://play.golang.org/p/0WL_5Q6mao](http://play.golang.org/p/0WL_5Q6mao)

这儿我们通过类型名称来访问内部类型的字段和方法。然而，这些字段和方法也同样被提升到了外部类型：

```
admin.Notify()

// Output
User: Sending User Email To AriesDevil<ariesdevil@xxoo.com>
```

详细代码：[http://play.golang.org/p/2snaaJojRo](http://play.golang.org/p/2snaaJojRo)

所以通过外部类型来调用 `Notify` 方法，本质上是内部类型的方法。

下面是 Go 语言中内部类型[方法集提升](http://golang.org/ref/spec#Method_sets)的规则：

给定一个结构体类型 `S` 和一个命名为 `T` 的类型，方法提升像下面规定的这样被包含在结构体方法集中：

* 如果 `S` 包含一个匿名字段 `T`，`S` 和 `*S` 的方法集都包含接受者为 `T` 的方法提升。

这条规则说的是当我们嵌入一个类型，嵌入类型的接受者为值类型的方法将被提升，可以被外部类型的值和指针调用。

*  对于 `*S` 类型的方法集包含接受者为 `*T` 的方法提升

这条规则说的是当我们嵌入一个类型，可以被外部类型的指针调用的方法集只有嵌入类型的接受者为指针类型的方法集，也就是说，当外部类型使用指针调用内部类型的方法时，只有接受者为指针类型的内部类型方法集将被提升。

* 如果 `S` 包含一个匿名字段 `*T`，`S` 和 `*S` 的方法集都包含接受者为 `T` 或者 `*T` 的方法提升

这条规则说的是当我们嵌入一个类型的指针，嵌入类型的接受者为值类型或指针类型的方法将被提升，可以被外部类型的值或者指针调用。

这就是语言规范里方法提升中仅有的三条规则，我根据这个推导出一条规则：

* 如果 `S` 包含一个匿名字段 `T`，`S` 的方法集不包含接受者为 `*T` 的方法提升。

这条规则说的是当我们嵌入一个类型，嵌入类型的接受者为指针的方法将不能被外部类型的值访问。这也是跟我们上面陈述的接口规则一致。

## 回答开头的问题

现在我们可以写程序来回答开头提出的两个问题了，首先我们让 `Admin` 类型实现 `Notifier` 接口：

```
func (a *Admin) Notify() error {
  log.Printf("Admin: Sending Admin Email To %s<%s>\n",
      a.Name,
      a.Email)

  return nil
}
```

`Admin` 类型实现的接口显示一条 admin 方面的信息。当我们使用 `Admin` 类型的指针去调用函数 `SendNotification` 时，这将帮助我们确定到底是哪个接口实现被调用了。

现在创建一个 `Admin` 类型的值并把它的地址传入 `SendNotification` 函数，来看看发生了什么：

```
func main() {
  admin := &Admin{
    User: User{
      Name:  "AriesDevil",
      Email: "ariesdevil@xxoo.com",
    },
    Level: "master",
  }

  SendNotification(admin)
}

// Output
Admin: Sending Admin Email To AriesDevil<ariesdevil@xxoo.com>
```

详细代码：[http://play.golang.org/p/JGhFaJnGpS](http://play.golang.org/p/JGhFaJnGpS)

预料之中，`Admin` 类型的接口实现被 `SendNotification` 函数调用。现在我们用外部类型来调用 `Notify` 方法会发生什么呢：

```
admin.Notify()

// Output
Admin: Sending Admin Email To AriesDevil<ariesdevil@xxoo.com>
```

详细代码：[http://play.golang.org/p/EGqK6DwBOi](http://play.golang.org/p/EGqK6DwBOi)

我们得到了 `Admin` 类型的接口实现的输出。`User` 类型的接口实现不被提升到外部类型了。

现在我们有了足够的依据来回答问题了：

* 编译器会因为我们同时有两个接口实现而报错吗？

不会，因为当我们使用嵌入类型时，类型名充当了字段名。嵌入类型作为结构体的内部类型包含了自己的字段和方法，且具有唯一的名字。所以我们可以有同一接口的内部实现和外部实现。

* 如果编译器接受这样的定义，那么当接口调用时编译器要怎么确定该使用哪个实现？

如果外部类型包含了符合要求的接口实现，它将会被使用。否则，通过方法提升，任何内部类型的接口实现可以直接被外部类型使用。

## 总结

在 Go 语言中，方法，接口和嵌入类型一起工作方式是独一无二的。这些特性可以帮助我们像面向对象那样组织结构然后达到同样的目的，并且没有其它复杂的东西。用本文中谈到的语言特色，我们可以以极少的代码来构建抽象和可伸缩性的框架。