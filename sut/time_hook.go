package sut

var SetTimeOffset = func(offset int) {
	panic(`
should be injected by main.
the side-effect will have per-thread visibility.
	`)
}