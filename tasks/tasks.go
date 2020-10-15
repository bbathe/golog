package tasks

import (
	"sync"
	"time"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/util"
)

var (
	mutexQuitChannels sync.Mutex
	quitChannels      []chan bool
	quitStartup       chan bool
)

// max returns the maximum of integers a or b
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Start starts all background tasks
func Start() {
	mutexQuitChannels.Lock()
	defer mutexQuitChannels.Unlock()

	if len(quitChannels) > 0 {
		// already did this
		return
	}

	// our quit channel
	quitStartup = make(chan bool)

	// fire off thread to collect spots from HamAlert, if configured
	if config.ClusterServices.HamAlert.Validate() == nil {
		StartHamAlerts()
	}

	// define tasks that run every minute
	tasksOneMinute := []func(){
		SourceFiles,
	}

	// add services that are configured
	if config.LogbookServices.ClubLog.Validate() == nil {
		tasksOneMinute = append(tasksOneMinute, QSLClublog)
	}
	if config.LogbookServices.EQSL.Validate() == nil {
		tasksOneMinute = append(tasksOneMinute, QSLEqsl)
	}
	if config.LogbookServices.QRZ.Validate() == nil {
		tasksOneMinute = append(tasksOneMinute, QSLQrz)
	}

	// create quit channels
	quitChannels = make([]chan bool, 0, len(tasksOneMinute))

	// since we have a bunch of tasks to start and some sympathy toward our host
	// we are going to stagger the starting of the tasks
	// calculate spacing between starting tasks
	space := max(60/len(tasksOneMinute), 1)

	// schedule the tasks
	for _, fn := range tasksOneMinute {
		fn := fn

		// create recurring task
		q := util.ScheduleRecurring(fn, time.Duration(60)*time.Second)

		// and keep quit channel
		quitChannels = append(quitChannels, q)

		// pause while checking if we being signalled for shutdown
		for i := 0; i < space; i++ {
			select {
			case <-quitStartup:
				return
			default:
				// pause between starts
				time.Sleep(time.Duration(1) * time.Second)
			}
		}
	}
}

// Pause stops all background tasks, no cleanup
func Pause() {
	// stop startup if its still going
	close(quitStartup)

	mutexQuitChannels.Lock()
	defer mutexQuitChannels.Unlock()

	// stop collecting HamAlert spots
	StopHamAlerts()

	// stop tasks
	for _, q := range quitChannels {
		close(q)
	}
	quitChannels = make([]chan bool, 0)
}

// Shutdown stops all background tasks, and runs cleanup tasks
func Shutdown() {
	Pause()

	//	concurrent shutdown tasks
	var wg sync.WaitGroup

	// final tasks before shutting down
	// send all remaining QSOs to QSL
	// & cleanup
	tasks := []func(){
		Cleanup,
	}

	// add services that are configured
	if config.LogbookServices.TQSL.Validate() == nil {
		tasks = append(tasks, QSLLotwFinal)
	}
	if config.LogbookServices.ClubLog.Validate() == nil {
		tasks = append(tasks, QSLClublogFinal)
	}
	if config.LogbookServices.EQSL.Validate() == nil {
		tasks = append(tasks, QSLEqslFinal)
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
