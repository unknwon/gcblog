---
Author: skoo
Date: 2014-1-11
---

# goroutine 与调度器

> 本文为转载，原文地址：http://www.bigendian123.com/go/2013/11/29/golang-schedule/

我们都知道 Go 语言是原生支持语言级并发的，这个并发的最小逻辑单元就是 goroutine。goroutine 就是 Go 语言提供的一种用户态线程，当然这种用户态线程是跑在内核级线程之上的。当我们创建了很多的 goroutine，并且它们都是跑在同一个内核线程之上的时候，就需要一个调度器来维护这些 goroutine，确保所有的 goroutine 都使用 cpu，并且是尽可能公平的使用 cpu 资源。

这个调度器的原理以及实现值得我们去深入研究一下。支撑整个调度器的主要有 4 个重要结构，分别是 M、G、P、Sched，前三个定义在 runtime.h 中，Sched 定义在 proc.c 中。

Sched 结构就是调度器，它维护有存储M和G的队列以及调度器的一些状态信息等。
M 代表内核级线程，一个M就是一个线程，goroutine 就是跑在 M 之上的；M是一个很大的结构，里面维护小对象内存 cache（mcache）、当前执行的 goroutine、随机数发生器等等非常多的信息。
P 全称是 Processor，处理器，它的主要用途就是用来执行 goroutine 的，所以它也维护了一个 goroutine 队列，里面存储了所有需要它来执行的 goroutine，这个P的角色可能有一点让人迷惑，一开始容易和 M 冲突，后面重点聊一下它们的关系。
G 就是 goroutine 实现的核心结构了，G 维护了 goroutine 需要的栈、程序计数器以及它所在的 M 等信息。
理解 M、P、G 三者的关系对理解整个调度器非常重要，我从网络上找了一个图来说明其三者关系：

![](11-golang-schedule-pics/golang-car.jpg)

地鼠(gopher)用小车运着一堆待加工的砖。M 就可以看作图中的地鼠，P 就是小车，G 就是小车里装的砖。一图胜千言啊，弄清楚了它们三者的关系，下面我们就开始重点聊地鼠是如何在搬运砖块的。

### 启动过程

在关心绝大多数程序的内部原理的时候，我们都试图去弄明白其启动初始化过程，弄明白这个过程对后续的深入分析至关重要。在 asm_amd64.s 文件中的汇编代码 _rt0_amd64 就是整个启动过程，核心过程如下：

```
CALL	runtime·args(SB)
CALL	runtime·osinit(SB)
CALL	runtime·hashinit(SB)
CALL	runtime·schedinit(SB)

// create a new goroutine to start program
PUSHQ	$runtime·main·f(SB)		// entry
PUSHQ	$0			// arg size
CALL	runtime·newproc(SB)
POPQ	AX
POPQ	AX

// start this M
CALL	runtime·mstart(SB)
```

启动过程做了调度器初始化 runtime·schedinit 后，调用 runtime·newproc 创建出第一个 goroutine，这个 goroutine 将执行的函数是 runtime·main，这第一个 goroutine 也就是所谓的主 goroutine。我们写的最简单的 Go 程序 ”hello，world” 就是完全跑在这个 goroutine 里，当然任何一个 Go 程序的入口都是从这个 goroutine 开始的。最后调用的 runtime·mstart 就是真正的执行上一步创建的主 goroutine。

启动过程中的调度器初始化 runtime·schedinit 函数主要根据用户设置的 GOMAXPROCS 值来创建一批小车(P)，不管 GOMAXPROCS 设置为多大，最多也只能创建 256 个小车(P)。这些小车(p)初始创建好后都是闲置状态，也就是还没开始使用，所以它们都放置在调度器结构(Sched)的 pidle 字段维护的链表中存储起来了，以备后续之需。

查看 runtime·main 函数可以了解到主 goroutine 开始执行后，做的第一件事情是创建了一个新的内核线程(地鼠 M)，不过这个线程是一个特殊线程，它在整个运行期专门负责做特定的事情——系统监控(sysmon)。接下来就是进入 Go 程序的 main 函数开始 Go 程序的执行。

至此，Go 程序就被启动起来开始运行了。一个真正干活的 Go 程序，一定创建有不少的 goroutine，所以在 Go 程序开始运行后，就会向调度器添加 goroutine，调度器就要负责维护好这些 goroutine 的正常执行。

