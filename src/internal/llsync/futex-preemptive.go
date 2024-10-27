//go:build scheduler.threads

package llsync

// A futex is a way for userspace to wait with the pointer as the key, and for
// another thread to wake one or all waiting threads keyed on the same pointer.
//
// A futex does not change the underlying value, it only reads it before to prevent
// lost wake-ups.
type Futex struct {
	Uint32
}
