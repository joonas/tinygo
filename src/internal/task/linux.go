//go:build linux && !baremetal

package task

import "unsafe"

type pthread_mutex struct {
	// 40 bytes on a 64-bit system, 24 bytes on a 32-bit system
	state1 uint64
	state2 [4]uintptr
}

// pthread_mutex_t and pthread_cond_t are both initialized to zero in
// PTHREAD_*_INITIALIZER.

type sem struct {
	// 64 bytes on 64-bit systems, 32 bytes on 32-bit systems:
	//     volatile int __val[4*sizeof(long)/sizeof(int)];
	state [4 * unsafe.Sizeof(uintptr(0)) / 4]int32
}
