# 事件、观察者、责任链和多态

## 事件

在事件驱动的系统中，我们经常使用回调函数来接收和处理事件。例如GUI框架中的一个窗口是这个样子：

```go
type Window struct {
    // A callback for click events
    onClick func(x, y int)
}

// SetOnClickListener registers a callback for click events.
func (w *Window) SetOnClickListener(listener func(x, y int)) {
    w.onClick = listener
}

// NotifyClick notifies a click event to the registered callback.
func (w *Window) NotifyClick(x, y int) {
    if w.onClick != nil {
        w.onClick(x, y)
    }
}
```

用户可以创建一个窗口，并编写一个Listener函数来接收并处理鼠标单击消息：

```go
func main() {
    var window1 Window
    window1.SetOnClickListener(func(x, y int) {
        fmt.Printf("Draw a dot at (%v, %v)\n", x, y)
    })

    // Simulate a click event
    // Should be called when receiving a click event from the OS.
    window1.NotifyClick(100, 200)
}
```

这个程序会在单击消息发生时，在鼠标所点位置上绘制一个圆点。

## 观察者

到目前为止一切都简单而美好，但不要忘了`Window`类型是GUI*框架*中的一部分，一个设计良好的框架除了要简单易用，还要兼顾可扩展性。上述`Window`类型也许做到了简单易用，但和良好的可扩展性基本不沾边。例如，用户要封装一个自己的窗口类型`MyWindow`，它在鼠标单击事件发生时在单击位置绘制一个实心圆圈，大概可以这么做：

```go
type MyWindow struct {
    Window
}

func NewMyWindow() *MyWindow {
    var win MyWindow
    win.SetOnClickListener(func(x, y int) {
        fmt.Printf("Draw a solid circle at (%v, %v)\n", x, y)
    })
    return &win
}
```

这样以来，`MyWindow`的实例确实可以在单击位置绘制一个实心圆圈，可是它占用了OnClickListener，为它的用户造成了不便：用户如果使用`SetOnClickListener`来接收事件，则必将覆盖掉`MyWindow`的默认行为。用户想要在单击时先画一个实心圆再在圆心处画一个不同颜色的点，就只能把已经在`MyWindow`里写好的“画圆”代码重新写一遍（或者，更有可能是复制粘贴过来）。

很明显，这个`Window`类型的设计使用了极简的观察者模式（Observer Pattern），它只允许单一观察者。似乎完整实现观察者模式就能解决上述问题：

```go
type Window struct {
    // Observers for click events
    onClickListeners []func(x, y int)
}

// AddOnClickListener adds a listener for click events to the listener list.
func (w *Window) AddOnClickListener(listener func(x, y int)) {
    w.onClickListeners = append(w.onClickListeners, listener)
}

// NotifyClick notifies a click event to the registered listeners.
func (w *Window) NotifyClick(x, y int) {
    for _, listener := range w.onClickListeners {
        listener(x, y)
    }
}
```

这次`Window`使用`AddOnClickListener`代替了之前的`SetOnClickListener`，并使用一个slice来保存所有的Listener。`NotifyClick`遍历这个slice并逐一调用所有Listener。

`MyWindow`类型也做出相应地修改：

```go
type MyWindow struct {
    Window
}

func NewMyWindow() *MyWindow {
    var win MyWindow
    win.AddOnClickListener(func(x, y int) {
        fmt.Printf("Draw a solid circle at (%v, %v)\n", x, y)
    })
    return &win
}
```

如此，`MyWindow`的用户也可以自由地接收、处理事件了：

```go
func main() {
    var window2 = NewMyWindow()
    // The behavior of window2 is overwritten
    window2.AddOnClickListener(func(x, y int) {
        fmt.Printf("Draw a dot at (%v, %v)\n", x, y)
    })

    window2.NotifyClick(100, 200)
}
```

运行这个代码，单击鼠标，窗口上会显示一个实心圆和圆心处的一个点。当然，这一切都只能存在于我们的脑海中，实际上我们看到的只是如下两行输出：

```
Draw a solid circle at (100, 200)
Draw a dot at (100, 200)
```

先绘制圆，再绘制点，点会显示在圆之上。

## 责任链

