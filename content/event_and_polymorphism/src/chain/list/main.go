package main

import "fmt"

// ClickHandler is a handler for click events.
// A handler can call next to pass the event to the next handler.
type ClickHandler func(x, y int, next func(x, y int))

// onClickChainItem is a list item of the chain handler list.
type onClickChainItem struct {
	Handler ClickHandler
	Next    *onClickChainItem
}

// Call passes a click event to a handler item.
func (item *onClickChainItem) Call(x, y int) {
	if item == nil {
		return
	}
	item.Handler(x, y, item.Next.Call)
}

type Window struct {
	// A list(chain) of handlers for click events.
	onClickChainHead *onClickChainItem
}

// AddOnClickHandler adds a handler to the click event handler chain.
func (w *Window) AddOnClickHandler(handler ClickHandler) {
	old := w.onClickChainHead
	w.onClickChainHead = &onClickChainItem{Handler: handler, Next: old}
}

// NotifyClick notifies pass a click event to the handler chain.
func (w *Window) NotifyClick(x, y int) {
	w.onClickChainHead.Call(x, y)
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
