package main

import "fmt"

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

func main() {
	var window1 Window
	window1.SetOnClickListener(func(x, y int) {
		fmt.Printf("Draw a dot at (%v, %v)\n", x, y)
	})

	// Simulate a click event
	// Should be called when receiving a click event from the OS.
	window1.NotifyClick(100, 200)

	var window2 = CreateMyWindow()
	// The behavior of window2 is overwritten
	window2.SetOnClickListener(func(x, y int) {
		fmt.Printf("Draw a dot at (%v, %v)\n", x, y)
	})
	window2.NotifyClick(100, 200)
}

type MyWindow struct {
	Window
}

func CreateMyWindow() *MyWindow {
	var win MyWindow
	win.SetOnClickListener(func(x, y int) {
		fmt.Printf("Draw a solid circle at (%v, %v)\n", x, y)
	})
	return &win
}
