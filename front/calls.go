package front

// Calls is a queue of rule calls on the game's engine E, made by the
// screens and run by the app before the frame's actions.
type Calls[E any] struct{ q []func(E) }

// Add queues f.
func (c *Calls[E]) Add(f func(E)) { c.q = append(c.q, f) }

// Take hands over the calls queued since the last Take.
func (c *Calls[E]) Take() []func(E) {
	q := c.q
	c.q = nil
	return q
}
