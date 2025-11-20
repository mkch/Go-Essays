# 如何实现“互斥接口”

## 接口

在Go语言中，接口（interface）与实现（implementation）是松耦合的。一个类型只要具备一个接口所要求的方法，就被认为实现了该接口。

```go
import "fmt"

type WidgetID string

type Widget interface {
    // ID returns the unique identifier of the widget.
    ID() WidgetID
}

type myWidget struct {
    id WidgetID
}

// ID method implements [Widget] interface.
func (w *myWidget) ID() WidgetID {
    return w.id
}

// String method implements [fmt.Stringer] interface.
func (w *myWidget) String() string {
    return fmt.Sprintf("myWidget(%q)", w.id)
}

// myWidget implements [Widget] interface.
var w Widget = &myWidget{"widget1"}
```

在上述代码中，`myWidget`类型拥有`Widget`接口要求的`ID() WidgetID`方法，于是它就实现了`Widget`接口。`myWidget`并没和`Widget`耦合，甚至删除`Widget`接口也不会影响`myWidget`的完整性，`myWidget`只是“凑巧”实现了`Widget`而已。

同样地，`myWidget`也“凑巧”实现了`fmt.Stringer`接口。一个类型同时实现多个接口在Go语言代码中是普遍现象，例如`*os.File`就同时实现了`io.Reader`、`io.Writer`以及`io.ReaderAt`等众多接口。

## 互斥接口悖论

在某些情况下，多个接口之间在语义上是互斥的。例如声明式（Declarative）UI中的`StatefulWidget`和`StatelessWidget`：一个Widget不可能（也不应该）既是有状态的（stateful）又同时是无状态的（stateless）。但是，由于Go语言中的实现和接口是松耦合的，这种语义上的“互斥性”在语法层面无法保证。

```go
type WidgetState interface {
    // Build builds the widget with its state.
    Build() Widget
}

type StatefulWidget interface {
    Widget
    // CreateState creates the state for the widget.
    CreateState() WidgetState
}

type StatelessWidget interface {
    Widget
    // Build builds the widget.
    Build() Widget
}
```

在上述代码中，`StatefulWidget`和`StatelessWidget`两个接口的区别在于前者要求实现`CreateState() WidgetState`方法， 而后者要求实现`Build() Widget`方法。但是它们并没有（不能）阻止一个用户类型同时实现这两个互斥的接口。

```go
// wiredState is a [WidgetState] implementation.
type wiredState struct{}

func (s *wiredState) Build() Widget {
    return &myWidget{id: "wired-stateful"}
}

// wiredWidget is a [StatefulWidget] and [StatelessWidget] implementation.
type wiredWidget struct {
    id WidgetID
}

// ID methods implements [Widget] interface.
func (w *wiredWidget) ID() WidgetID {
    return w.id
}

// CreateState method and ID method implement [StatefulWidget] interface.
func (w *wiredWidget) CreateState() WidgetState{
    return &wiredState{}
}

// Build method and ID method implement [StatelessWidget] interface.
func (w *wiredWidget) Build() Widget {
    return &myWidget{id: "wired-stateless"}
}
```

在上述代码中，`wiredWidget`正如其名是一个奇怪的Widget，它既是`StatefulWidget`也是`StatelessWidget`。这就给UI框架后续处理制造了麻烦。

```go
// Render is a hypothetical framework function that renders a widget.
func Render(widget Widget) {
    if stateful, ok := widget.(StatefulWidget); ok {
        state := stateful.CreateState()
        _ = state // Handle the state...
    } else if stateless, ok := widget.(StatelessWidget); ok {
        child := stateless.Build()
        _ = child // Handle the built child widget...
    } else {
        // Handle other widget types...
    }
}
```

当把一个`wiredWidget`传递给`Render`函数时，它只会被当作`StatefulWidget`处理，因为从`Render`或者整个框架的视角来看，一个`Widget`要么是stateful要么是stateless，不应该也不可能有一个同时处理二者的逻辑。但是从用户的视角来看就比较难以解释了：你既然允许（不阻止）我同时实现二者，那为什么不能同时处理呢？这也和“A well-designed API should be easy to use and hard to misuse” （一个设计良好的API应该易于使用且难以误用）相违背。

## 解决方案

上述难题其实有那么一个不太优雅但足够好用的解决方案。我们知道，Go语言是不允许函数（方法）重载（Overloading）的，即在给定上下文中不能存在两个名称相同但参数或返回值不同的函数。利用这点来可以实现多个接口之间的互斥。

```go
type StatefulWidget interface {
    Widget
    CreateState() WidgetState
    // Exclusive is a marker method to make [StatefulWidget] mutually exclusive with [StatelessWidget].
    Exclusive(StatefulWidget)
}

type StatelessWidget interface {
    Widget
    // Build builds the widget.
    Build() Widget
    // Exclusive is a marker method to make [StatefulWidget] mutually exclusive with [StatelessWidget].
    Exclusive(StatelessWidget)
}
```

我们为`StatefulWidget`接口增加一个`Exclusive(StatefulWidget)`方法，为`StatelessWidget`接口增加一个`Exclusive(StatelessWidget)`方法。由于一个类型不能拥有两个同名方法，所以绝对不可能写出一个同时实现这两个接口的类型。

用户类型如果想要实现`StatefulWidget`就添加一个`Exclusive(StatefulWidget)`方法：

```go
// Exclusive method implements [StatefulWidget].
func(w *customWidget)Exclusive(StatefulWidget){ /* NOP */}
```

如果想要实现`StatelessWidget`就添加一个`Exclusive(StatelessWidget)`方法:

```go
// Exclusive method implements [StatelessWidget].
func(w *customWidget)Exclusive(StatelessWidget){ /* NOP */}
```

但是无法同时拥有二者，因为Go的语法不允许。
