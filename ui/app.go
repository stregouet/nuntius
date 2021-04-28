package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type Application struct {
  tapp *views.Application
}

func NewApp() Application {
  return Application{tapp: &views.Application{}}
}
