# packages
**任何**go 程序都由packages 构成 ，程序的入口在 package main 中，引入外部package的方式为import


`package main # 代表程序的入口`
`import # 用于引入package`
```go
package main
import (
	"fmt"
	"math"
)
func main() {
    fmt.Println("Hello")
}

```

~~math.pi~~ 包的导出必须是大写字母开头  math.Pi

---
# function 函数
```go

#一个标准的函数声明
func fun_name(param1 T, param2 ....V) S {}
# func 保留字，表示声明了一个函数
# T, V, S 为参数的类型, go的参数的类型在参数的后面
```
> 注意到当多个参数同属一个类型时， 可以省略写法  
> x int, y int --> x, y int


```go
# 在go中，函数可以返回多个参数
func swap(x, y string) (string, string) {
	return y, x
}
# go 可以给返回参数命名, 可以理解为声明变量，该变量用于默认的return 返回
func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return
}

```

# variables 变量

```go
# go 的显示变量声明为var
## 声明的变量同属一种类型
var c, python, java bool # 未初始化时，变量的类型不能省略
## 声明的变量不是同种类型时
var (
    a = 1 # 声明时初始化，可以省略数据类型
    b bool
)

# 简短变量声明 :=

func test() { k := true}
```
> 简短变量的声明必须在函数内部，var声明可用于函数内外

# Type 数据类型
基本的数据类型：
1. bool
2. string
3. int int8 int16 int32 int64 uint ....
4. byte
5. float32 float64

```go
# go的类型转换必须时显示的
var i int = 42
var f float64 = float64(i)
var u uint = uint(f) # ✔
var u uint := f # ❌
i := 42
f := float64(i)
u := uint(f) # ✔
u := f # ❌
```

#  Flow control 控制流


#### for 
```go
# go 的for 循环不需要"()" 包裹其组件
# 初始化声明 ; 条件表达式; 后置语句声明

func main() {
	sum := 0
	for i := 0; i < 10; i++ {
		sum += i
	}
	fmt.Println(sum)
}
```
> 另外 后置语句以及初始化声明都可以省略，但是条件表达式不可省略否则会成为死循环

#### if 
```go

if 初始化声明; 条件表达式 {}
# 同样， 初始化声明可以省略
if 条件表达式 {}
if 初始化声明; 条件表达式 {} else {}
```

#### switch
```go
# go 的switch 只会执行第一个选中的case, 之后不会顺序执行， break是隐含的
switch X {
	case A:
	# 语句块
	case B:
	# 语句块
	default:
	# 语句块	
	}
# X 可以省略， 这样就演变成了 if else 控制链
t := time.Now()
switch {
	case t.Hour() < 12:
		fmt.Println("Good morning!")
	case t.Hour() < 17:
		fmt.Println("Good afternoon.")
	default:
		fmt.Println("Good evening.")
}
```

#### defer 延迟执行
> 1.由defer声明的执行语句，会在其父函数return时执行
> 2.另外被defer声明的执行语句，会依次压栈，满足LIFO 后进先出

```go
func main() {
	fmt.Println("counting")

	for i := 0; i < 10; i++ {
		defer fmt.Println(i)
	}

	fmt.Println("done")
}
# test out
# counting done 9 8 7 6 5 4 3 2 1 0 
```
#### pointers 指针
`var p *int`

```go
#指针变量其存储的是目标对象的地址，因此指针存储大小是固定的（同一机器）
#指针变量的类型反应了目标对象的数据类型
import "fmt"

func main() {
	i, j := 42, 2701 // i,j 的地址是对象存储的实际地址，因此存储大小随对象类型而变化
	p := &i         // point to i
	fmt.Println(*p) // read i through the pointer
	*p = 21         // set i through the pointer
	fmt.Println(i)  // see the new value of i

	p = &j         // point to j
	*p = *p / 37   // divide j through the pointer
	fmt.Println(j) // see the new value of j
}
```
> 注意go没有指针算数运算规则

#### struct 结构体