到目前为止一切都简单而美好，但不要忘了，软件开发中永远不变的就是“变化”。很快，用户就提出了新的需求：我要使用`MyWindow`但是在鼠标单击时显示一个穿在线上的圆，线要画在圆之下，看起来就像穿糖葫芦一样”。也许对用户来说“糖葫芦”意味着儿时的美好时光，但对我们来说“糖葫芦”意味着要改变画圆和画点的顺序。这个需求在目前的框架下是无法实现的，因为`Window.NotifyClick`是按照Listener加入的顺序依次调用的，用户无法在调用`NewMyWindow`之前调用`AddOnClickListener`。除非……“把`MyWindow`的代码复制过去”。当然，修改`Window.NotifyClick`采用倒序遍历也是行不通的，因为这是“拆东墙补西墙”。

造成当前困境的核心原因在于`Window`的设计不够灵活，无法应对当前和未来可能出现的变化。如果能设计一种方案让用户既可以添加Listener还能决定Listener的调用顺序，那就一劳永逸了。考虑到用户可能会“犯傻”，所以也不能给他太多的自由，毕竟一个好的系统既要好用又要“不容易被误用”。于是“决定Listener的调用顺序”就降级为“决定自己的Listener和其他Listener之间的调用顺序”，这里“其他Listener”作为一个整体其调用顺序并不由当前用户决定。

修改后的`Window`是这个样子的：

```go
// ClickHandler is a handler for click events.
// A handler can call next to pass the event to the next handler.
type ClickHandler func(x, y int, next func(x, y int))

type Window struct {
    // A chain of handlers for click events.
    onClickChain func(x, y int)
}

// AddOnClickHandler adds a handler to the click event handler chain.
func (w *Window) AddOnClickHandler(handler ClickHandler) {
    old := w.onClickChain
    w.onClickChain = func(x, y int) {
        next := func(x, y int) {
            if old != nil {
                old(x, y)
            }
        }
        handler(x, y, next)
    }
}

// NotifyClick notifies pass a click event to the handler chain.
func (w *Window) NotifyClick(x, y int) {
    if w.onClickChain != nil {
        w.onClickChain(x, y)
    }
}
```

为了让用户能决定何时调用“其他Handler”，我们把Listener升级为`Handler`类型。`Handler`最后一个参数`next`是一个函数，调用这个函数意味着调用“其他Handler”，或者说调用在当前Handler被添加之前已经添加的Handler。这样以来，Handler的编写者可以在自己的代码中自由地决定何时调用next，或者压根不掉用next。

所有被添加的Handler被保存在`Window.onClickChain`字段中。你没看错，`onClickChain`是函数类型，而不是链表之类。为什么可以用一个函数来保存一个“链”？这是因为Go的函数字面量是闭包（Closure），闭包中除了函数代码还包括代码用到的外部变量。如果一个函数闭包中包含了第二个函数类型的变量，那个变量（也是一个闭包）又包含了第三个函数类型的变量，以此类推，是不是就形成了一个“链条”？此时如果第一个函数调用了它所包含的第二个函数，而第二个函数调用了它包含的第三个函数……于是所有函数都被依次执行。

上面一段所描述的东西看起非常复杂，如同“意大利通心粉”一样剪不断理还乱。但是正如某知名人士所说“Shut the XX up and show me the code”，软件开发中的很多些东西是不能用自然语言而非代码来清晰描述的。这个听起来烧脑的过程是由`Window.AddOnClickHandler`来实现的，怎么样，看到代码是不是一下子就清晰了？至少，每一行代码都能看懂了。

`AddOnClickHandler`首先把当前的`w.onClickChain`保存到`old`中，然后设置`w.onClickChain`为一个新的函数。这个新的函数中调用了用户提供的handler，并把next作为最后一个参数传递给它。这个next也是一个函数，它本质上就是调用old。

如此以来，当`AddOnClickHandler`返回时，`w.onClickChain`被替换为一个新的函数（记作f），当f被执行时，它首先调用用户提供的handler，用户在handler中可以调用next，而next调用之前的（old）`w.onClickChain`，也就是`AddOnClickHandler`之前的“其他Handler”。由于用户可以自行决定何时调用next，于是问题完美解决：

```go
type MyWindow struct {
    Window
}

func NewMyWindow() *MyWindow {
    var win MyWindow
    win.AddOnClickHandler(func(x, y int, next func(x, y int)) {
        fmt.Printf("Draw a solid circle at (%v, %v)\n", x, y)
    })
    return &win
}

func main() {
    var window2 = NewMyWindow()
    // Users are free to call next as they wish
    window2.AddOnClickHandler(func(x, y int, next func(x, y int)) {
        fmt.Printf("Draw a line from (%v, %v) to (%v, %v)\n", x-10, y, x+10, y)
        next(x, y)
    })

    window2.NotifyClick(100, 200)
}
```

