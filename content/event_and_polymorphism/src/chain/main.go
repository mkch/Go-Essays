package main

import "fmt"

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
