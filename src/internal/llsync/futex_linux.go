//go:build scheduler.threads

package llsync

import (
	"unsafe"
)

// Atomically check for cmp to still be equal to the futex value and if so, go
// to sleep. Return true if we were definitely awoken by a call to Wake or
// WakeAll, and false if we can't be sure of that.
func (f *Futex) Wait(cmp uint32) bool {
	tinygo_futex_wait((*uint32)(unsafe.Pointer(&f.Uint32)), cmp)

	// We *could* detect a zero return value from the futex system call which
	// would indicate we got awoken by a Wake or WakeAll call. However, this is
	// what the manual page has to say:
	//
	// > Note that a wake-up can also be caused by common futex usage patterns
	// > in unrelated code that happened to have previously used the futex
	// > word's memory location (e.g., typical futex-based implementations of
	// > Pthreads mutexes can cause this under some conditions). Therefore,
	// > callers should always conservatively assume that a return value of 0
	// > can mean a spurious wake-up, and use the futex word's value (i.e., the
	// > user-space synchronization scheme) to decide whether to continue to
	// > block or not.
	//
	// I'm not sure whether we do anything like pthread does, so to be on the
	// safe side we say we don't know whether the wakeup was spurious or not and
	// return false.
	return false
}

// Wake a single waiter.
func (f *Futex) Wake() {
	tinygo_futex_wake((*uint32)(unsafe.Pointer(&f.Uint32)), 1)
}

// Wake all waiters.
func (f *Futex) WakeAll() {
	const maxInt32 = 0x7fff_ffff
	tinygo_futex_wake((*uint32)(unsafe.Pointer(&f.Uint32)), maxInt32)
}

//export tinygo_futex_wait
func tinygo_futex_wait(addr *uint32, cmp uint32)

//export tinygo_futex_wake
func tinygo_futex_wake(addr *uint32, num uint32)
