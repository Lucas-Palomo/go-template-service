package template

import (
	"bytes"
	"fmt"
	"github.com/open2b/scriggo"
	"github.com/open2b/scriggo/native"
	"io"
	"io/fs"
	"os"
	"sync"
)

type Service struct {
	templates map[string]string
	viewData  native.Declarations
	dir       fs.FS
	mutex     sync.RWMutex
}

func New(dir string) *Service {
	return &Service{
		dir: os.DirFS(dir),
		viewData: native.Declarations{
			"ctx": (*native.Declarations)(nil),
		},
		templates: make(map[string]string),
	}
}

func (svc *Service) AddTemplate(name string, path string) error {
	if _, ok := svc.templates[name]; ok {
		return fmt.Errorf("template already registered: %s", name)
	}

	_, err := svc.dir.Open(path)
	if err != nil {
		return err
	}

	svc.templates[name] = path
	return nil
}

func (svc *Service) AddFunc(name string, fn native.Declaration) {
	svc.mutex.Lock()
	svc.viewData[name] = fn
	svc.mutex.Unlock()
}

func (svc *Service) Render(name string, context native.Declarations, writer io.Writer) error {
	path, ok := svc.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	svc.mutex.Lock()
	for key, value := range svc.viewData {
		context[key] = value
	}
	svc.mutex.Unlock()

	template, err := scriggo.BuildTemplate(
		svc.dir,
		path,
		&scriggo.BuildOptions{
			Globals: context,
		},
	)

	if err != nil {
		return err
	}

	return template.Run(
		writer,
		nil,
		nil,
	)
}

func (svc *Service) RenderAsString(name string, context native.Declarations) (string, error) {
	var buf bytes.Buffer

	err := svc.Render(name, context, &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
