package template

import (
	"bytes"
	"fmt"
	"github.com/open2b/scriggo"
	"github.com/open2b/scriggo/native"
	"io"
	"io/fs"
	"os"
)

type Service struct {
	templates map[string]*scriggo.Template
	dir       fs.FS
}

func New(dir string) *Service {
	return &Service{
		templates: make(map[string]*scriggo.Template),
		dir:       os.DirFS(dir),
	}
}

func (svc *Service) Register(name string, path string) error {
	tmpl, err := scriggo.BuildTemplate(
		svc.dir,
		path,
		&scriggo.BuildOptions{
			Globals: native.Declarations{
				"ctx": map[string]any{},
			},
		})

	if err != nil {
		return err
	}

	svc.templates[name] = tmpl
	return nil
}

func (svc *Service) Render(name string, context map[string]any, writer io.Writer) error {
	tmpl, ok := svc.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return tmpl.Run(writer, map[string]interface{}{"ctx": context}, nil)
}

func (svc *Service) RenderAsString(name string, context map[string]any) (string, error) {
	var buf bytes.Buffer

	err := svc.Render(name, context, &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
