package rules1e

// bytesOf copies src's values into dst (the typed Pool fields into a later
// layout's bytes), as copy does.
func bytesOf[T ~uint8](dst []byte, src []T) {
	for i := range min(len(dst), len(src)) {
		dst[i] = byte(src[i])
	}
}
