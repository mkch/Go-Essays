package main

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

type WidgetState interface {
	// Build builds the widget with its state.
	Build() Widget
}

type StatefulWidget interface {
	Widget
	// CreateState creates the state for the widget.
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

// wiredState is a [WidgetState] implementation.
type wiredState struct{}

func (s *wiredState) Build() Widget {
	return &myWidget{id: "wired-stateful"}
}

// wiredWidget implements neither [StatefulWidget] nor [StatelessWidget]
// due to the absence of Exclusive methods.
type wiredWidget struct {
	id WidgetID
}

// ID methods implements [Widget] interface.
func (w *wiredWidget) ID() WidgetID {
	return w.id
}

func (w *wiredWidget) CreateState() WidgetState {
	return &wiredState{}
}

func (w *wiredWidget) Build() Widget {
	return &myWidget{id: "wired-stateless"}
}

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

func main() {
	fmt.Println(w)
}
