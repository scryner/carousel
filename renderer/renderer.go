package renderer

import (
	"io"
)

type RenderFunc func(io.Writer)

type Renderer interface {
	GetRenderFunc(r io.Reader) (RenderFunc, error)
}
