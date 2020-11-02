package tasks

type GoLogTask int

const (
	TaskSourceFiles GoLogTask = iota
	TaskQSLLoTW
	TaskQSLEQSL
	TaskQSLQRZ
	TaskQSLClubLog
	TaskHamAlert

	TaskLast // so we can get the number of tasks defined
)

type GoLogTaskStatus int

const (
	TaskStatusOK GoLogTaskStatus = iota
	TaskStatusFailed
	TaskStatusNotRunning
)

// allow callers to register to recieve event after any status change occurs
type TaskStatusChangeEventHandler func([]GoLogTaskStatus)

func Attach(handler TaskStatusChangeEventHandler) int {
	statusHandlers = append(statusHandlers, handler)
	h := len(statusHandlers) - 1

	return h
}

func Detach(handle int) {
	statusHandlers[handle] = nil
}

func publishTaskStatusChange() {
	for _, h := range statusHandlers {
		if h != nil {
			h(statuses[:])
		}
	}
}

func setTaskStatus(t GoLogTask, s GoLogTaskStatus) {
	mutexTaskStatuses.Lock()
	defer mutexTaskStatuses.Unlock()

	if t < TaskLast {
		statuses[t] = s
		publishTaskStatusChange()
	}
}

func taskWrapper(task GoLogTask, function func() error) func() {
	t := task
	fn := function

	return func() {
		err := fn()
		if err != nil {
			setTaskStatus(t, TaskStatusFailed)
		} else {
			setTaskStatus(t, TaskStatusOK)
		}
	}
}