### 创建 goroutine(G)

在Go程序中，时常会有类似代码：

	go do_something()
	
go 关键字就是用来创建一个 goroutine 的，后面的函数就是这个 goroutine 需要执行的代码逻辑。go 关键字对应到调度器的接口就是 runtime·newproc。`runtime·newproc` 干的事情很简单，就负责制造一块砖(G)，然后将这块砖(G)放入当前这个地鼠(M)的小车(P)中。

每个新的 goroutine 都需要有一个自己的栈，G结构的 `sched` 字段维护了栈地址以及程序计数器等信息，这是最基本的调度信息，也就是说这个 goroutine 放弃 cpu 的时候需要保存这些信息，待下次重新获得 cpu 的时候，需要将这些信息装载到对应的 cpu 寄存器中。

假设这个时候已经创建了大量的 goroutne，就轮到调度器去维护这些 goroutine 了。

### 创建内核线程(M)

![](11-golang-schedule-pics/golang-car.jpg)

Go 程序中没有语言级的关键字让你去创建一个内核线程，你只能创建 goroutine，内核线程只能由 runtime 根据实际情况去创建。runtime 什么时候创建线程？以地鼠运砖图来讲，砖(G)太多了，地鼠(M)又太少了，实在忙不过来，刚好还有空闲的小车(P)没有使用，那就从别处再借些地鼠(M)过来直到把小车(p)用完为止。这里有一个地鼠(M)不够用，从别处借地鼠(M)的过程，这个过程就是创建一个内核线程(M)。创建M的接口函数是:

	void newm(void (*fn)(void), P *p)

newm 函数的核心行为就是调用 clone 系统调用创建一个内核线程，每个内核线程的开始执行位置都是 runtime·mstart 函数。参数 p 就是一辆空闲的小车(p)。

每个创建好的内核线程都从 runtime·mstart 函数开始执行了，它们将用分配给自己小车去搬砖了。

### 调度核心

newm 接口只是给新创建的M分配了一个空闲的 P，也就是相当于告诉借来的地鼠(M)——“接下来的日子，你将使用 1 号小车搬砖，记住是 1 号小车；待会自己到停车场拿车。”，地鼠(M)去拿小车(P)这个过程就是 `acquirep`。 runtime·mstart 在进入 `schedule` 之前会给当前 M 装配上 P，runtime·mstart 函数中的代码：

```
} else if(m != &runtime·m0) {
	acquirep(m->nextp);
	m->nextp = nil;
}
schedule();
```

if 分支的内容就是为当前 M 装配上 P，`nextp` 就是 newm 分配的空闲小车(P)，只是到这个时候才真正拿到手罢了。没有 P，M 是无法执行 goroutine 的，就像地鼠没有小车无法运砖一样的道理。对应 acquirep 的动作是 releasep，把 M 装配的 P 给载掉；活干完了，地鼠需要休息了，就把小车还到停车场，然后睡觉去。

地鼠(M)拿到属于自己的小车(P)后，就进入工场开始干活了，也就是上面的 `schedule` 调用。简化 schedule 的代码如下：

```
static void
schedule(void)
{
	G *gp;

	gp = runqget(m->p);
	if(gp == nil)
		gp = findrunnable();

	if (m->p->runqhead != m->p->runqtail &&
		runtime·atomicload(&runtime·sched.nmspinning) == 0 &&
		runtime·atomicload(&runtime·sched.npidle) > 0)  // TODO: fast atomic
		wakep();

	execute(gp);
}
```

schedule 函数被我简化了太多，主要是我不喜欢贴大段大段的代码，因此只保留主干代码了。这里涉及到 4 大步逻辑：

![](11-golang-schedule-pics/gopher-bz.jpg)

