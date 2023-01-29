package widgets

import (
	"errors"

	"github.com/aditya-K2/tview"
	"github.com/gdamore/tcell/v2"
)

var (
	berr         = errors.New("Couldn't Get Base Selection in Interactive View")
	defaultfg    = tcell.ColorGreen
	defaultbg    = tcell.ColorDefault
	Defaultstyle = tcell.StyleDefault.
			Foreground(defaultfg).
			Background(defaultbg)
	OutOfBound = -1
)

type _range struct {
	Start int
	End   int
}

type InteractiveView struct {
	visual        bool
	disableVisual bool
	vrange        *_range
	baseSel       int
	View          *tview.Table
	capture       func(e *tcell.EventKey) *tcell.EventKey
	vhandler      func(start int, end int)
	visualCapture func(start, end int, e *tcell.EventKey) *tcell.EventKey
	content       func() [][]*tview.TableCell
}

func GetCell(text string, st tcell.Style) *tview.TableCell {
	return tview.NewTableCell(text).
		SetAlign(tview.AlignLeft).
		SetStyle(st)
}

// f should return [][]*tview.TableCell that is then used to set
// the content of the View.
func (i *InteractiveView) SetContentFunc(f func() [][]*tview.TableCell) {
	i.content = f
}

// Sets Input Capture. Default Keys for Interactive View can't be
// overridden.
func (i *InteractiveView) SetCapture(
	f func(e *tcell.EventKey) *tcell.EventKey) {
	i.capture = f
}

func (i *InteractiveView) SetVisualCapture(f func(start, end int, e *tcell.EventKey) *tcell.EventKey) {
	i.visualCapture = f
}

func (i *InteractiveView) DisableVisualMode(disable bool) {
	i.disableVisual = disable
}

func NewInteractiveView() *InteractiveView {
	view := tview.NewTable()
	view.SetSelectable(true, false)
	view.SetBackgroundColor(tcell.ColorDefault)

	i := &InteractiveView{
		View:          view,
		vrange:        &_range{},
		visual:        false,
		disableVisual: false,
	}

	_capture := func(e *tcell.EventKey) *tcell.EventKey {
		if i.pcapture(e) != nil {
			return i.capture(e)
		}
		return nil
	}

	draw := func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		i.View.Clear()
		s := i.content()
		for _i, v := range s {
			if !i.disableVisual {
				b := ""
				if i.visual && (_i >= i.vrange.Start && _i <= i.vrange.End) {
					b = "[blue::]█[::]"
				}
				// Copying the Same Style for the visual block as the
				// first cell
				_b := *s[_i][0]
				if !_b.NotSelectable {
					i.View.SetCell(_i, 0, _b.SetText(b))
				}
			}
			for _j := range v {
				i.View.SetCell(_i, _j+1,
					v[_j])
			}
		}
		return i.View.GetInnerRect()
	}
	i.View.SetDrawFunc(draw)
	view.SetInputCapture(_capture)
	return i
}

func (i *InteractiveView) exitVisualMode() {
	if i.vrange.Start < i.baseSel {
		i.View.Select(i.vrange.Start, -1)
	} else if i.vrange.End > i.baseSel {
		i.View.Select(i.vrange.End, -1)
	}
	i.vrange.Start = OutOfBound
	i.vrange.End = OutOfBound
	i.baseSel = OutOfBound
}

func (i *InteractiveView) enterVisualMode() {
	row, _ := i.View.GetSelection()
	i.baseSel = row
	i.vrange.Start, i.vrange.End = row, row
}

func (i *InteractiveView) ExitVisualMode() {
	if i.visual {
		i.exitVisualMode()
		i.visual = !i.visual
	}
}

func (i *InteractiveView) toggleVisualMode() {
	if i.visual {
		i.exitVisualMode()
	} else if !i.visual {
		i.enterVisualMode()
	}
	i.visual = !i.visual
}

func (i *InteractiveView) getHandler(
	s string) func(e *tcell.EventKey) *tcell.EventKey {
	vr := i.vrange
	check := func() {
		if vr.Start <= -1 {
			vr.Start = 0
		}
		if vr.End <= -1 {
			vr.End = 0
		}
		if vr.End >= i.View.GetRowCount() {
			vr.End = i.View.GetRowCount() - 1
		}
		if vr.Start >= i.View.GetRowCount() {
			vr.Start = i.View.GetRowCount() - 1
		}
	}
	funcMap := map[string]func(e *tcell.EventKey) *tcell.EventKey{
		"up": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				check()
				if vr.End > i.baseSel {
					vr.End--
				} else if vr.Start <= i.baseSel {
					vr.Start--
				}
				if i.baseSel == -1 {
					panic(berr)
				}
			}
			// Skip Non Selectable Rows
			r, c := i.View.GetSelection()
			if r > 0 && i.View.GetCell(r-1, c).NotSelectable {
				for r = r - 1; r > 0 && i.View.GetCell(r, c).NotSelectable; r-- {
				}
				// If the cell that is about to be selected is first one
				// and isn't selectable just remain where you are.
				if r == 0 && i.View.GetCell(0, 1).NotSelectable {
					return nil
				}
				i.View.Select(r, c)
				return nil
			}
			return e
		},
		"down": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				check()
				if vr.Start < i.baseSel {
					vr.Start++
				} else if vr.Start == i.baseSel {
					vr.End++
				}
				if i.baseSel == -1 {
					panic(berr)
				}
			}
			// Skip Non Selectable Rows
			// Using 1 instead of c here because sometimes visual column is selected
			// which doesn't return correct value for NotSelectable
			r, _ := i.View.GetSelection()
			if r < i.View.GetRowCount()-1 && i.View.GetCell(r+1, 1).NotSelectable {
				for r = r + 1; r < i.View.GetRowCount()-1 && i.View.GetCell(r, 1).NotSelectable; r++ {
				}
				i.View.Select(r, 1)
				// If the cell that is about to be selected is first one
				// and isn't selectable just remain where you are.
				if r == i.View.GetRowCount()-1 && i.View.GetCell(r, 1).NotSelectable {
					return nil
				}
				return nil
			}
			return e
		},
		"exitvisual": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				i.exitVisualMode()
				i.visual = false
				return nil
			}
			return e
		},
		"top": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				i.vrange.Start = 0
				i.vrange.End = i.baseSel
				i.View.ScrollToBeginning()
			}
			return e
		},
		"bottom": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				i.vrange.Start = i.baseSel
				i.vrange.End = i.View.GetRowCount() - 1
				i.View.ScrollToEnd()
			}
			return e
		},
	}
	if val, ok := funcMap[s]; ok {
		return val
	} else {
		return nil
	}
}

// Default Capture Method. Can not be overridden.
func (i *InteractiveView) pcapture(e *tcell.EventKey) *tcell.EventKey {
	switch e.Rune() {
	case 'j':
		{
			return i.getHandler("down")(e)
		}
	case 'k':
		{
			return i.getHandler("up")(e)
		}
	case 'v':
		{
			i.toggleVisualMode()
			return nil
		}
	case 'g':
		{
			return i.getHandler("top")(e)
		}
	case 'G':
		{
			return i.getHandler("bottom")(e)
		}
	default:
		{
			if e.Key() == tcell.KeyEscape {
				return i.getHandler("exitvisual")(e)
			}
			if i.visual {
				if i.visualCapture != nil {
					if i.visualCapture(i.vrange.Start, i.vrange.End, e) == nil {
						i.toggleVisualMode()
					}
				}
			}
			return e
		}
	}
}
