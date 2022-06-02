## 第一章

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

## 第二章

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

## 第三章

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

### 3.Printf()小技巧

> %后的[1]表示重复使用第一个操作数，#表示输出前缀

```
o := 0666
fmt.Printf("%d %[1]o %#[1]o\n",o)	//438 666 0666
```

