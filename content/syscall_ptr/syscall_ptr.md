# Go调用syscall时指针的处理

## runtime.KeepAlive

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

## runtime.Pinner

用于固定一个Go对象在内存中的位置，阻止GC移动或释放它。主要用于在非Go内存中存储Go指针。Pinner是为了应对未来可能引入的moving GC（[issues/46787](https://github.com/golang/go/issues/46787)）。目前（2025年10月）Go并没有使用moving GC（[“Go has a non-moving GC”](https://go.dev/doc/gc-guide)）。被pin的对象不可能为栈内存，它必然escape到heap、必然被keep alive直到unpin。

## //go:uintptrescapes

用于指定一个函数的`uintptr`类型的参数实际上是Go对象指针（从Go指针强制转换而来），该对象会espace到heap且keep alive直到该函数调用结束。参看此Directive的[文档](https://pkg.go.dev/cmd/compile#hdr-Compiler_Directives)。

## 总结

### 约束

这三者中`runtime.Pinner`的约束最强，包含pin、keep alive和escape to heap。

`//go:uintptrescapes`次之，只包含keep alive和escape to heap。

`runtime.KeepAlive`最弱。

### 使用场景

`runtime.Pinner`用于调用没有使用`//go:uintptrescapes`修饰的的syscall函数，且指针所指Go对象不会espcape、调用之后不再有引用的场景。

`//go:uintptrescapes`用于标定必须把Go指针转换为uintptr传递（受限于被调用函数的原型）且被调用函数不会存储该指针到非Go内存的场景（因为该directive并不pin该指针）。

`runtime.Pinner`由于需要在非Go内存中存储Go指针的场景。

***注：在使用cgo时，[“Go pointers passed as function arguments to C functions have the memory they point to implicitly pinned for the duration of the call.”](https://pkg.go.dev/cmd/cgo#hdr-Passing_pointers)***
