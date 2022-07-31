## 第1章 入门

### 1.拼接多个字符串

P6

```go
 strings.Join(os.Arg[1:],” ”)
```

### 2.bufio

P6

> 可以简便高效地处理输入和输出。其中一个最有用的特性是称为扫描器（Scanner）的类型，可以读取输入，以行或者单词为单位断开，这是处理以**行**为单位的输入内容的最简单方式

```go
	counts := make(map[string]int)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		counts[input.Text()]++
	}
```

> 扫描器从程序的标准输入中进行读取。每次调用input.Scan()读取下一行，并将结尾的换行符去掉。通过input.Text()获取读到的内容。Scan在没有新行时返回false

### 3.os

> 1.os.Open()
>
> 返回两个值，第一个是打开的文件(*os.File),第二个是err，记得写完后加Close()
>
> ```go
> f,err := os.Open(arg)
> f.close()
> ```
>
> 2.命令行参数：os.Args
>
> os.Args是一个字符串slice（动态容量的顺序数组），os.Args[0]是命令本身的名字，因此通常写成os.Args[1:]

### 4.io/ioutil

```go
	counts := make(map[string]int)
	for _, filename := range os.Args[1:] {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dup3: %v\n", err)
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			counts[line]++
		}
	}
```

> 1.ioutil.ReadFile()
>
> ReadFile函数返回一个可以转化为字符串的字节slice和一个err

### 5.底层的一个总结

> bufio.Scanner,ioutil.ReadFile,ioutil.WriteFile使用*os.File中的Read和Write方法

### 6.Web服务器

> 开启一个端口，每次访问输出一个gif动画

```go
package main

import (
	"image"
	"image/color"
	"image/gif"
	"io"
	"math"
	"math/rand"
)

import (
	"log"
	"net/http"
)

var palette = []color.Color{color.White, color.Black}

const (
	whiteIndex = 0 // first color in palette
	blackIndex = 1 // next color in palette
)

func main() {
	handler:= func(w http.ResponseWriter,r *http.Request) {
		lissajous(w)
	}
	http.HandleFunc("/", handler) // each request calls handler
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func lissajous(out io.Writer) {
	const (
		cycles  = 5     // number of complete x oscillator revolutions
		res     = 0.001 // angular resolution
		size    = 1000   // image canvas covers [-size..+size]
		nframes = 64    // number of animation frames
		delay   = 8     // delay between frames in 10ms units
	)
	freq := rand.Float64() * 3.0 // relative frequency of y oscillator
	anim := gif.GIF{LoopCount: nframes}
	phase := 0.0 // phase difference
	for i := 0; i < nframes; i++ {
		rect := image.Rect(0, 0, 2*size+1, 2*size+1)
		img := image.NewPaletted(rect, palette)
		for t := 0.0; t < cycles*2*math.Pi; t += res {
			x := math.Sin(t)
			y := math.Sin(t*freq + phase)
			img.SetColorIndex(size+int(x*size+0.5), size+int(y*size+0.5),
				blackIndex)
		}
		phase += 0.1
		anim.Delay = append(anim.Delay, delay)
		anim.Image = append(anim.Image, img)
	}
	gif.EncodeAll(out, &anim) // NOTE: ignoring encoding errors
}
```

## 第2章 程序结构

### 1.flag

> flag.Bool()
>
> 创建一个新的布尔类型标识变量。有三个参数：标识的名字，变量的默认值，以及当用户通过非法标识，非法参数，或者-h,-help时输出的信息
>
> 访问必须通过：*变量名
>
> 使用标识前，必须调用flag.Parse()来更新标识变量默认值

### 2.new

语法糖，返回的是地址

### 3.三种返回带有bool类型变量的操作

```go
v,ok = m[key]	//map查询
v,ok = x.(T)	//类型断言
v,ok = <-ch		//通道接受
```

### 4.String

> 类似java的toString()
>
> 类型声明String方法后，通过fmt包作为字符串输出时会根据String方法输出

### 5.容易BUG的

```go
var cwd string
func init(){
	cwd,err := os.Getwd()	//编译错误：未使用cwd
	if err != nil{
		log.Fatalf("os.Getwd failed:%v",err)
	}
}
修改：
var cwd string
func init(){
	var err error
	cwd,err = os.Getwd()
	if err != nil{
		log.Fatalf("os.Getwd failed:%v",err)
	}
}
```

