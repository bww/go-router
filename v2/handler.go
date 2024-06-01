package router

// Route handler
type Handler func(*Request, Context) (*Response, error)

// Middle provides functionality to wrap a handler producing another handler
type Middle interface {
	Wrap(Handler) Handler
}

// Middleware function wrapper
type MiddleFunc func(Handler) Handler

func (m MiddleFunc) Wrap(h Handler) Handler {
	return m(h)
}
