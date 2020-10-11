package tasks

import (
	"sync"
	"time"
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

// scheduleRecurring creates a recurring task, returns quit channel
func scheduleRecurring(what func(), delay time.Duration) chan bool {
	ticker := time.NewTicker(delay)
	quit := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				what()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return quit
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

	// define tasks that run every minute
	tasksOneMinute := []func(){
		SourceFiles,
		QSLClublog,
		QSLEqsl,
		QSLQrz,
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
		q := scheduleRecurring(fn, time.Duration(60)*time.Second)

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

// Shutdown stops all background tasks
func Shutdown() {
	// stop startup if its still going
	close(quitStartup)

	mutexQuitChannels.Lock()
	defer mutexQuitChannels.Unlock()

	// stop tasks
	for _, q := range quitChannels {
		close(q)
	}
	quitChannels = make([]chan bool, 0)

	//	concurrent shutdown tasks
	var wg sync.WaitGroup

	// final tasks before shutting down
	// send all reaming QSOs to QSL
	// & cleanup
	tasks := []func(){
		QSLLotwFinal,
		QSLClublogFinal,
		QSLEqslFinal,
		QSLQrzFinal,
		Cleanup,
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
