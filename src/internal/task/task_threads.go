//go:build scheduler.threads

package task

import (
	"sync/atomic"
	"unsafe"
)

// If true, print verbose debug logs.
const verbose = false

// Scheduler-specific state.
type state struct {
	// Goroutine ID. The number here is not really significant and after a while
	// it could wrap around. But it is useful for debugging.
	id uint64

	// Semaphore to pause/resume the thread atomically.
	sem sem
}

// Goroutine counter, starting at 0 for the main goroutine.
var goroutineID uint64

var mainTask Task

func OnSystemStack() bool {
	runtimePanic("todo: task.OnSystemStack")
	return false
}

// Initialize the main goroutine state. Must be called by the runtime on
// startup, before starting any other goroutines.
func Init() {
	// Sanity check. With ThinLTO, this should be getting optimized away.
	if unsafe.Sizeof(pthread_mutex{}) != tinygo_mutex_size() {
		panic("internal/task: unexpected sizeof(pthread_mutex_t)")
	}
	if unsafe.Alignof(pthread_mutex{}) != tinygo_mutex_align() {
		panic("internal/task: unexpected _Alignof(pthread_mutex_t)")
	}
	if unsafe.Sizeof(sem{}) != tinygo_sem_size() {
		panic("semaphore is an unexpected size!")
	}
	if unsafe.Alignof(sem{}) != tinygo_sem_align() {
		panic("semaphore is an unexpected alignment!")
	}

	mainTask.init()
	tinygo_task_set_current(&mainTask)
}

func (t *Task) init() {
	sem_init(&t.state.sem, 0, 0)
}

// Return the task struct for the current thread.
func Current() *Task {
	t := (*Task)(tinygo_task_current())
	if t == nil {
		runtimePanic("unknown current task")
	}
	return t
}

// Pause pauses the current task, until it is resumed by another task.
// It is possible that another task has called Resume() on the task before it
// hits Pause(), in which case the task won't be paused but continues
// immediately.
func Pause() {
	// Wait until resumed
	t := Current()
	if verbose {
		println("*** pause:  ", t.state.id)
	}
	if sem_wait(&t.state.sem) != 0 {
		runtimePanic("sem_wait error!")
	}
}

// Resume the given task.
// It is legal to resume a task before it gets paused, it means that the next
// call to Pause() won't pause but will continue immediately. This happens in
// practice sometimes in channel operations, where the Resume() might get called
// between the channel unlock and the call to Pause().
func (t *Task) Resume() {
	if verbose {
		println("*** resume: ", t.state.id)
	}
	// Increment the semaphore counter.
	// If the task is currently paused in sem_wait, it will resume.
	// If the task is not yet paused, the next call to sem_wait will continue
	// immediately.
	if sem_post(&t.state.sem) != 0 {
		runtimePanic("sem_post: error!")
	}
}

// Start a new OS thread.
func start(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	t := &Task{}
	t.state.id = atomic.AddUint64(&goroutineID, 1)
	if verbose {
		println("*** start:  ", t.state.id, "from", Current().state.id)
	}
	t.init()
	errCode := tinygo_task_start(fn, args, t, t.state.id)
	if errCode != 0 {
		runtimePanic("could not start thread")
	}
}

type AsyncLock struct {
	// TODO: lock on macOS needs to be initialized with a magic value
	pthread_mutex
}

func (l *pthread_mutex) Lock() {
	errCode := pthread_mutex_lock(l)
	if errCode != 0 {
		runtimePanic("mutex Lock has error code")
	}
}

func (l *pthread_mutex) TryLock() bool {
	return pthread_mutex_trylock(l) == 0
}

func (l *pthread_mutex) Unlock() {
	errCode := pthread_mutex_unlock(l)
	if errCode != 0 {
		runtimePanic("mutex Unlock has error code")
	}
}

//go:linkname runtimePanic runtime.runtimePanic
func runtimePanic(msg string)

// Using //go:linkname instead of //export so that we don't tell the compiler
// that the 't' parameter won't escape (because it will).
//
//go:linkname tinygo_task_set_current tinygo_task_set_current
func tinygo_task_set_current(t *Task)

// Here same as for tinygo_task_set_current.
//
//go:linkname tinygo_task_start tinygo_task_start
func tinygo_task_start(fn uintptr, args unsafe.Pointer, t *Task, id uint64) int32

//export tinygo_task_current
func tinygo_task_current() unsafe.Pointer

//export tinygo_mutex_size
func tinygo_mutex_size() uintptr

//export tinygo_mutex_align
func tinygo_mutex_align() uintptr

//export pthread_mutex_lock
func pthread_mutex_lock(*pthread_mutex) int32

//export pthread_mutex_trylock
func pthread_mutex_trylock(*pthread_mutex) int32

//export pthread_mutex_unlock
func pthread_mutex_unlock(*pthread_mutex) int32

//export sem_init
func sem_init(s *sem, pshared int32, value uint32) int32

//export sem_wait
func sem_wait(*sem) int32

//export sem_post
func sem_post(*sem) int32

//export tinygo_sem_size
func tinygo_sem_size() uintptr

//export tinygo_sem_align
func tinygo_sem_align() uintptr