```go
type Vertex struct {
	X int
	Y int
}
# 通过"."运算来访问字段
v := Vertex{1, 2}
v.X = 4

# 指向结构体的指针
p := &v
p.X = 1e9
```
> 注意到指针变量和结构体变量都能通过"."访问结构体的字段内容，本质是指针是间接寻址，而变量是直接寻址

#### arrays 数组

`var a [10]int`   # var x [size]T
> 由于数组在定义时必须指定size，因此实际使用时切片更多

#### slices 切片

>切片是一个动态大小，灵活的数组**视图**，切片本身并不存储数据，底层结构仍然是数组
```go
a[low : high] // 切片是一个左闭右开
primes := [6]int{2, 3, 5, 7, 11, 13}
var s []int = primes[1: 4]

```
#### make  构造函数
`a := make([]int, 5) // len(a)=5` 
`b := make([]int, 0, 5) // len(b)=0, cap(b)=5`
> 第一个参数为数据类型，第二个为长度，第三个为容量

#### append  插入函数
`func append(s []T, vs ...T) []T // 第一俄国参数为 要插入的切片s, 剩余参数为要插入的值`

#### Range 区间
> 通过range 在切片或者map上生产一个迭代器

```go
var pow = []int{1, 2, 4, 8, 16, 32, 64, 128}

func main() {
	for i, v := range pow {
		fmt.Printf("2**%d = %d\n", i, v)
	}
}
// i,v 分别代表一次迭代返回的index索引以及value值

# index 和value可以选择性省略，但是需要用"_" 缺省表示
// for i, _ := range pow
// for _, value := range pow
```

#### Maps 散列表

```go

var m map[T]V // 表示声明一个map变量，key 为T数据类型， value为V数据类型，此时m 为 nil， 没有key，也不可添加键值对
m = make(map[T]V) // 对m 进行初始化，实际是分配了内存空间
m[t] = v // 在m中新增一对键值对
```

#### function values
> Functions are values too. They can be passed around just like other values.


```go
package main

import (
	"fmt"
	"math"
)

func compute(fn func(float64, float64) float64) float64 {
	return fn(3, 4)
}

func main() {
	hypot := func(x, y float64) float64 {
		return math.Sqrt(x*x + y*y)
	}
	fmt.Println(hypot(5, 12))

	fmt.Println(compute(hypot))
	fmt.Println(compute(math.Pow))
}
```

#### Function closures 函数闭包
> A closure is a function value that references variables from outside its body, the function is "bound" to the variables.

```go
package main

import "fmt"

func adder() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}

func main() { 
	pos, neg := adder(), adder()
	for i := 0; i < 10; i++ {
		fmt.Println(
			pos(i),
			neg(-2*i),
		)
	}
}

```
#### methods 方法
> A method is a function with a special receiver argument

```go
 func (v T) funcName() S {

 }
 # 定义该方法后， 对于数据类型T定义的对象实体，则可以调用该方法, 但注意v.funcName() 在funcName函数内部v是原对象的一个深拷贝，修改操作不会镜像到原对象上
```
> 因为go 没有classes 关键字，无法定义类对象, 即 字段+方法的对象

```go
 # 想要对某个数据类型封装方法，该数据类型必须和方法同属一个package
 # 因此，如果想要对基本数据类型封装方法，可以重新声明一个数据类型
 type MyFloat float64 // MyFloat 和 float64 底层是同一个, 仅仅是数据类型不一样
```
> You can only declare a method with a receiver whose type is defined in the same package as the method

#### Pointer receivers 指针接收器
> 对于 func (v T) funcName()无法对原对象进行修改的弊端, 指针接收器则弥补了这一问题

```go
# 因为在做拷贝时, 拷贝的是指针的内容，也就是实际对象的存储地址，因此在通过指针访问时，函数内部的修改会镜像到原对象上 == 浅拷贝
```
#### interfaces 接口
> An interface type is defined as a set of method signatures

```go
# empty interface
var i interface{} //声明了一个空的接口对象

```
>An empty interface may hold values of any type. Every type implements at least zero methods


