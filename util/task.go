package util

import "time"

// ScheduleRecurring creates a recurring task, returns quit channel
func ScheduleRecurring(what func(), delay time.Duration) chan bool {
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
