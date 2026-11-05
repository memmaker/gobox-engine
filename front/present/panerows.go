package present

import (
	"fmt"
	"time"

	"gobox/front"
)

// paged is v with the pane cut to its Rows lines from p.paneFrom (paneMore
// when lines follow; a key then moves on) and the pictures that cycle
// through their Frames at the frame the clock is on.
func (p *Presenter) paged(v front.View) front.View {
	p.paneMore = false
	if v.Pane != nil {
		v.Pane = wrapPane(v.Pane)
	}
	if pn := v.Pane; pn != nil && pn.Rows > 0 && len(pn.Lines) > pn.Rows {
		if sig := fmt.Sprint(pn.Lines); sig != p.paneSig {
			p.paneSig, p.paneFrom = sig, 0
		}
		from := min(p.paneFrom, len(pn.Lines)-pn.Rows)
		p.paneMore = from+pn.Rows < len(pn.Lines)
		c := *pn
		c.Lines = pn.Lines[from : from+pn.Rows]
		v.Pane = &c
	}
	for i, pc := range v.Pictures {
		if f, ok := spriteFrame(pc, time.Now()); ok { // a DAX animation block (PIC, SPRIT, FINAL): its frames at their own delays
			v.Pictures = append(v.Pictures[:i:i], v.Pictures[i:]...)
			v.Pictures[i].Frame = f
			pc = v.Pictures[i]
		}
		if len(pc.Frames) > 0 && pc.FrameMs > 0 {
			v.Pictures = append(v.Pictures[:i:i], v.Pictures[i:]...) // not the game's slice
			for j := i; j < len(v.Pictures); j++ {
				if q := v.Pictures[j]; len(q.Frames) > 0 && q.FrameMs > 0 {
					v.Pictures[j].Frame = q.Frames[int(time.Now().UnixMilli()/int64(q.FrameMs))%len(q.Frames)]
				}
			}
			break
		}
	}
	return v
}
