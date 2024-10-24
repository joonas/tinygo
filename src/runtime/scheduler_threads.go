//go:build scheduler.threads

package runtime

import "internal/task"

const hasScheduler = false // not using the cooperative scheduler

// Because we just use OS threads, we don't need to do anything special here. We
// can just initialize everything and run main.main on the main thread.
func run() {
	initHeap()
	task.Init()
	initAll()
	callMain()
}

// Pause the current task for a given time.
//
//go:linkname sleep time.Sleep
func sleep(duration int64) {
	if duration <= 0 {
		return
	}

	sleepTicks(nanosecondsToTicks(duration))
}

func deadlock() {
	// TODO: exit the thread via pthread_exit.
	task.Pause()
}

func resumeTask(t *task.Task) {
	t.Resume()
}

func Gosched() {
	// Each goroutine runs in a thread, so there's not much we can do here.
	// There is sched_yield but it's only really intended for realtime
	// operation, so is probably best not to use.
}

func addTimer(tim *timerNode) {
	// TODO: I think we can implement this by having a single goroutine (thread)
	// process all timers.
	runtimePanic("todo: addTimer")
}

func removeTimer(tim *timer) bool {
	runtimePanic("todo: removeTimer")
	return false
}

func runqueueForGC() *task.Queue {
	// There is only a runqueue when using the cooperative scheduler.
	return nil
}
