package present

// memo is the one shape of a value derived from the game's State or the
// front's state: built by one function, kept while its key stays (the State's
// Version, and the front-only inputs the value also reads), never rebuilt
// every frame.
type memo[K comparable, V any] struct {
	key K
	val V
	ok  bool
}

// get is the value for key: built by build when key is not the last one's.
func (m *memo[K, V]) get(key K, build func() V) V {
	return m.update(key, func(V) V { return build() })
}

// update is get for a value that carries something over from the last one
// (the zero V the first time): build gets it.
func (m *memo[K, V]) update(key K, build func(prev V) V) V {
	if !m.ok || m.key != key {
		m.key, m.val, m.ok = key, build(m.val), true
	}
	return m.val
}
