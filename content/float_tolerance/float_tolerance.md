# 浮点数比较的容差

在Go语言以及其他大多数编程语言中，浮点数都是使用IEEE 754标准来表示的。这种表示方式使用有限二进制位数来表示实数，因此存在精度限制和舍入误差。这些误差可能导致浮点数的比较结果不符合预期，例如大家常见的`0.1 + 0.2 != 0.3`问题。

## 容差的概念

为了正确比较两个浮点数，编程中引入了一个概念叫做“容差”（tolerance）。容差是一个小的正数，用于定义两个浮点数比较相等时的误差容忍范围。换句话说，如果两个浮点数之间的绝对差值不大于这个容差，那么它们就被认为是相等的。

## 绝对容差

绝对容差（absolute tolerance）是最常见的容差比较方法，它直接比较两个浮点数之间的绝对差值是否小于等于一个预定义的容差值。例如在下列代码中使用`1e-8`作为绝对容差来比较`0.1+0.2`是否等于`0.3`：

```go
a := 0.1
b := 0.2
sum := a + b
fmt.Printf("%.20f\n", sum) // 输出0.30000000000000004441
fmt.Printf("%.20f\n", 0.3) // 输出0.29999999999999998890

// 直接比较两个浮点数，由于精度问题，结果不符合预期
fmt.Println(sum == 0.3) // 输出false

// 使用容差进行比较
const tolerance = 1e-8
if math.Abs(sum-0.3) <= tolerance {
    fmt.Println("相等") // 输出这一行
} else {
    fmt.Println("不相等")
}
```

绝对容差法非常简单直观，但是如何正确地选择容差是一个挑战。例如，同样的`tolerance`在下列代码中就会导致错误的结果。

```go
a := 100000000.0
b := 1.1
product := a * b
fmt.Printf("%.20f\n", product)     // 输出110000000.00000001490116119385
fmt.Printf("%.20f\n", 110000000.0) // 输出110000000.00000000000000000000

// 直接比较两个浮点数，由于精度问题，结果不符合预期
fmt.Println(product == 110000000.0) // 输出false

// 使用容差进行比较
const tolerance = 1e-8 // 过大的容差将导致错误的相等判断
if math.Abs(product-110000000.0) <= tolerance {
    fmt.Println("相等")
} else {
    fmt.Println("不相等") // 输出这一行
}
```

## 相对容差

在上个例子，绝对容差比较法之所以会失效，是因为IEEE 754浮点数的精度在不同量级的数值上表现不同。对于绝对值较小的值，浮点数的精度较高，可以选用较小的绝对容差；而对于绝对值于较大的值，浮点数的精度则较低，必须选择相对较大的容差。例如在0.3和-0.3附近，float64的精度约为`2.78e-17`，因此选择`1e-8`的绝对容差是足够的；但在110000000.0和-110000000.0附近，float64的精度约为`1.49e-8`，因此选择`1e-8`的绝对容差就过于严格了。可以看出，绝对容差的选择需要根据待比较数值的量级来调整，否则可能会导致错误的比较结果。

为了解决绝对容差过于生硬的问题，相对容差（relative tolerance）比较法就应运而生了。相对容差比较法把待比较值的大小纳入考量，比较的两个数之间的差值与它们的量级之间的比率。例如下列函数`relClose()`使用a和b的绝对值中较大值的量级来计算相对容差：

```go
// relClose 使用相对容差来比较两个浮点数是否接近
func relClose(a, b, tolerance float64) bool {
    diff := math.Abs(a - b)
    // 考虑a和b的大小，计算相对容差
    tolerance = tolerance * math.Max(math.Abs(a), math.Abs(b))
    fmt.Printf("diff: %.20f tolerance: %.20f\n", diff, tolerance)
    return diff <= tolerance
}
```

如果使用`1e-9`作为tolerance参数调用`relClose()`：

当待比较的值在`1e-1`的量级时（例如0.3）参与比较的容差大概为`1e-10`。此时既不会因为过于严格而导致0.1+0.2不等于0.3也不会因为过于宽松而导致0.1+0.2等于0.25。

当待比较的值在`1e8`的量级时(例如110000000.0)参与比较的容差大概为`1e-1`。此时既不会因为过于严格而导致100000000.0×1.1不等于110000000.0也不会因为过于宽松而导致100000000.0×1.1等于115000000.0。

## 混合容差

相对容差比较法虽然在不同量级的数值上表现更好，但在某些情况下也可能会失效。例如当待比较的数值非常接近于0时，参与比较的相对容差可能会非常小，从而导致不符合预期的比较结果。例如：