## 第3章 基本数据

### 1.rune == int32  byte==uint8

> rune常用于指明一个值是Unicode码点
>
> byte强调一个值是原始数据
>
> uintptr无符号整数，足以完整存放指针，仅用于底层编程

### 2.&^运算

> z=x&^y，若y的某位是1 ，则z对应的位等于0，否则等于x的对应位。
>
> 例如：
>
> ​		 00100010	x
>
> &^	00000110	y
>
> ​		 00100000	z

### 3.Pintf()小技巧

> %后的[1]表示重复使用第一个操作数，#表示输出前缀

```
o := 0666
fmt.Printf("%d %[1]o %#[1]o\n",o)	//438 666 0666
```

### 4.字符串

#### utf8

P51

> Unicode和UTF-8
>
> 按照UTF-8：
>
> utf8.RuneCountInString()：计算字节数（以utf-8的格式，而不是一个个字节来算）
>
> utf8.DecodeRuneInString()：返回r（文字符号本身）和一个值（按utf8编码所占字节数，这个值可以用来定位下一个文字符号）
>
> UTF-8解码器出错时会产生一个专门的字符‘\uFFFD’

#### 4个标准包：bytes,strings,strconv,unicode

> bytes.Buffer类型好像挺重要的

#### path/filepath(不太重要的包)

P54

> 用来操作文件路径等**具有层次结构**的名字
>
> path处理以‘/’分段的路径字符串，不分平台，如URL地址的路径部分
>
> path/filepath根据平台的规则处理文件名

#### 字符串和数字的相互转换

P56

> 整数 -->字符串 ：fmt.Sprintf或者strconv.Itoa()
>
> strconv.FormatInt()可以按不同进位制格式化数字
>
> 字符串 --> 整数 ：strconv.Atoi或者strconv.ParseInt()

### 常量

> 常类生成器iota:从0开始取值，逐项加1

```go
type Weekday int
const(
	Sunday Weekday = iota	//0
	Monday					//1
	Tuesday					//2
)

type Flags uint
const(
	FlagUp Flags = 1<<iota	// xxxxx001
	FlagBroadcast			// xxxxx010
	FlagLoopback			// xxxxx100
)
```

> 无类型常类：编译器从属类型待定的常类表示成某些值，这些值比基本类型精度更高，可以认为精度达到256位
>
> 只有常量才可以是无类型的

## 第4章 复合数据类型

### 4.1 数组

1.数组元素类型是可比较的，那么数组就是可比较的（使用==操作符）

2.go把数组和其他类型的类型都看作值传递

### 4.2 slice

1.一种轻量级的数据结构

2.三个属性：指针，长度和容量

3.一个案例：假如我想直接从第1个元素开始存数据，可以用以下这种方法

```go
//可以不设置索引为0的元素
months := [...]string{1:"January",/*...*/,12:"December"}
```

4.将一个slice左移n个元素的简单方法是连续调用三次reverse()，第一次反转前n个元素，第二次反转剩下的，最后整个再反转

5.slice无法比较

6.检查一个slice是否为空， 使用len(s) == 0，而不是s == nil

```go
var s []int	
s=nil
s=[]int(nil)
s=[]int{}
```



#### 4.2.1 append函数

1.一次性可以添加多个元素，甚至于添加一个slice

#### 4.2.2 slice就地修改

1.可以用来实现栈

### 4.3 map

1.键类型key必须是可以通过操作符”==“来进行比较的数据类型

2.无法获取元素地址

> 原因之一：map的增长可能会导致已有元素被重新散列到新的存储位置

3.区分元素不存在，或者这个元素为0时：

```go
if ags,ok := ages["bob"];!ok{/*...*/}
```

4.用map来实现集合类型（复杂的例子见p74）

### 4.4 结构体

#### 4.4.1 结构体字面量

#### 4.4.2 结构体比较

1.结构体的所有成员变量都可以比较，那么这个结构体就是可比较的。这种情况下，用“==”比较时，依次按照顺序比较两个结构体变量的成员变量

#### 4.4.3 结构体嵌套和匿名成员

### 4.5 JSON

1.omitempty：如果这个成员的值是零值或者为空，则不输出到JSON中

