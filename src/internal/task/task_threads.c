//go:build none

#include <pthread.h>
#include <stdint.h>
#include <stdio.h>
#include <semaphore.h>

// Pointer to the current task.Task structure.
// Ideally the entire task.Task structure would be a thread-local variable but
// this also works.
static __thread void *current_task;

struct state_pass {
    void *(*start)(void*);
    void *args;
    void *task;
    sem_t startlock;
};

// Helper to start a goroutine while also storing the 'task' structure.
static void* start_wrapper(void *arg) {
    struct state_pass *state = arg;
    void *(*start)(void*) = state->start;
    void *args = state->args;
    current_task = state->task;
    sem_post(&state->startlock);
    start(args);
    return NULL;
};

// Start a new goroutine in an OS thread.
int tinygo_task_start(uintptr_t fn, void *args, void *task, uint64_t id, void *context) {
    struct state_pass state = {
        .start     = (void*)fn,
        .args      = args,
        .task      = task,
    };
    sem_init(&state.startlock, 0, 0);
    pthread_t thread;
    int result = pthread_create(&thread, NULL, &start_wrapper, &state);

    // Wait until the thread has been crated and read all state_pass variables.
    sem_wait(&state.startlock);

    return result;
}

// Return the current task (for task.Current()).
void* tinygo_task_current(void) {
    return current_task;
}

// Set the current task at startup.
void tinygo_task_set_current(void *task, void *context) {
    current_task = task;
}

uintptr_t tinygo_mutex_size(void) {
    return sizeof(pthread_mutex_t);
}

uintptr_t tinygo_mutex_align(void) {
    return _Alignof(pthread_mutex_t);
}

uintptr_t tinygo_sem_size(void) {
    return sizeof(sem_t);
}

uintptr_t tinygo_sem_align(void) {
    return _Alignof(sem_t);
}
