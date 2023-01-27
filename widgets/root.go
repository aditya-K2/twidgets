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
	m := &View01{}

	iv := NewInteractiveView()
	Root := tview.NewPages()
	Root.AddPage("iview", iv.View, true, true)

	m.Root = Root
	return m
}

func (m *View01) OpenCenteredWidget(t CenteredWidget) {
	p := *(t.Primitive())
	closec := make(chan bool)
	currentTime := time.Now().String()
	sHandler := t.SelectionHandler()
	_, _, w, h := m.Root.GetRect()

	close := func() {
		m.Root.RemovePage(currentTime)
	}
	draw := func() {
		m.Root.AddPage(currentTime, t.Primitive(), false, true)
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
			sHandler(
				p.GetCell(
					p.GetSelection()).Text)
			close()
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
						_, _, _w, _h := m.Root.GetRect()
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