1. `runqget`, 地鼠(M)试图从自己的小车(P)取出一块砖(G)，当然结果可能失败，也就是这个地鼠的小车已经空了，没有砖了。
2. `findrunnable`, 如果地鼠自己的小车中没有砖，那也不能闲着不干活是吧，所以地鼠就会试图跑去工场仓库取一块砖来处理；工场仓库也可能没砖啊，出现这种情况的时候，这个地鼠也没有偷懒停下干活，而是悄悄跑出去，随机盯上一个小伙伴(地鼠)，然后从它的车里试图偷一半砖到自己车里。如果多次尝试偷砖都失败了，那说明实在没有砖可搬了，这个时候地鼠就会把小车还回停车场，然后 `睡觉` 休息了。如果地鼠睡觉了，下面的过程当然都停止了，地鼠睡觉也就是线程 sleep 了。
3. `wakep`, 到这个过程的时候，可怜的地鼠发现自己小车里有好多砖啊，自己根本处理不过来；再回头一看停车场居然有闲置的小车，立马跑到宿舍一看，你妹，居然还有小伙伴在睡觉，直接给屁股一脚，“你妹，居然还在睡觉，老子都快累死了，赶紧起来干活，分担点工作。”，小伙伴醒了，拿上自己的小车，乖乖干活去了。有时候，可怜的地鼠跑到宿舍却发现没有在睡觉的小伙伴，于是会很失望，最后只好向工场老板说——”停车场还有闲置的车啊，我快干不动了，赶紧从别的工场借个地鼠来帮忙吧。”，最后工场老板就搞来一个新的地鼠干活了。
4. `execute`，地鼠拿着砖放入火种欢快的烧练起来。

*注： “地鼠偷砖”叫 [work stealing](http://supertech.csail.mit.edu/papers/steal.pdf)，一种调度算法。*

### 调度点

当我们翻看 channel 的实现代码可以发现，对 channel 读写操作的时候会触发调用 runtime·park 函数。 goroutine 调用 park 后，这个 goroutine 就会被设置位 waiting 状态，放弃 cpu。被 park 的 goroutine 处于 waiting 状态，并且这个 goroutine 不在小车(P)中，如果不对其调用 runtime·ready，它是永远不会再被执行的。除了 channel 操作外，定时器中，网络 poll 等都有可能 park goroutine。

除了 park 可以放弃 cpu 外，调用 runtime·gosched 函数也可以让当前 goroutine 放弃 cpu，但和 park 完全不同；gosched 是将 goroutine 设置为 runnable 状态，然后放入到调度器全局等待队列（也就是上面提到的工场仓库，这下就明白为何工场仓库会有砖块(G)了吧）。

除此之外，就轮到系统调用了，有些系统调用也会触发重新调度。Go 语言完全是自己封装的系统调用，所以在封装系统调用的时候，可以做不少手脚，也就是进入系统调用的时候执行 entersyscall，退出后又执行 exitsyscall 函数。 也只有封装了 entersyscall 的系统调用才有可能触发重新调度，它将改变小车(P)的状态为 syscall。还记一开始提到的 sysmon 线程吗？这个系统监控线程会扫描所有的小车(P)，发现一个小车(P)处于了 syscall 的状态，就知道这个小车(P)遇到了 goroutine 在做系统调用，于是系统监控线程就会创建一个新的地鼠(M)去把这个处于 syscall 的小车给抢过来，开始干活，这样这个小车中的所有砖块(G)就可以绕过之前系统调用的等待了。被抢走小车的地鼠等系统调用返回后，发现自己的车没，不能继续干活了，于是只能把执行系统调用的 goroutine 放回到工场仓库，自己 `睡觉` 去了。

从 goroutine 的调度点可以看出，调度器还是挺粗暴的，调度粒度有点过大，公平性也没有想想的那么好。总之，这个调度器还是比较简单的。

### 现场处理

goroutine 在 cpu 上换入换出，不断上下文切换的时候，必须要保证的事情就是 `保存现场` 和 `恢复现场`，保存现场就是在 goroutine 放弃 cpu 的时候，将相关寄存器的值给保存到内存中；恢复现场就是在 goroutine 重新获得 cpu 的时候，需要从内存把之前的寄存器信息全部放回到相应寄存器中去。

goroutine 在主动放弃 cpu 的时候(park/gosched)，都会涉及到调用 runtime·mcall 函数，此函数也是汇编实现，主要将 goroutine 的栈地址和程序计数器保存到 G 结构的 `sched` 字段中，mcall 就完成了现场保存。恢复现场的函数是 runtime·gogocall，这个函数主要在 `execute` 中调用，就是在执行 goroutine 前，需要重新装载相应的寄存器。