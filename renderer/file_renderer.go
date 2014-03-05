package renderer

import (
	"bytes"
	"carousel/templates"
	"code.google.com/p/go.tools/present"
	"fmt"
	"github.com/suapapa/go_hangul/encoding/cp949"
	"html/template"
	"io"
	"io/ioutil"
	"unicode/utf8"
)

var _utf8_bom_header []byte = []byte{0xef, 0xbb, 0xbf}

type FileRenderer struct{}

func (rend *FileRenderer) GetRenderFunc(r io.Reader) (renderFunc RenderFunc, err error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	// validate document is utf-8
	if !utf8.Valid(b) {
		// it may be cp949, so convert it to utf-8
		var b2 []byte
		b2, err = cp949.From(b)
		if err != nil {
			return
		}

		b = b2
	}

	// determine document has BOM
	hasBom := false
	for i, bom := range _utf8_bom_header {
		if bom == b[i] {
			hasBom = true
		} else {
			hasBom = false
		}
	}

	// eliminate BOM if it exist
	if hasBom {
		b = b[len(_utf8_bom_header):]
	}

	// parse
	nr := bytes.NewBuffer(b)
	doc, err := present.Parse(nr, "slides", 0)
	if err != nil {
		err = fmt.Errorf("while parsing: %v", err.Error())
		return
	}

	tmpl := present.Template()
	tmpl = tmpl.Funcs(template.FuncMap{"playable": playable})

	tmpl, err = parseTemplates(tmpl, templates.Action_tmpl, templates.Slides_tmpl)
	if err != nil {
		err = fmt.Errorf("while templating: %v", err.Error())
		return
	}

	renderFunc = RenderFunc(func(w io.Writer) {
		doc.Render(w, tmpl)
	})

	return
}

func playable(c present.Code) bool {
	return false
}

func parseTemplates(t *template.Template, ss ...string) (*template.Template, error) {
	if len(ss) == 0 {
		return nil, fmt.Errorf("no arguments")
	}

	for i, s := range ss {
		tmpl := t.New(fmt.Sprintf("tmpl_%d", i))

		_, err := tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}
