---
Author: Justin Huang
Date: 2014-1-20
---

# Go Slice 机制解析

> 本文为转载，原文地址：http://sharecore.net/blog/2013/09/29/slice-mechanics-by-rob/

Rob Pike写了篇关于Go的数组与切片的文章：[Arrays, slices (and strings): The mechanics of ‘append’](http://blog.golang.org/slices) ，介绍了slice的实现和一些常见的操作。其部分内容与我[这篇文章](http://sharecore.info/blog/2013/07/23/the-trap-of-go-slice-appending/)是重复的,所以就不一一翻译了，而是挑选部分内容记下，算是对我[这篇文章](http://sharecore.info/blog/2013/07/23/the-trap-of-go-slice-appending/)的一个内容补充。

对于以下内容的理解，首先需要理解[这篇文章](http://sharecore.info/blog/2013/07/23/the-trap-of-go-slice-appending/)来提到的关于slice的结构定义，即可以用一个包含长度和一个指向数组的指针(当然还有容量)的struct来描述。

[more]

```
type sliceHeader struct{
	Ptr *int //指向分配的数组的指针
	Len int  // 长度
	Cap int  // 容量
}
```

通过以下方式来定义一个slice:

```
	arr := [5]int{1, 2, 3, 4, 5}
	slice:=arr[2:4]
```

以上slice其实等同于以下定义：

```
slice:=sliceHeader{
	Len:2,
	Cap:3,
	Ptr:&arr[2]
}
```

了解了以上的定义，我们再看几种情况：

- 将slice当做参数传递时，其实相当于值传递了一个sliceHeader，这个sliceHeader包含了一个指向原数组的指针:

```
func AddOneToEachElement(slice []int) {
  for i := range slice {
      slice[i]++
  }
}

func SubtractOneFromLength(slice []byte) []byte {
    return slice[0 : len(slice)-1]
}

func main() {
  arr := [5]int{1, 2, 3, 4, 5}
  slice := arr[2:4]
  //执行对元素+1的操作
    fmt.Println("Before Array:", arr)
  AddOneToEachElement(slice)
  fmt.Println("After Array:", arr)

    //执行长度切割操作
    fmt.Println("Before: len(slice) =", len(slice))
  newSlice := SubtractOneFromLength(slice)
  fmt.Println("After:  len(slice) =", len(slice))
  fmt.Println("After:  len(newSlice) =", len(newSlice))
}
```

输出结果：

```
Before Array: [1 2 3 4 5]
After Array: [1 2 4 5 5] //数组的值被相应的改变

Before: len(slice) = 2
After:  len(slice) = 2//原来slice的长度并没有改变，说明slice参数是用的值传递
After:  len(newSlice) = 1
```

- 如果需要将slice当做引用传递，需要使用slice指针。（其实，用指针来表示引用传递，几乎就是Go中指针的唯一作用了，Go并不支持指针运算等特性）

```
func PtrSubtractOneFromLength(slicePtr *[]int) {
	slice := *slicePtr
	*slicePtr = slice[0 : len(slice)-1]
}

func main() {
	arr := [5]int{1, 2, 3, 4, 5}
	slice := arr[2:4]
	fmt.Println("Before: len(slice) =", len(slice))
	PtrSubtractOneFromLength(&slice)
	fmt.Println("After:  len(slice) =", len(slice))
}
```

输出结果：

```
Before: len(slice) = 2
After:  len(slice) = 1//原来slice的长度发生变化，所以使用指针时，是引用传递
```

- 关于Cap字段的值，是由指针所指的数组的长度和来指定的。坎以下代码，我们对一个slice进行扩展：

```
func Extend(slice []int, element int) []int {
	n := len(slice)
	slice = slice[0 : n+1]
	slice[n] = element
	return slice
}

func main() {
	arr := [5]int{}
	slice := arr[0:0]
	for i := 0; i < 20; i++ {
	    slice = Extend(slice, i)
	    fmt.Println(slice)
	}
}
```

输出结果：

```
[0]
[0 1]
[0 1 2]
[0 1 2 3]
[0 1 2 3 4]
  
panic: runtime error: slice bounds out of range
  
goroutine 1 [running]:
main.main()
  /home/justinhuang/src/go/src/test/slice1.go:17 +0x82
```

通过上面输出看到，当slice的长度达到数组长度(5)时，将出现越界的错误。

当然，你也可以使用内置方法cap()来获取slice的容量。

```
if cap(slice) == len(slice) {
    fmt.Println("slice is full!")
}
```

- 字符串可以当做一个字节slice([]byte)来操作，

```
slash := "/usr/ken"[0] //将得到字节值：'/'
usr := "/usr/ken"[0:4] // 将得到字符串："/usr"
```

以下可以将字符串反转为一个slice（[]byte）

	slice := []byte(usr)
	
字符串的操作，在我实现的razor视图引擎里大量用到：[razorparser.go](https://github.com/JustinHuang917/gof/blob/master/goftool/parser/razor/razorparser.go)

更多关于slice的make,copy，append操作，详见我[这篇文章](http://sharecore.info/blog/2013/07/23/the-trap-of-go-slice-appending/)