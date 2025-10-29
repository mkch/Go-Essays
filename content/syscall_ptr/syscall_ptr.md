# Go调用syscall时指针的处理

在Go语言中使用syscall调用外部函数时，需要保证传递给syscall的指针所引用的内存在syscall完成之前不被GC释放和移动。为达成此保证，可使用如下3种手段。

## 1）runtime.KeepAlive

用于阻止finalizer过早运行。Go编译器可能会在一个变量没有超出作用域但已没有事实引用的位置插入对finalizer的调用。`KeepAlive(x)`告诉编译器对象`x`在此之前不能被释放、它的finalizer不会被调用。
KepAlive虽然是一个函数，但其本质上是一个Compiler Directive：

```go
func KeepAlive(x any) {
    if cgoAlwaysFalse {
        println(x)
    }
}
```

其中`cgoAlwaysFalse`是一个永远为`false`的`bool`变量。

## 2）runtime.Pinner

用于固定一个Go对象在内存中的位置，阻止GC移动或释放它。主要用于在非Go内存中存储Go指针。Pinner是为了应对未来可能引入的moving GC。目前（2025年10月）Go并没有使用moving GC[^1]。

被pin的对象必然escape、必然被keep alive直到unpin。

## 3）//go:uintptrescapes

用于指定一个函数的`uintptr`类型的参数实际上是Go对象指针（从Go指针强制转换而来），该对象会espace到heap且keep alive直到该函数调用结束。参看此Directive的[文档](https://pkg.go.dev/cmd/compile#hdr-Compiler_Directives)。

此类函数时会用到诸如`syscall.Syscall(…… uintptr(unsafe.Pointer(p)) ……)`的形式，而`unsafe.Pointer`的[转换规则](https://pkg.go.dev/unsafe#Pointer)(4)中提到“编译器在处理传递给用汇编实现的函数的参数时，如果参数是从`unsafe.Pointer`转换为`uintptr`的，则会保证此指针所引用的对象在调用结束前不被释放和移动（is retained and not moved until the call completes）”。据此可推导出，被标定为`//go:uintptrescapes`的函数，如果正确地使用unsafe转换来调用syscall，那么它的参数所引用的内存也会和cgo调用[^2]一样被pin。

## 总结

### 约束

这三者中`runtime.Pinner`的约束最强，包含pin、keep alive和escape to heap。

`//go:uintptrescapes`如果被正确使用，也包含pin、keep alive和escape to heap三重约束。

`runtime.KeepAlive`最弱。

### 使用场景

`runtime.KeepAlive`用于保证finalizer不会被过早调用。这种情况一般发生在obj的一部分被传递（复制）给了syscall，并且在syscall之后没有其他对obj的引用的场合。此时如果不在syscall之后使用`runtime.KeepAlive(obj)`，则obj的finalizer可能在syscall完成之前被调用。参看`runtime.KeepAlive`的[文档](https://pkg.go.dev/runtime#KeepAlive)以及这个可运行的[例子](https://go.dev/play/p/ebh1TIrgtDO)。

`//go:uintptrescapes`用于标定必须把Go指针转换为uintptr传递给syscall的场景。

`runtime.Pinner`最为灵活，可以手工控制何时pin以及何时unpin，用于需要在非Go内存中存储Go指针的场景，例如在Windows中把Go指针作为`GWLP_USERDATA`调用[SetWindowLongPtr](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowlongptra)。

[^1]: [“Go has a non-moving GC.”](https://go.dev/doc/gc-guide#Tracing_Garbage_Collection)

[^2]: [Go pointers passed as function arguments to C functions have the memory they point to implicitly pinned for the duration of the call.](https://pkg.go.dev/cmd/cgo#hdr-Passing_pointers)
