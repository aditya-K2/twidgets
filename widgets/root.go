package widgets

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type View01 struct {
	Root *tview.Pages
}

type CenteredWidget interface {
	Primitive() *tview.Table
	ContentHandler()
	SelectionHandler() func(s string)
	Size(mw, mh int) (int, int, int, int)
}

func NewView01() *View01 {
	v := &View01{}

	Root := tview.NewPages()
	v.Root = Root

	return v
}

func (v *View01) openCenteredWidget(t CenteredWidget) {
	p := *(t.Primitive())
	closec := make(chan bool)
	currentTime := time.Now().String()
	sHandler := t.SelectionHandler()
	_, _, w, h := v.Root.GetRect()

	close := func() {
		v.Root.RemovePage(currentTime)
	}
	draw := func() {
		v.Root.AddPage(currentTime, t.Primitive(), false, true)
		p.SetRect(t.Size(w, h))
	}
	redraw := func() {
		close()
		draw()
	}
	delete := func() {
		close()
		closec <- true
	}

	capture := func(e *tcell.EventKey) *tcell.EventKey {
		if e.Key() == tcell.KeyEscape {
			delete()
			return nil
		} else if e.Key() == tcell.KeyEnter {
			sHandler(p.GetCell(p.GetSelection()).Text)
			delete()
			return nil
		}
		return e
	}
	p.SetInputCapture(capture)

	t.ContentHandler()

	resizeHandler := func() {
		dur := 500
		tck := time.NewTicker(time.Duration(dur) * time.Millisecond)
		go func() {
			for {
				select {
				case <-tck.C:
					{
						_, _, _w, _h := v.Root.GetRect()
						if _w != w || _h != h {
							w = _w
							h = _h
							redraw()
						}
					}
				case <-closec:
					{
						return
					}
				}
			}
		}()
	}
	resizeHandler()

	draw()
}

func (v *View01) OpenListMenu(
	title string, list []string, shandler func(s string)) {
	m := newMenu()
	m.Content(list)
	m.SetSelectionHandler(shandler)
	m.Title(title)
	v.openCenteredWidget(m)
}
