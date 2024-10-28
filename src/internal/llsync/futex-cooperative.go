package llsync

import (
	"internal/task"
	_ "unsafe" // for //go:linkname
)

// A futex is a way for userspace to wait with the pointer as the key, and for
// another thread to wake one or all waiting threads keyed on the same pointer.
//
// A futex does not change the underlying value, it only reads it before to prevent
// lost wake-ups.
type Futex struct {
	Uint32
	waiters task.Stack
}

// Atomically check for cmp to still be equal to the futex value and if so, go
// to sleep. Return true if we were definitely awoken by a call to Wake or
// WakeAll, and false if we can't be sure of that.
func (f *Futex) Wait(cmp uint32) (awoken bool) {
	if f.Uint32.v != cmp {
		return false
	}

	// Push the current goroutine onto the waiter stack.
	f.waiters.Push(task.Current())

	// Pause until the waiters are awoken by Wake/WakeAll.
	task.Pause()

	// We were awoken by a call to Wake or WakeAll. There is no chance for
	// spurious wakeups.
	return true
}

// Wake a single waiter.
func (f *Futex) Wake() {
	if t := f.waiters.Pop(); t != nil {
		resumeTask(t)
	}
}

// Wake all waiters.
func (f *Futex) WakeAll() {
	for t := f.waiters.Pop(); t != nil; t = f.waiters.Pop() {
		resumeTask(t)
	}
}

//go:linkname resumeTask runtime.resumeTask
func resumeTask(*task.Task)
