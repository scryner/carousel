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
	"os"
	"path/filepath"
	"unicode/utf8"
)

var _utf8_bom_header []byte = []byte{0xef, 0xbb, 0xbf}

type FileRenderer struct {
	filename    string
	rendFun     renderFunc
	playEnabled bool
	socketAddr  string
}

func NewFileRenderer(filename string, playEnabled bool, socketAddr string) *FileRenderer {
	return &FileRenderer{
		filename:    filename,
		playEnabled: playEnabled,
		socketAddr:  socketAddr,
	}
}

func (rend *FileRenderer) Render(w io.Writer) error {
	if rend.rendFun == nil {
		if err := rend.Refresh(); err != nil {
			return err
		}
	}

	return rend.rendFun(w)
}

func (rend *FileRenderer) Refresh() error {
	var err error
	rend.rendFun, err = getRenderFunc(rend.filename, rend.playEnabled, rend.socketAddr)

	return err
}

func getRenderFunc(filename string, playEnabled bool, socketAddr string) (rendFunc renderFunc, err error) {
	// read file
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
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

	// set playable
	if playEnabled {
		present.PlayEnabled = true
	} else {
		present.PlayEnabled = false
	}

	// parse
	nr := bytes.NewBuffer(b)
	doc, err := parseDocument(nr, filepath.Dir(filename), "slides", 0)
	if err != nil {
		err = fmt.Errorf("while parsing: %v", err.Error())
		return
	}

	// templating
	tmpl := present.Template()
	tmpl = tmpl.Funcs(template.FuncMap{"playable": playable})

	tmpl, err = parseTemplates(tmpl, templates.Action_tmpl, templates.Slides_tmpl)
	if err != nil {
		err = fmt.Errorf("while templating: %v", err.Error())
		return
	}

	rendFunc = renderFunc(func(w io.Writer) error {
		data := struct {
			*present.Doc
			Template    *template.Template
			PlayEnabled bool
			SocketAddr  string
		}{doc, tmpl, playEnabled, socketAddr}
		return tmpl.ExecuteTemplate(w, "root", data)
	})

	return
}

func playable(c present.Code) bool {
	return present.PlayEnabled && c.Play
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

func parseDocument(r io.Reader, dir, name string, mode present.ParseMode) (*present.Doc, error) {
	readFile := func(filename string) ([]byte, error) {
		return ioutil.ReadFile(filepath.Join(dir, filename))
	}

	ctx := present.Context{ReadFile: readFile}
	return ctx.Parse(r, name, mode)
}
