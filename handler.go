package router

// Route handler
type Handler func(*Request, Context) (*Response, error)

// Middle provides functionality to wrap a handler producing another handler
type Middle interface {
	Wrap(Handler) Handler
}

// Middle function wrapper
type MiddleFunc func(Handler) Handler

// MiddleFunc conforms to Middle by calling itself with the handler
func (m MiddleFunc) Wrap(h Handler) Handler {
	return m(h)
}

// A set of middleware
type Middles []Middle

// Middles is itself conforms to Middle and wraps a handler with all
// of its sub-middleware
func (m Middles) Wrap(h Handler) Handler {
	for _, e := range m {
		h = e.Wrap(h)
	}
	return h
}
