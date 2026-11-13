package engine

// Lists (ovr26:0D9A, remade): a script's list (askList) publishes the
// lines as the Ask (AskList); scrolling, paging and the highlight are the
// presenter's.

// ListNode is a line of a list: header lines (Flag non-zero) and Dim ones
// are not picked.
type ListNode struct {
	Text  string
	Flag  byte
	Dim   bool // (extra) greyed out and not picked, as headers
	Grey  bool // (extra) greyed out but picked like any line (a creation drop-down's entry that forces changes)
	Spell int  // (extra) the line's spell, 0 for none: the presenter shows its rules
	// Without (extra): options left out while this line is the one chosen
	Without []string
}

// skipped: the line is not picked.
func (p *ListNode) skipped() bool { return p.Flag != 0 || p.Dim }

// newList = ovr26:1140: n empty nodes (n<2 gives 1 node).
func newList(n int) []*ListNode {
	if n < 1 {
		n = 1
	}
	l := make([]*ListNode, n)
	for i := range l {
		l[i] = &ListNode{}
	}
	return l
}
