package ui

import (
	"github.com/stregouet/nuntius/lib"
	"github.com/stregouet/nuntius/widgets"
)


type Tab struct {
    Content widgets.Widget
    Title string
    machine *lib.Machine
}
