//go:build gc.leaking

package runtime

// This GC implementation is the simplest useful memory allocator possible: it
// only allocates memory and never frees it. For some constrained systems, it
// may be the only memory allocator possible.

import (
	"internal/llsync"
	"unsafe"
)

// Ever-incrementing pointer: no memory is freed.
var heapptr llsync.Uintptr

// Total amount allocated for runtime.MemStats
var gcTotalAlloc llsync.Uint64

// Total number of calls to alloc()
var gcMallocs llsync.Uint64

// Total number of objected freed; for leaking collector this stays 0
const gcFrees = 0

// Inlining alloc() speeds things up slightly but bloats the executable by 50%,
// see https://github.com/tinygo-org/tinygo/issues/2674.  So don't.
//
//go:noinline
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	// TODO: this can be optimized by not casting between pointers and ints so
	// much. And by using platform-native data types (e.g. *uint8 for 8-bit
	// systems).
	size = align(size)

	// Track statistics. These are stored separately so are not strictly atomic,
	// which means that ReadMemStats might read a _slightly_ inconsistent state.
	gcTotalAlloc.Add(uint64(size))
	gcMallocs.Add(1)

	nextAddr := heapptr.Add(size)
	for nextAddr >= heapEnd {
		// Try to increase the heap and check again.
		if growHeap() {
			continue
		}
		// Failed to make the heap bigger, so we must really be out of memory.
		runtimePanic("out of memory")
	}
	addr := nextAddr - size

	pointer := unsafe.Pointer(addr)
	zero_new_alloc(pointer, size)
	return pointer
}

func realloc(ptr unsafe.Pointer, size uintptr) unsafe.Pointer {
	newAlloc := alloc(size, nil)
	if ptr == nil {
		return newAlloc
	}
	// according to POSIX everything beyond the previous pointer's
	// size will have indeterminate values so we can just copy garbage
	memcpy(newAlloc, ptr, size)

	return newAlloc
}

func free(ptr unsafe.Pointer) {
	// Memory is never freed.
}

// ReadMemStats populates m with memory statistics.
//
// The returned memory statistics are up to date as of the
// call to ReadMemStats. This would not do GC implicitly for you.
func ReadMemStats(m *MemStats) {
	totalAlloc := gcTotalAlloc.Load()
	m.HeapIdle = 0
	m.HeapInuse = totalAlloc
	m.HeapReleased = 0 // always 0, we don't currently release memory back to the OS.

	m.HeapSys = m.HeapInuse + m.HeapIdle
	m.GCSys = 0
	m.TotalAlloc = totalAlloc
	m.Mallocs = gcMallocs.Load()
	m.Frees = gcFrees
	m.Sys = uint64(heapEnd - heapStart)
	// no free -- current in use heap is the total allocated
	m.HeapAlloc = totalAlloc
	m.Alloc = m.HeapAlloc
}

func GC() {
	// No-op.
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// No-op.
}

func initHeap() {
	// preinit() may have moved heapStart; reset heapptr
	heapptr.Store(heapStart)
}

// setHeapEnd sets a new (larger) heapEnd pointer.
func setHeapEnd(newHeapEnd uintptr) {
	// This "heap" is so simple that simply assigning a new value is good
	// enough.
	heapEnd = newHeapEnd
}
