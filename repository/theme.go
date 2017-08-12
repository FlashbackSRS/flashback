package repo

import (
	"html/template"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/controllers"
)

// Theme is a wrapper around a fb.Theme object
type Theme struct {
	*fb.Theme
}

// FuncMap returns the model controller's FuncMap, if any.
func (c *PouchCard) FuncMap(face int) (template.FuncMap, error) {
	mc, err := c.getModelController()
	if err != nil {
		return nil, err
	}
	if funcMapper, ok := mc.(controllers.FuncMapper); ok {
		return funcMapper.FuncMap(c.Card, face), nil
	}
	return nil, nil
}
