package tasks

import (
	"context"
	"sync"
	"time"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/util"
)

var (
	mutexQuitChannels sync.Mutex
	quitChannels      []chan bool
	allowStart        = true
	running           bool
	startCancel       context.CancelFunc
	startDone         chan struct{}

	mutexTaskStatuses sync.Mutex
	statuses          [int(TaskLast)]GoLogTaskStatus
	statusHandlers    []TaskStatusChangeEventHandler
)

// Start starts all background tasks
func Start() {
	mutexQuitChannels.Lock()
	if running || !allowStart {
		mutexQuitChannels.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	running = true
	startCancel = cancel
	startDone = done
	mutexQuitChannels.Unlock()

	defer close(done)

	// set status of all tasks to not running
	for t := 0; t < int(TaskLast); t++ {
		setTaskStatus(GoLogTask(t), TaskStatusNotRunning)
	}

	// define tasks that run every minute
	tasksOneMinute := []func(){
		taskWrapper(TaskSourceFiles, SourceFiles),
	}

	// add services that are configured
	if config.LogbookServices.TQSL.Validate() == nil {
		tasksOneMinute = append(tasksOneMinute, taskWrapper(TaskQSLLoTW, QSLLotw))
	}
	if config.LogbookServices.QRZ.Validate() == nil {
		tasksOneMinute = append(tasksOneMinute, taskWrapper(TaskQSLQRZ, QSLQrz))
	}
	if config.LogbookServices.ClubLog.Validate() == nil {
		tasksOneMinute = append(tasksOneMinute, taskWrapper(TaskQSLClubLog, QSLClublog))
	}

	// since we have a bunch of tasks to start and some sympathy toward our host
	// we are going to stagger the starting of the tasks
	// calculate spacing between starting tasks
	space := 60 / len(tasksOneMinute)
	if space < 1 {
		space = 1
	}

	// fire off long-lived thread to collect spots from HamAlert, if configured
	// statues are updated directly in the HamAlerts module
	if ctx.Err() == nil && config.ClusterServices.HamAlert.Validate() == nil {
		StartHamAlerts()
	}

	// schedule the tasks
	for _, fn := range tasksOneMinute {
		if !scheduleTask(ctx, fn) {
			return
		}
		if !waitTaskGap(ctx, space) {
			return
		}
	}
}

// scheduleTask starts one recurring task unless startup has been cancelled.
// The quit channel is closed and not stored when startup is already cancelled,
// so Pause does not close it a second time.
func scheduleTask(ctx context.Context, fn func()) bool {
	if ctx.Err() != nil {
		return false
	}

	q := util.ScheduleRecurring(fn, 60*time.Second)

	mutexQuitChannels.Lock()
	defer mutexQuitChannels.Unlock()

	if ctx.Err() != nil {
		close(q)
		return false
	}

	quitChannels = append(quitChannels, q)
	return true
}

// waitTaskGap waits between task starts. Cancel returns immediately.
func waitTaskGap(ctx context.Context, space int) bool {
	timer := time.NewTimer(time.Duration(space) * time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// Pause stops all background tasks, no cleanup
func Pause() {
	mutexQuitChannels.Lock()
	cancel := startCancel
	done := startDone
	wasRunning := running
	mutexQuitChannels.Unlock()

	// Cancel outside the lock. Start waits on this context without holding the mutex.
	if cancel != nil {
		cancel()
	}
	if wasRunning && done != nil {
		<-done
	}

	StopHamAlerts()

	mutexQuitChannels.Lock()
	defer mutexQuitChannels.Unlock()

	for _, q := range quitChannels {
		close(q)
	}
	quitChannels = nil
	running = false
	startCancel = nil
	startDone = nil

	// set status of all tasks to not running
	for t := 0; t < int(TaskLast); t++ {
		setTaskStatus(GoLogTask(t), TaskStatusNotRunning)
	}
}

// Shutdown stops all background tasks, and runs cleanup tasks
func Shutdown() {
	mutexQuitChannels.Lock()
	allowStart = false
	mutexQuitChannels.Unlock()

	Pause()

	//	concurrent shutdown tasks
	var wg sync.WaitGroup

	// final tasks before shutting down
	// backup qsos & send all remaining QSOs to logbook services
	tasks := []func(){
		BackupQSOs,
	}

	// add services that are configured
	if config.LogbookServices.TQSL.Validate() == nil {
		tasks = append(tasks, QSLLotwFinal)
	}
	if config.LogbookServices.ClubLog.Validate() == nil {
		tasks = append(tasks, QSLClublogFinal)
	}
	if config.LogbookServices.QRZ.Validate() == nil {
		tasks = append(tasks, QSLQrzFinal)
	}

	// spinup all the shutdown tasks
	for _, t := range tasks {
		wg.Add(1)

		go func(fn func()) {
			defer wg.Done()

			fn()
		}(t)
	}

	// wait for them to complete
	wg.Wait()
}
