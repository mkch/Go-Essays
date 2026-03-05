package main

import (
	"fmt"
	"math"
)

func main() {
	absToleranceDemo()
	absToleranceFailureDemo()
	relToleranceDemo()
	relToleranceFailureDemo()
	isCloseDemo()
}

func absToleranceDemo() {
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

}

func absToleranceFailureDemo() {
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
		fmt.Println("product 和 110000000.0 被认为是相等的")
	} else {
		fmt.Println("product 和 110000000.0 被认为是不相等的") // 输出这一行
	}
}

// relClose 使用相对容差来比较两个浮点数是否接近
func relClose(a, b, tolerance float64) bool {
	diff := math.Abs(a - b)
	// 考虑a和b的大小，计算相对容差
	tolerance = tolerance * math.Max(math.Abs(a), math.Abs(b))
	fmt.Printf("diff: %.20f tolerance: %.20f\n", diff, tolerance)
	return diff <= tolerance
}

func relToleranceDemo() {
	const relTolerance = 1e-9
	a := 0.1
	b := 0.2
	sum := a + b

	if relClose(sum, 0.3, relTolerance) {
		fmt.Println("相等") // 输出这一行
	} else {
		fmt.Println("不相等")
	}

	a = 100000000.0
	b = 1.1
	product := a * b
	if relClose(product, 110000000.0, relTolerance) {
		fmt.Println("相等") // 输出这一行
	} else {
		fmt.Println("不相等")
	}
}

func relToleranceFailureDemo() {
	const relTolerance = 1e-9
	a := 1e-10
	temp := a + 0.3
	b := temp - 0.3
	fmt.Printf("%.20f\n", a) // 输出 0.00000000010000000000
	fmt.Printf("%.20f\n", b) // 输出 0.00000000010000000827
	if relClose(a, b, relTolerance) {
		fmt.Println("相等")
	} else {
		fmt.Println("不相等") // 输出这一行
	}
}

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

// mixedClose2 是 mixedClose 的一个变种，直接将绝对容差和相对容差结合在一起进行比较
func mixedClose2(a, b, absTolerance, relTolerance float64) bool {
	diff := math.Abs(a - b)
	return diff <=
		absTolerance+relTolerance*math.Max(math.Abs(a), math.Abs(b))
}

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

func isCloseDemo() {
	a := 0.1
	b := 0.2
	sum := a + b
	if !isClose(sum, 0.3, 1e-8, nil) {
		panic("应该相等")
	}

	if isClose(sum, 0.25, 1e-8, nil) {
		panic("应该不相等")
	}

	a = 100000000.0
	b = 1.1
	product := a * b
	if !isClose(product, 110000000.0, 1e-5, new(1e-8)) {
		panic("应该相等")
	}

	if isClose(product, 115000000.0, 1e-5, nil) {
		panic("应该不相等")
	}
}
