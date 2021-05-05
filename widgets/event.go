package widgets

type AppEvent int

const (
    QUIT_EVENT AppEvent = iota
    REDRAW_EVENT AppEvent = iota
)
