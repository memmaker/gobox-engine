package importpod

func (it *Item) Name() string {
	n := int(it.PName[0])
	if n > 0x28 {
		n = 0x28
	}
	return string(it.PName[1 : n+1])
}