虽然用户依然必须在调用`NewMyWindow`之后才能调用`AddOnClickHandler`，但是他在Handler中可以自行决定何时调用`MyWindow`的默认实现（next）。于是，当用户单击鼠标时，成功画出了压在线上的圆，也就是他最爱吃的糖葫芦。当然，这一切都只能存在于我们的脑海中，实际上我们看到的只是如下两行输出：

```text
Draw a line from (90, 200) to (110, 200)
Draw a solid circle at (100, 200)
```

先绘制一根稍长的直线，再在其上绘制一个实心圆，aka. 糖葫芦。当然，如果用户不喜欢糖葫芦了，想画一个Ø也不是不可能。

为了突出重点，这个例子以及上面的例子都没有给出圆的直径，毕竟这些绘图操作需要我们去脑补，那就不不妨再脑补出一个合适的直径吧。

熟悉设计模式的朋友可能早就看出来了，这种“链式”处理事件的方式，正是责任链模式（Chain of Responsibility Pattern）的一个简单变种。

## 中间件

将这种链式消息处理模式应用于HTTP服务器，即可便捷地实现中间件（Middleware）架构：

```go
package main

import (
    "io"
    "log"
    "net/http"
    "time"
)

type Handler func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)

type Middleware http.HandlerFunc

func (m *Middleware) Use(handler Handler) {
    old := *m
    *m = func(w http.ResponseWriter, r *http.Request) {
        next := func(w http.ResponseWriter, r *http.Request) {
            if old != nil {
                old(w, r)
            }
        }
        handler(w, r, next)
    }
}

func main() {
    // Middleware to handle root path "/"
    var rootMiddleware Middleware = func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, "Hello!")
    }

    // Add Authorization middleware.
    rootMiddleware.Use(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
        if r.FormValue("user") != "admin" {
            code := http.StatusUnauthorized
            http.Error(w, http.StatusText(code), code)
            return
        }
        next(w, r)
    })

    // Add log middleware
    rootMiddleware.Use(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
        start := time.Now()
        next(w, r)
        duration := time.Since(start)
        log.Printf("Path: %v Duration: %v", r.URL.Path, duration)
    })


    http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        rootMiddleware(w, r)
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`Middleware.Use`和之前的`Window.AddOnClickHandler`几乎一模一样，都是实现Handler链式存储的基础。

得益于可在Handler中自由使用（或不使用）next，log中间件分别在处理HTTP请求之前和之后添加代码用于计时，鉴权中间件在鉴权失败后输出Unauthorized错误代码而不再处理HTTP请求。

## 多态

Go是不是一门面向对象（OO）语言？这个问题可能见仁见智，但Go确实可以实现OO的三要素：封装、继承、多态。其中封装最为直观：使用非导出的（unexported，即小写字母开头）标识符来实现信息隐藏。继承可以大致类比到Go中`struct`的嵌入，例如在`MyWindow`中嵌入`Window`以复用Window的代码。多态稍微复杂一些。

所谓“多态”就是“一个接口，多种实现”。当然，这里的“接口”是泛指代码中的“合约或规范”，并不严格对应Go中的`interface`。这里的“实现”是指代码中遵循合约或规范的行为所产生的后果（执行的代码）。

回到本文最开头使用回调函数来接收事件的例子，一个`Window`对象可以调用`NotifyClick`方法这就是`Window`类型的一个“接口”。通过调用`SetOnClickListener`为`Window`对象设置不同的回调以改变`NotifyClick`方法被调用后的行为，就是所谓的“多种实现”。合起来就是“多态”。

上一节关于中间件的例子中，`Middleware`类型的变量可以使用适当类型的`w`和`r`参数来调用就是它的“接口”，调用后实际执行的log或鉴权行为就是它的“实现”。通过`Use`不同的Handler来改变它的行为，就是所谓的“多态”。

在上述责任链和中间件的例子中，对`next`的调用和面向对象语言中通过`super`或其他方式对父类同名虚函数的调用有异曲同工之处。

在本文的例子中，多态是通过为一个`func`类型的变量赋不同的值来实现的，实际上也可以用`interface`来替代`func`类型。Go里的`func`类型就是退化了的只有一个方法的`interface`。Go里面的多态必须有`interface`或`func`类型的参与，或者说没有使用`interface`和`func`类型的Go代码是不具有多态性的。
