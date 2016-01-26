package ticketswap

import (
	"fmt"
	"net/http"
	"sort"
	"time"
)

// Task defines a task that periodically fetches, creates and sends ads to the `AdsCh` channel.
type Task struct {
	URL    string
	AdsCh  chan Advertisements
	StopCh chan struct{}
}

// NewBot creates a Task via specified `url`
func NewTask(url string) *Task {
	return &Task{url, make(chan Advertisements), make(chan struct{})}
}

// Start periodically fetches ads from the remote host and sends them to the ads channel.
func (task *Task) Start(timeout time.Duration) {
	go func() {
		for {
			select {
			case <-task.StopCh:
				close(task.AdsCh)
				fmt.Println("stop task...")
				return
			case <-time.After(timeout * time.Second):
				ads, err := task.fetchAds()
				if err == nil && len(ads) != 0 {
					task.AdsCh <- ads
				}
				fmt.Printf("found %d ads.\n", len(ads))
			}
		}
	}()
}

// Stop stops periodically updates, that are started by `Start`
func (task *Task) Stop() {
	task.StopCh <- struct{}{}
}

// fetchAds gets available ads from the remote host
func (task *Task) fetchAds() (Advertisements, error) {
	// 1. fetch html
	response, _ := http.Get(task.URL)

	// 2. parse html
	defer response.Body.Close()
	ads, err := NewAdvertisements(response.Body)
	if err != nil {
		return nil, err
	}
	sort.Sort(ads)
	return ads, nil
}
