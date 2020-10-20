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
)

// allow callers to register to recieve event after any status change occurs
type TaskStatusChangeEventHandler func([]GoLogTaskStatus)

func Attach(handler TaskStatusChangeEventHandler) int {
	statusHandlers = append(statusHandlers, handler)
	return len(statusHandlers) - 1
}

func Detach(handle int) {
	statusHandlers[handle] = nil
}

func publishTaskStatusChangeChange() {
	for _, h := range statusHandlers {
		h(statuses)
	}
}

func setTaskStatus(t GoLogTask, s GoLogTaskStatus) {
	mutexTaskStatuses.Lock()
	defer mutexTaskStatuses.Unlock()

	if t < TaskLast {
		statuses[t] = s
		publishTaskStatusChangeChange()
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
