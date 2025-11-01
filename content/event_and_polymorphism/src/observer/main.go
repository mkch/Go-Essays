package main

import "fmt"

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

func main() {
	var window2 = NewMyWindow()
	// The behavior of window2 is overwritten
	window2.AddOnClickListener(func(x, y int) {
		fmt.Printf("Draw a dot at (%v, %v)\n", x, y)
	})

	window2.NotifyClick(100, 200)
}
