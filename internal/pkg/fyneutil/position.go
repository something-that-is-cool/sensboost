package fyneutil

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func RightBottomCorner(obj fyne.CanvasObject) fyne.CanvasObject {
	return container.NewBorder(nil, nil, nil,
		container.NewBorder(nil, obj, nil, nil))
}