> 如果想把0值输出到JSON，那么可以用指针的方式定义结构体成员

2.unmarshal阶段，JSON字段的名称关联到Go结构体成员名称的时候是忽略大小写的

### 4.6 文本和HTML模板

1.text/template 和 html/template：可以将程序变量的值代入到文本或者HTML模板中

## 第5章 函数

### 5.1 函数声明

1.偶尔可以看到没有函数体的函数声明，说明这个函数使用了Go以外的语言实现

```go
func Sin(x float64) float64//使用汇编语言实现
```

### 5.2 递归

1.许多编程语言使用了固定长度的栈，大小在64KB~2MB之间。Go语言的实现使用了可变长度的栈，栈的大小随着使用而增长，最大可达1GB左右

### 5.3 多返回值

### 5.4 错误

#### 5.4.1错误处理策略（感觉比较关键）

5种情况

1.将错误传递下去

> 当错误最终被程序的main函数处理时，应当能够提供一个从最根本问题到总体故障的清晰因果链
>
> 设计一个错误消息的时候应当慎重，确保每一条消息的描述都是有意义的，包含充足的相关信息，并且保持一致性。

2.对于不固定或者不可预测的错误，可以在短暂的间隔后对操作进行重试，超出一定重试次数和限定时间后再报错退出

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

//!+
// WaitForServer 尝试连接URL对应的服务器.
// 在一分钟内使用指数退避策略进行重试.
// 所有尝试失败后返回错误.
func WaitForServer(url string) error {
	const timeout = 1 * time.Minute
	deadline := time.Now().Add(timeout)
	for tries := 0; time.Now().Before(deadline); tries++ {
		_, err := http.Head(url)
		if err == nil {
			return nil // success
		}
		log.Printf("server not responding (%s); retrying...", err)
		time.Sleep(time.Second << uint(tries)) // 指数退避策略
	}
	return fmt.Errorf("server %s failed to respond after %s", url, timeout)
}
```

3.如果还是不能顺利进行下去，调用者能够输出错误，然后优雅地停止程序，但一般这样的处理应该留给主程序部分

> 一种好办法是通过log.Fatalf

4.在一些错误情况下，仅记录下错误信息，然后程序继续运行

> 可以通过log包记录

5.在罕见的情况下，直接安全地忽略掉这个错误

#### 5.4.2文件结束标识

1.由文件结束引起的读取错误：io.EOF

### 5.5 函数变量

### 5.6 匿名函数

1.闭包的理解：函数变量不仅是一段代码还可以拥有状态。

> 里层的匿名函数能够获取和更新外层函数的局部变量。这些隐藏的变量引用就是我们把函数归类为引用类型而且函数变量无法进行比较的原因。Go程序员通常把函数变量称为闭包

2.深度优先和广度优先的两个示例

3.for each的一些注意事项，采用for each的方式进行循环时，创建的是一个可访问的存储位置，而不是固定的值。这个位置的值会不断地在迭代中更新。

> 可以通过引入一个内部变量来解决该问题
>
> ```go
> for _,dir := range tempDir(){
>     dir := dir//内部申明dir，并以外部dir初始化
> }
> ```
>
> 

### 5.7 变长函数

1.变长函数的类型和一个带有普通slice参数的函数类型不同

```
func f(...int){}
func g([]int){}
fmt.Printf("%T\n",f)//func(...int)
fmt.Printf("%T\n",g)//func([]int)
```



### 5.8 延迟函数调用

1.defer也可以用来调试一个复杂的函数，即在函数的“入口”和“出口”设置调试行为

> 书中的提示，别忘了给defer后的函数加上额外的圆括号。否则入口操作会在函数退出时执行，而出口操作永远不会调用。关于入口出口的详情，P113。
>
> ```go
> func bigSlowOperation() {
> 	defer trace("bigSlowOperation")() // don't forget the extra parentheses
> 	// ...lots of work...
> 	time.Sleep(10 * time.Second) // simulate slow operation by sleeping
> }
> 
> func trace(msg string) func() {
> 	start := time.Now()
> 	log.Printf("enter %s", msg)
> 	return func() { log.Printf("exit %s (%s)", msg, time.Since(start)) }
> }
> ```

2.延迟执行的匿名函数可以观察到函数的返回结果

3.延迟执行的匿名函数能改变外层函数返回给调用者的结果

```go
func double(x int)(result int){
	defer func(){result += x}()
	return x
}
```



### 5.9 宕机

1.每一个goroutine都会在宕机的时候显示一个函数调用的栈跟踪消息（常用于诊断）

### 5.10 恢复

1.不应该尝试恢复从另外一个包内发生的宕机

> 有选择性的使用recover，p119预期的宕机

## 第6章 方法

### 6.1 方法声明

### 6.2 指针接收者的方法

### 6.3 通过结构体内嵌组成类型

```go
type Point struct{ X, Y float64 }