#### 深拷贝和浅拷贝
> 指针类型的对象在拷贝时，是浅拷贝，因为拷贝的内容是地址
> 值类型的对象在拷贝时，是深拷贝，因为拷贝的内容是具体的数据结构，具体的数据值
> 都是拷贝内容，关键的区别在于存储的内容是什么 


#### Type assertions
> 所谓类型断言，指在未声明或者判断的前提下，断定对象的类型
```go
var i interface{} = "hello"

s, ok := i.(string)
fmt.Println(s, ok)

f, ok := i.(float64)
fmt.Println(f, ok)
```


#### type parameters 类型参数
`func funcName [T comparable] (s []T, x T) int //  comparable 作为对类型的一个约束,这行代码声明了一个函数，并且只要T数据类型是可比较的类型，就可以作为参数的类型` 
```go

func Index[T comparable](s []T, x T) int {
	for i, v := range s {
		// v and x are type T, which has the comparable
		// constraint, so we can use == here.
		if v == x {
			return i
		}
	}
	return -1
}


func main() {
	// Index works on a slice of ints
	si := []int{10, 20, 15, -10}
	fmt.Println(Index(si, 15))

	// Index also works on a slice of strings
	ss := []string{"foo", "bar", "baz"}
	fmt.Println(Index(ss, "hello"))
}
```

>This declaration means that s is a slice of any type T that fulfills the built-in constraint comparable. x is also a value of the same type.

#### Generic types 通用类型
> 上述的T数据类型含有comparable约束，如果不想含有任何约束，则是Generic types
> 
`[T any] 通用类型  `   
```go
type List[T any] struct {
	next *List[T]
	val  T
}

func main() {
}
```
---
#### Goroutines 协程

`go f(x, y, z)`
>A goroutine is a lightweight thread managed by the Go runtime.
```go
package main

import (
	"fmt"
	"time"
)

func say(s string) {
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func main() {
	go say("world")
	say("hello")
}

```
#### channels 通道
>Channels are a typed conduit through which you can send and receive values with the channel operator, <-.
```go
var v T
ch := make(chan T)
ch <- v    // Send v to channel ch.
v := <-ch  // Receive from ch, and
           // assign value to v.
```
> 默认情况下（无缓存）, 发送方协程和接收方协程会阻塞，直到另一方准备就绪， 这使得协程可以完成同步(按序执行)而无需额外的锁或者状态变量


> 一个通道内有两个等待队列:发送等待队列和接收等待队列，当读操作被阻塞，这个读的 goroutine 就被放入接收等待队列，当写操作触发时，调度器会检查对面的接收等待队列。如果发现队列里有等待者，就直接把写入的数据拷贝给那个等待的 goroutine，然后把对方从等待队列中移出，放入调度器的可运行队列（唤醒它）。整个过程中，这个写的 goroutine 不阻塞，继续执行。
```go
func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)
}
```
#### Buffered Channels
```go
# 通过该方式初始化的通道，只有在缓存满的时候，发送协程才会阻塞；缓冲空的时候，接受协程才会被阻塞
func main() {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	fmt.Println(<-ch)
	fmt.Println(<-ch)
}

```


#### Select 
>A select blocks until one of its cases can run, then it executes that case. It chooses one at random if multiple are ready. 
> 被select声明的，其包裹的所有case都会被监听， select所在的协程会一直被阻塞，直到有一个case可以执行时

```go

func fibonacci(c, quit chan int) {
	x, y := 0, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}

func main() {
	c := make(chan int)
	quit := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(<-c)
		}
		quit <- 0
	}()
	fibonacci(c, quit)
}
```


#### sync.Mutex 互斥信号量

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// SafeCounter is safe to use concurrently.
type SafeCounter struct {
	mu sync.Mutex
	v  map[string]int
}

// Inc increments the counter for the given key.
func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.v[key]++
	c.mu.Unlock()
}

// Value returns the current value of the counter for the given key.
func (c *SafeCounter) Value(key string) int {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	return c.v[key]
}

func main() {
	c := SafeCounter{v: make(map[string]int)}
	for i := 0; i < 1000; i++ {
		go c.Inc("somekey")
	}

	time.Sleep(time.Second)
	fmt.Println(c.Value("somekey"))
}

```