```go
const relTolerance = 1e-9
a := 1e-10
temp := a + 0.3
b := temp - 0.3
fmt.Printf("%.20f\n", a)
fmt.Printf("%.20f\n", b)
if relClose(a, b, relTolerance) {
    fmt.Println("相等")
} else {
    fmt.Println("不相等") // 输出这一行
}
```

由于a的值`1e-10`非常接近于0，`relClose()`函数计算出的相对容差也非常小（接近`1e-19`），导致a和b被认为是不相等的。

为了克服这个问题，我们可以使用混合容差（mixed tolerance）比较法，它结合了绝对容差和相对容差的优点，在比较两个浮点数时同时考虑它们的绝对差值和相对差值。例如下列函数`mixedClose()`使用绝对容差和相对容差来比较两个浮点数：

```go
// mixedClose 使用混合容差来比较两个浮点数是否接近
func mixedClose(a, b, absTolerance, relTolerance float64) bool {
    diff := math.Abs(a - b)
    // 先使用绝对容差进行比较
    if diff <= absTolerance {
        return true
    }
    // 再使用相对容差进行比较
    return diff <= relTolerance*math.Max(math.Abs(a), math.Abs(b))
}
```

在使用混合容差比较法的`mixedClose()`函数中，绝对容差比较和相对容差比较中只要有一个通过就返回true。这样在a和b非常接近于0时，绝对容差比较可以确保它们被认为是相等的；而在a和b的量级较大时，虽然绝对容差比较有可能漏报，但随后的相对容差比较可以确保它们被正确地比较。

## 一些变种

混合容差比较法在实际应用中有一个常见的变种：把绝对容差和相对容差相加后使用，例如：

```go
// mixedClose2 是 mixedClose 的一个变种，直接将绝对容差和相对容差结合在一起进行比较
func mixedClose2(a, b, absTolerance, relTolerance float64) bool {
    diff := math.Abs(a - b)
    return diff <=
        absTolerance+relTolerance*math.Max(math.Abs(a), math.Abs(b))
}
```

此时，绝对容差和相对容差的影响是叠加的。当a和b非常接近于0时，`relTolerance*math.Max(math.Abs(a), math.Abs(b))`也接近于0，此时`absTolerance`的影响占主导；当a和b的量级较大时，`relTolerance*math.Max(math.Abs(a), math.Abs(b))`的影响占主导。由于两个容差是相加后使用的，所以总体来说`mixedClose2()`的接近判别要比`mixedClose()`要宽松一些。

相对容差本质上相当于`relTolerance`乘以一个系数，这个系数的计算也存在多种选择，例如：

* 使用a和b其中任意一个的绝对值，通常使用b。此时b被看作是参考值，a是拿来和b做比较的：`tol * abs(b)`。
* 使用a和b的绝对值中的较大值。`tol * max( abs(a), abs(b) )`。即上述`mixedClose()`和`mixedClose2()`中使用的方法。
* 使用a和b的绝对值中的较小值。`tol * min( abs(a), abs(b) )`。
* 使用a和b的算数平均值的绝对值。`tol * abs(a + b)/2`。

## 容差的选择

选择合适的容差值是一个经验性的过程，通常需要根据具体的应用场景和数值的量级来调整。一般来说，`1e-9`或`1e-10`是常用的相对容差值。绝对容差值则需要根据待比较数值的具体含义来选择，并没有一个统一的标准。因此在编写一个通用的浮点数比较函数`isClose()`时，最好同时支持绝对容差和相对容差，并且允许用户根据需要进行配置。例如：

```go
// isClose 使用混合容差来比较两个浮点数是否接近。
// absTolerance是绝对容差。
// relTolerance指向相对容差，如果为nil则使用默认值 1e-9。
func isClose(a, b, absTolerance float64, relTolerance *float64) bool {
    relT := 1e-9
    if relTolerance != nil {
        relT = *relTolerance
    }
    diff := math.Abs(a - b)
    return diff <= absTolerance ||
        diff <= relT*math.Max(math.Abs(a), math.Abs(b))
}
```

上述函数的`relTolerance`参数是指向相对容差的指针，如果用户传入`nil`则使用默认值`1e-9`。这样用户在调用`isClose()`函数时既可以指定绝对容差又可以指定相对容差，也可以只指定绝对容差而使用默认的相对容差。在指定相对容差时，可以使用Go1.26引入的`new()`函数来创建一个指向float64类型的指针，例如:

```go
isClose(a, b, 1e-5, new(1e-8)) // 绝对容差为1e-5，相对容差为1e-8
```
