package renderer

import (
	"io"
)

type renderFunc func(io.Writer) error

type Renderer interface {
	Render(w io.Writer) error
	Refresh() error
}
