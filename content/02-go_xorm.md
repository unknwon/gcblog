---
Author: Lunny
Date: 2014-1-2
---

# xorm 的七种武器

[xorm](https://github.com/lunny/xorm) 是一个简单而强大的 Go 语言开源 ORM 库. 通过它可以使数据库操作非常简便。

了解过 Go 的人可能会有疑问，Go 已经提供了 database/sql 接口，操作各种数据库接口都一致了，还有必要再使用 ORM 吗？也有人觉得对于复杂 SQL 语句，ORM 是无法应付的。

是的，[xorm](https://github.com/lunny/xorm) 不是为了取代 SQL，它甚至可以和 SQL 混用，它是在 databse/sql 接口的基础之上提供了更多的特性。我们将这些功能和特性比喻成七种武器，来帮助开发者快速的完成数据库的操作。

[more]

## 安装

当然，第一步，我们必须要安装 [xorm](https://github.com/lunny/xorm)：

如果你有安装 gopm，强烈建议使用 [gopm](https://github.com/gpmgo/gopm) 来进行 go 的包管理：

	gopm get github.com/lunny/xorm

如果没有安装 gopm，当然也可以直接用 go 工具进行安装：

	go get github.com/lunny/xorm

## 数据库及驱动支持

[xorm](https://github.com/lunny/xorm) 当前支持如下 5 种数据库驱动和 4 种数据库。

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* MsSql: [github.com/lunny/godbc](https://github.com/lunny/godbc)


## 第一种武器：结构体和数据库表的映射

在 [xorm](https://github.com/lunny/xorm) 中，我们用一个结构体和数据库中的表进行对应，结构体中的 field 和数据库中的 column 进行对应，通过在 field 后面的 tag 来进行一些特殊的设置，例如：unique，表示唯一索引；pk 表示主键；version 表示乐观锁字段，等等。写过 sql 语句的同学可能会觉得这些标记很熟悉，是的，大部分和 sql 语句里面的 column 定义类似。

	type User struct {
		Id int64
		Name string `xorm:"unique not null"`
		Age int
		Avatar []byte
		Created time.Time `xorm:"created"`
		Updated time.Time `xorm:"updated"`
		Version int `xorm:"version"`
	}


对应的过程必然涉及到命名的映射。默认的 [xorm](https://github.com/lunny/xorm) 提供了 SnakeMapper，SameMapper, PrefixMapper,
SuffixMapper 几种命名方式和方案，基本可以满足各种需要。

## 第二种武器：连写神技

使用可以连写的 API 对于一个拥有语法提示的编辑器简直就是效率利器，那么我们还有什么理由不支持呢。看看我们这个能写多长：

	engine.Where("id=?", 1).And("name = ?", "xlw").Or("age > ?", 20).Asc("name").Desc("id").Find(&users)

[xorm](https://github.com/lunny/xorm) 主要的操作实际上是两个 RAW 函数和七个 ORM 函数：

### RAW 函数

* Query：执行 SQL 查询语句

		results, err := engine.Query("select * from user")

* Exec：执行 SQL 执行语句

		affected, err := engine.Exec("update user set .... where ...")

### ORM 函数

* Insert：插入一条或多条数据

		affected, err := engine.Insert(&struct)
		// INSERT INTO struct () values ()
		affected, err := engine.Insert(&struct1, &struct2)
		// INSERT INTO struct1 () values ()
		// INSERT INTO struct2 () values ()
		affected, err := engine.Insert(&sliceOfStruct)
		// INSERT INTO struct () values (),(),()
		affected, err := engine.Insert(&struct1, &sliceOfStruct2)
		// INSERT INTO struct1 () values ()
		// INSERT INTO struct2 () values (),(),()

* Get：获取单条数据

		has, err := engine.Get(&user)
		// SELECT * FROM user LIMIT 1

* Find：获取多条数据

		err := engine.Find(&users)
		// SELECT * FROM user

* Iterate & Rows：获取多条数据并逐条处理

		err := engine.Iterate(...， func(...) {
			// ...
		})
		// SELECT * FROM user

		rows, err := engine.Rows(...)
		// SELECT * FROM user
		for rows.Next() {
			rows.Scan(&user)
		}

* Update：更新一条或多条记录

		affected, err := engine.Update(&user)
		// UPDATE user SET

* Delete：删除数据

		affected, err := engine.Delete(&user)
		// DELETE FROM user Where ...

* Count：根据查询条件计算数量

		counts, err := engine.Count(&user)
		// SELECT count(*) AS total FROM user

这些函数放到连写的最后，前面可以采用各种条件的连写。

## 第三种武器：表结构同步

随着需求的改变，有的时候，我们不得不去修改原有的数据结构，那么此时，我们可能就要去修改数据库的表结构或者索引之类。这是一个繁琐的工作，而且很可能会漏掉之类。通过 [xorm](https://github.com/lunny/xorm) 的 `Sync` 函数，这个工作将变得简单得多。

	err := engine.Sync(new(User))

只需要在程序启动时，执行 `Sync`，并将需要同步的一个或者多个表对应的 Struct 作为参数传入，那么 engine 将自动的检测并新增表，新增字段，新增索引。

是不是很简单，很强大。当然，其实也可以做到自动删除列，但是这样的话，很有可能会引起数据丢失，因此当前只提供了自动新增的功能。这个功能后面还会继续增强。

## 第四种武器：混合事务

当使用事务处理时，需要创建 Session 对象。在进行事物处理时，可以混用 ORM 方法和 RAW 方法，如下代码所示：

	session := engine.NewSession()
	defer session.Close()

	// add Begin() before any action
	err := session.Begin()
	user1 := Userinfo{Username: "xiaoxiao", Departname: "dev", Alias: "lunny"}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		return
	}

	_, err = session.Where("id = ?", 2).Update(&Userinfo{Username: "yyy"})
	if err != nil {
		session.Rollback()
		return
	}
	
	_, err = session.Exec("delete from userinfo where username = ?", user2.Username)
	if err != nil {
		session.Rollback()
		return
	}
	
	// add Commit() after all actions
	err = session.Commit()
	if err != nil {
		return
	}

## 第五种武器：数据库缓存

[xorm](https://github.com/lunny/xorm) 内置了一致性缓存支持，根据测算，开启缓存后，查询性能提高了 3-5 倍。不过缓存默认并没有开启。要开启缓存，需要在engine创建完后进行配置，如：

启用一个全局的内存缓存

	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	engine.SetDefaultCacher(cacher)

上述代码采用了 LRU 算法的一个缓存，缓存方式是存放到内存中，缓存 struct 的记录数为 1000 条，缓存针对的范围是所有具有主键的表，没有主键的表中的数据将不会被缓存。
如果只想针对部分表，则：

	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	engine.MapCacher(&user, cacher)

如果要禁用某个表的缓存，则：

	engine.MapCacher(&user, nil)

设置完之后，其它代码基本上就不需要改动了，缓存系统已经在后台运行。

当前实现了内存存储的 CacheStore 接口 MemoryStore，如果需要采用其它设备存储，可以实现 CacheStore 接口。

不过需要特别注意不适用缓存或者需要手动编码的地方：

* 当使用了 `Distinct`,`Having`,`GroupBy` 方法将不会使用缓存

* 在 `Get` 或者 `Find` 时使用了 `Cols`,`Omit` 方法，则在开启缓存后此方法无效，系统仍旧会取出这个表中的所有字段。

* 在使用 Exec 方法执行了方法之后，可能会导致缓存与数据库不一致的地方。因此如果启用缓存，尽量避免使用 Exec。如果必须使用，则需要在使用了 Exec 之后调用 ClearCache 手动做缓存清除的工作。比如：

		engine.Exec("update user set name = ? where id = ?", "xlw", 1)
		engine.ClearCache(new(User))

## 第六种武器：乐观锁

很多从事金融软件的朋友都会关注这个问题，乐观锁普遍的在 Java，.Net的ORM框架中被实现。使用如下：

要使用乐观锁，需要使用 version 标记

	type User struct {
		Id int64
		Name string
		Version int `xorm:"version"`
	}

在 Insert 时，version 标记的字段将会被设置为 1，在 Update 时，Update 的内容必须包含 version 原来的值。

	var user User
	engine.Id(1).Get(&user)
	// SELECT * FROM user WHERE id = ?
	engine.Id(1).Update(&user)
	// UPDATE user SET ..., version = version + 1 WHERE id = ? AND version = ?

## 第七种武器：数据库反转

现在是不是有点冲动想开始用[ xorm](https://github.com/lunny/xorm) 来管理你的数据库了呢？可以已有数据库怎么处理呢？要重写很多代码吗？

还在羡慕 Java 的 Hibernate 可以通过数据库自动生成 Bean 和 DAO 代码吗？现在 Go 也有了，而且支持生成 c++ 代码，欢呼吧。

[xorm](https://github.com/lunny/xorm) 自带了一个命令行工具，当前提供了反转命令。通过执行

	go install github.com/lunny/xorm/xorm

即可安装该工具。安装完成后，我们就可以来使用 `xorm` 这个命令了。[xorm](https://github.com/lunny/xorm) 的反转命令当前支持 sqlite3,mysql,postgres 以及 mssql 四种数据库。命令执行如下：

	xorm reverse sqite3 test.db templates/goxorm

其中，第二和第三个参数为数据库驱动的连接参数，最后一个参数为模板路径。对的，你可以在这个模板的基础上进行修改，最后也许你的大部分 models 代码都可以自动生成。我的一个同事已经通过这个工具，生成了一个 Mysql 数据库的 C++ 操作代码。

## 最后

[xorm](https://github.com/lunny/xorm) 项目已经发展了半年多，目前我们已经有两位长期的贡献者，很多同学也给出了各种方面的建议和修正。目前 [xorm](https://github.com/lunny/xorm) 还在不断的成长，欢迎大家多提意见建议贡献代码。