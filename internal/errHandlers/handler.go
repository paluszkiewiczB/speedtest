package errHandlers

import "log"

func NewPrintln() *Handler {
	l := &funcDelegatingHandler{d: log.Println}
	return &Handler{l: l}
}

func New(logger Logger) *Handler {
	return &Handler{l: logger}
}

func NewFuncDelegating(f func(v ...interface{})) *Handler {
	return &Handler{l: &funcDelegatingHandler{d: f}}
}

type Handler struct {
	l Logger
}

func (h Handler) Handle(err error) {
	h.l.Log(err)
}

type Logger interface {
	Log(v ...interface{})
}

type funcDelegatingHandler struct {
	d func(v ...interface{})
}

func (f *funcDelegatingHandler) Log(v ...interface{}) {
	f.d(v...)
}