type ColoredPoint struct {
	Point
	Color color.RGBA
}

//!-decl

func (p Point) Distance(q Point) float64 {
	dX := q.X - p.X
	dY := q.Y - p.Y
	return math.Sqrt(dX*dX + dY*dY)
}

func (p *Point) ScaleBy(factor float64) {
	p.X *= factor
	p.Y *= factor
}
```

1.Point的方法都被纳入到ColoredPoint类型中

> 注意调用方法的时候：
>
> ```go
> red := color.RGBA{255, 0, 0, 255}
> blue := color.RGBA{0, 0, 255, 255}
> var p = ColoredPoint{Point{1, 1}, red}
> var q = ColoredPoint{Point{5, 4}, blue}
> fmt.Println(p.Distance(q.Point)) // "5"
> p.ScaleBy(2)
> q.ScaleBy(2)
> fmt.Println(p.Distance(q.Point)) // "10"
> ```

### 6.4 方法变量与表达式

### 6.5 示例：位向量

1.位向量作为一种数据结构，可以很好的实现一个专门设计过的集合

### 6.6 封装

## 第7章 接口

### 7.1 接口即约定

1.可以把一种类型替换为满足同一接口的另一种类型的特性称为可取代性

### 7.2 接口类型

1.io.Writer是一个广泛使用的接口，负责所有可以写入字节的类型的抽象

> 包括文件、内存缓冲区、网络连接、HTTP客户端、打包器(archiver)、散列器(hasher)等
>
> io包的其他有用接口：
>
> Reader抽象了所有可以读取字节的类型
>
> Closer抽象了所有可以关闭的类型（比如文件或者网络连接）
>
> ```go
> type Writer interface {
> 	Write(p []byte) (n int, err error)
> }
> type Reader interface {
> 	Read(p []byte) (n int, err error)
> }
> type Closer interface {
> 	Close() error
> }
> ```

2.可以通过组合已有接口得到新接口

```go
type ReadWriter interface{
	Reader
	Writer
}
```

### 7.3 实现接口

1.与基于类的语言（它们显式地声明了一个类型实现的所有接口）不同，Go中可以在需要的时候才定义新的抽象和分组，并且不用修改原有类型的定义

> 理解一波：以后写代码写着写着，发现需要抽象了，这时候再定义接口？

### 7.4 使用flag.Value来解析参数

### 7.5 接口值

1.接口：动态类型和值

2.一个接口值是否为nil取决于它的动态类型

3.一般来讲，在编译时无法知道一个接口值的动态类型会是什么，所以通过接口来做调用必然需要使用**动态分发**

> 编译器必须生成一段代码来从类型描述符拿到其中的某个方法地址，再间接调用该方法地址

4.两个接口值都是nil或者二者的动态类型完全一致且二者动态值相等，则两个接口相等

> 因为接口可比较，所以可以作为map键，也可以作为switch的操作数

5.如果两个接口动态类型一致，对应的动态值是不可比较的，则这个比较会以崩溃的方式失败

6.空的接口值（不含任何信息）和仅仅动态值为nil的接口值是不一样的

### 7.6 使用sort.Interface来排序

1.排序算法需要知道三个信息：序列长度、比较两个元素的含义、如何交换两个元素

```go
package sort
type Interface interface{
	Len() int
	Less(i,j int) bool//i,j是序列元素的下标
	Swap(i,j int)
}
```

2.字符串slice的排序太常见了，sort包提供了一个直接排序的Strings函数，即sort.Strings(xxx)

3.sort.Reverse函数值得一看，因为它使用了一个重要概念：组合

> reverse的Less方法直接调用了内嵌的sort.Interface值的Less方法，但只交换传入的下标，就可以颠倒排序结果

### 7.7 http.Handler接口

（本节可以学到目前写的服务器代码是如何一步步简化的，看似简单的实现，里面包含了很多隐藏点）

1.http.Handler

```go
/*
实现了Handler接口的对象可以注册到HTTP服务端，为特定的路径及其子树提供服务。
ServeHTTP应该将回复的头域和数据写入ResponseWriter接口然后返回。返回标志着该请求已经结束，HTTP服务端可以转移向该连接上的下一个请求。
*/
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
/*
书：
ListenAndServe函数需要一个服务器地址，比如"localhost:8000"，以及一个Handler接口的实例（用来接收所有请求）
Go中文网：
ListenAndServe监听TCP地址addr，并且会使用handler参数调用Serve函数处理接收到的连接。handler参数一般会设为nil，此时会使用DefaultServeMux。
*/
func ListenAndServe(addr string, handler Handler) error
```

2.net/http包提供了一个请求多工转发器ServeMux，用来简化URL和处理程序之间的关联

> 在使用这个之前，需要通过switch语句对req.URL.Path进行case划分，书上的例子如：case “/list”:XXX
>
> 一个ServeMux把多个http.Handler组合成单个http.Handler。Web服务器把请求分发到一个http.Handler，而不用管后面具体的类型是什么

3.http.HandlerFunc

```go
/*
Go中文网：
HandlerFunc type是一个适配器，通过类型转换让我们可以将普通的函数作为HTTP处理器使用。如果f是一个具有适当签名的函数，HandlerFunc(f)通过调用f实现了Handler接口。
*/
type HandlerFunc func(ResponseWriter, *Request)
/*
Go中文网：
ServeHTTP方法会调用f(w, r)
*/
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request)
```

> HandlerFunc演示了Go接口机制的一些不常见特性：
>
> 1）他是一个函数类型	2）他拥有自己的方法	3）他满足http.Handler
>
> 他的ServeHTTP方法就调用函数本身，所以HandlerFunc就是一个让函数值满足接口的一个**适配器**

4.

想要实现：有两个不同的Web服务器，在不同的端口监听，定义不同的URL，分发到不同的处理程序。

只需：简单地构造另一个ServeMux，再调用一次ListenAndServe即可

> 然而，大多数场景下，一个Web服务就够了。另外，一个程序可能在很多文件中来定义HTTP处理程序，如果每次都要显式注册在应用本身的ServeerMux实例上，那就太麻烦了
>
> 因此，net/http提供了一个全局ServeMux实例：DefaultServeMux，以及包级别的注册函数http.Handle和http.HandleFunc。要让DefaultServeMux作为服务器的主处理程序，无须把他传给ListenAndServe，直接传nil即可
>
> 所以现在都直接这样写，但是背后的逻辑需要知道
>
> ```go
> func main() {
> 	db := database{"shoes": 50, "socks": 5}
> 	http.HandleFunc("/list", db.list)
> 	http.HandleFunc("/price", db.price)
> 	log.Fatal(http.ListenAndServe("localhost:8000", nil))
> }
> ```

5.Web服务器每次都用一个新的goroutine来调用处理程序，所以处理程序必须要注意并发问题

### 7.8 error接口

1.一个更易用的封装函数fmt.Errorf（意思就是少用errors.New）

2.syscall里定义了其他表示错误的方式

> Errno是一个系统调用错误的高效表示手法

### 7.9 示例：表达式求值器

暂时未细看

### 7.10 类型断言

1.作用在接口上的操作

> 类型断言会检查操作数的动态类型是否满足指定的断言类型
>
> 格式：x.(T)，x是一个接口类型表达式，T是一个类型（称为断言类型）
>
> 1）如果T是一个具体类型，类型断言会检查x的动态类型是否就是T
>
> 换言之，类型断言就是用来从它的操作数中把具体类型的值提取出来的操作
>
> 2）如果T是一个接口类型，类型断言会检查x的动态类型是否满足T
>
> 类似java的多态

### 7.11 使用类型断言来识别错误

1.处理错误比较幼稚的实现：通过检查错误消息是否包含特定的字符串

> 可靠的方法：用专门的类型来表示结构化的错误值（感觉是需要API发布者提供）

### 7.12 通过接口类型断言来查询特性

1.io包的WriteString

> 向io.Writer写入字符串的推荐方法
>
> （这个规范有点君子协议，隐性的约定）

### 7.13类型分支

1.接口的两种风格

> **第一种风格**：比如io.Writer、io.Reader、http.Handler，
>
> 接口上的各种方法突出满足这个接口的具体类型之间的相似性，但隐藏了各个具体类型的布局和各自特有的功能。
>
> 强调方法，而不是具体类型
>
> **第二种风格**：
>
> 充分利用接口值可以容纳各种具体类型的能力这一特点，把接口作为这些类型的联合（union）来使用
>
> 类型断言用来在运行时区分这些类型并分别处理
>
> 强调满足接口的具体类型，而不是方法（经常都没有方法），也不注重信息隐藏，这种风格称为**可识别联合**

### 7.14 示例：基于标记的XML解析 

1.在xml中，如果需要构造一个完整文档树的结构，那么通过Marshal、Unmarshal这种函数非常方便，但是对于很多程序来说，这是不必要的。因此encoding/xml为解析API提供了基于标记的底层XML

> 解析器读入输入文本，输出一个标记流（主要包含四种类型：StartElement、EndElement、CharData、Comment），每次调用(*xml.Decoder).Token都会返回一个标记

### 7.15 一些建议

## 第8章 goroutine和通道

1.go中有两种并发风格

> 1）通信顺序进程：Communicating Sequential Process，CSP（本章）
>
> 2）共享内存多线程（下章）

### 8.1 goroutine

1. 除了从main返回或者退出程序之外，没有程序化的方法让一个goroutine来停止另一个，但有办法和goroutine通信来要求他自己停止

### 8.2 示例：并发时钟服务器

### 8.3 示例：并发回声服务器

### 8.4 通道

1. 同种类型的通道可以用==符号比较

> 当二者都是同一通道数据的引用时，比较值为true。通道也可和nil比较



2. 通道的操作：发送，接收，关闭

> 1）关闭后的发送操作将导致宕机
>
> 2）在一个已经关闭的通道上进行接收操作，会获得所有已经发送的值，直到通道为空。这时任何接收操作会立即完成，同时获取到一个通道元素类型对应的零值。

#### 8.4.1 无缓冲通道

1.通道发送消息时，可以分两类场景

> 1）通信本身以及通信发生的时间也很重要，当我们强调这方面时，把消息叫做事件（event）
>
> 2）事件没有携带额外信息时，它单纯的目的是进行同步。通常用一个struct{}元素类型的通道来强调它

#### 8.4.2 管道

1. 当关闭的通道被读完后，所有的后续接收操作顺畅进行，只是获取到的是零值

2. 如何知道通道关闭呢？

   > 当然可以通过:
   >
   > ```go
   > for{
   >     x,ok := <-naturals
   >     if !ok {
   >         break	//通道关闭且读完
   >     }
   >     //一些操作
   > }
   > ```
   >
   > 这种笨笨的方式，但是因为场景比较通用，golang里可以直接用range()
   >
   > ```go
   > for x := range naturals {
   > 	//一些操作
   > }
   > ```

   3.通道可以通过垃圾回收器根据它**是否可以访问**来决定是否回收它，而不是是否关闭

#### 8.4.3 单向通道类型

1.通道作为形参时，基本上都是被限制着不能发送或者不能接收

#### 8.4.4 缓冲通道

1.很好的场景：并发地向三个镜像地址发送请求，只接受第一个返回的请求

```go
func mirroredQuery() string {
    //这里如果不用缓冲通道的话，另外两个较慢的请求会被卡住，这种情况叫gorotine泄露
	responses := make(chan string, 3)	
    
	go func() { responses <- request("asia.gopl.io") }()
	go func() { responses <- request("europe.gopl.io") }()
	go func() { responses <- request("americas.gopl.io") }()
    //返回最快的
	return <-responses 
}
func request(hostname string) (response string){ /*...*/ }
```



### 8.5 并发循环

1.高度并行：由一些完全独立的子任务组成的问题

> 注意：在for循环中调用协程，最好都显示传参，因为for循环会对迭代值不断进行更新

2.为了防止goroutine泄露

> 1) 在知道协程确切数量的情况下，可以直接申请足够大的缓冲区
>
> 2) 在不知道确切数量的情况下，考虑sync.WaitGroup
>
> 一个食用trick，在这种不知道确切数量的时候，可以单独启用一个协程：
>
> ```go
> go func() {
> 	wg.wait()
> 	close(sizes)
> }
> //然后主goroutine中消费
> for sizes := range sizes{
> 	//操作
> }
> ```
>
> 



### 8.6 示例：并发的Web爬虫

1.使用容量为n的缓冲通道来建立一个并发原语，称为**计数信号量**

> 保持信号量操作离它所约束的I/O操作越近越好

2.范围限制可以帮助我们推导程序的正确性

> 信息隐藏帮助限制程序不同部分不经意的交互

### 8.7 使用select多路复用

1.除了context以外的一个时间相关demo

> ```go
> select {
>     case <- time.After(10*time.Second):
>     case <- anthorChan:
> }
> 
> tick := time.Tick(1*time.Second)
> select {
>     case <- tick:
>     case <- anthorChan:
> }
> 
> //然而，上述情况返回时，计时器还在运行，会造成goroutine泄露
> ticker := time.NewTicker(1*time.Second)
> <-ticker.C	//从ticker的通道接收
> ticker.Stop()	//造成ticker的goroutine终止
> ```

2.如果多个情况同时满足，select随机选择一个

3.如果不用default，好像会造成持续阻塞？所以可以加上什么都不做的default，变成非阻塞的接收操作

4.对于select，如果通道是nil，将永远不会被选择

### 8.8 示例：并发目录遍历

1.这块内容完全可以自己研究下，感觉很好玩，相当于在cmd中自己整一个自己想要的功能

### 8.9 取消

1.对于取消操作，需要一个可靠的机制：在一个通道上广播一个事件

> 利用通道关闭后的机制：接收操作立即返回，得到零值
>
> 利用上述机制创造一个广播机制：不在通道上发送值，而是关闭它，相关demo:
>
> ```go
> for {
> 		select {
> 		case <-done:
> 			// Drain fileSizes to allow existing goroutines to finish.
> 			for range fileSizes {
> 				// Do nothing.
> 			}
> 			return
> 		case size, ok := <-fileSizes:
> 			// ...
> 			//!-3
> 			if !ok {
> 				break loop // fileSizes was closed
> 			}
> 			nfiles++
> 			nbytes += size
> 		case <-tick:
> 			printDiskUsage(nfiles, nbytes)
> 		}
> 	}
> ```
>
> 

### 8.10 示例：聊天服务器

todo

## 第9章 使用共享变量实现并发

### 9.1 竟态

1.并发安全的定义

> 在一个串行程序中正确工作的函数，如果在**并发调用**时仍然能正确工作，那么这个函数是**并发安全**的
>
> 上述的”并发调用“是指：在没有额外同步机制的情况下，从两个或多个goroutine同时调用这个函数。

2.对于绝大多数变量，如果要回避并发访问的话，有两种做法:

> 1）限制变量只存在于一个goroutine内
>
> 2）维护一个更高层的**互斥不变量**

3.导出的包级别函数通常可以认为是并发安全的

> 因为包级别的变量无法限制在一个goroutine内，所以那些修改这些变量的函数就必须采用互斥机制

4.竟态是指在多个goroutine按某些交错顺序执行时程序无法给出正确的结果（而且很难再现）

5.**数据竟态**：发生于两个goroutine并发读写同一个变量且至少其中一个是写入时

> **定义非常重要！**
>
> 当数据竟态的变量类型是大于一个机器字长的类型时，会更加复杂
>
> 根据定义，有三种方法来避免数据竟态
>
> 1）不要修改变量
>
> 在创建其他goroutine之前就用完整的数据初始化map，并且不再修改
>
> 2）避免从多个goroutine访问同一个变量
>
> 其他goroutine无法直接访问相关变量，因此它们必须使用通道来向首先goroutine发送查询请求或者更新变量
>
> GO箴言：”不要通过共享内存来通信，通过通信来共享内存“
>
> 即使一个变量无法在整个生命周期受限于单个goroutine，加以限制仍然可以是解决并发问题的好办法，例如：可以通过借助通道来把共享变量的地址从上一步传到下一步，从而在多个goroutine间共享该变量。流水线的每一步里，在把变量地址传给下一步后就不再访问该变量了，这样所有对这个变量的访问都是串行的（**串行受限**）
>
> 3）允许多个goroutine访问，但同一时间内只有一个goroutine可以访问（**互斥机制**）

### 9.2 互斥锁：sync.Mutex

1.sync.Mutex可以理解为：一个容量为1的通道来保证同一时间最多只有一个goroutine能访问共享变量（**二进制信号量**）

> Lock和Unlock之间的代码可以自由地读取和修改共享变量，称为**临界区域**
>
> 函数、互斥锁、变量的组合方式称为：监控（monitor）模式
>
> 锁函数的话可以：
>
> ```go
> func XXX() int {
> 	mu.Lock()
> 	defer mu.Unlock()
> 	return xxx
> }
> ```

2.处理并发程序时，永远应当优先考虑清晰度，并且拒绝过早优化

3.Go语言的互斥量是不可再入的。互斥量的目的是在程序执行过程中维持基于共享变量的**特定不变量**（invariant）

> 其中一个不变量是：“没有goroutine正在访问这个共享变量”，但有可能互斥量也保护针对数据结构的其他不变量。
>
> 当goroutine获取一个互斥锁时，会假定这些不变量是满足的。当它获取互斥锁后，更新变量的值时会临时不满足之前的不变量，但是释放变量时，必须保证之前的不变量已经还原且又能重新满足。尽管一个可重入的互斥量可以保证没有其他goroutine可以访问共享变量，但是无法保护这些变量的其他不变量。（解释了为啥不可再入）

4.一个常见demo

> ```go
> func Deposit(amount int){
> 	mu.Lock()
> 	defer mu.Unlock()
> 	deposit(amount)
> }
> //这个函数要求已获得互斥锁
> func deposit(amout int){balance += amount}
> ```
>
> 上面这种拆分可以让deposit()完成实际业务逻辑

5.使用一个互斥量时，确保互斥量本身以及被保护的变量都没有导出

### 9.3 读写互斥锁：sync.RWMutex

1.仅在绝大部分goroutine都在获取读锁并且锁竞争比较激烈时，RWMutex才有优势。

（即goroutine一般都需要等待后才能获取 到锁）

> 因为RWMutex需要复杂的内部簿记工作，所以在竞争不激烈时它比普通的互斥锁慢

### 9.4 内存同步

1.对于读写互斥锁的解读有两个：

> 1）防止Balance操作插到其他操作中间（结合具体例子）
>
> 2）同步不仅涉及多个goroutine的执行顺序问题，同步还会影响内存
>
> 通道通信或者互斥锁操作这样的同步原语会让处理器把累计的写操作刷回内存并提交

2.这些并发问题都可以通过简单成熟的模式来避免

> 在可能的情况下，把变量限制到单个goroutine中
>
> 对于其他变量，使用互斥锁

### 9.5 延迟初始化：sync.Once

1.sycn包提供了针对一次性初始化问题的特化解决方案：sync.Once

> Once包含一个布尔变量和一个互斥量，布尔变量记录初始化是否完成，互斥量负责保护这个布尔变量和客户端的数据结构
>
> Once唯一的方法Do以初始化函数作为参数
>
> 
>
> Dome :
>
> ```go
> var loadIconsOnce sync.Once
> var icons map[string]image.Image
> //并发安全
> func Icon(name string) image.Image {
> 	loadIconsOnce.Do(loadIcons)//loadIcons:初始化逻辑
> 	return icons[name]
> }
> ```
>
> 
>
> 每次调用Do时会先锁定互斥量并检查里面的布尔类型。在第一次调用时，这个布尔类型为假，Do会调用loadIcons然后把变量设置为真。后续相当于空操作，只是通过互斥量的同步来保证loadIcons对内存产生的效果对所有goroutine可见

### 9.6 竟态检测器

1.无法保证肯定不会发生竟态

2.把-race加到go build, go test后面即可使用

### 9.7 示例：并发非阻塞缓存

1.两种方案处理并发

> 1）共享变量上锁
>
> 2）通信顺序进程

todo

### 9.8 goroutine与线程

#### 9.8.1 可增长的栈

#### 9.8.2 goroutine调度

#### 9.8.3 GOMAXPROCS

#### 9.8.4 goroutine没有标识

## 第10章 包和go工具

## 第11章 测试