package scaleway

import (
	"sync"
)

var locks = sync.Map{}

func lockLocalizedID(id string) (unlock func()) {
	lock, _ := locks.LoadOrStore(id, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	l.Debugf("LOCKING %s", id)
	return func() {
		l.Debugf("UNLOCKING %s", id)
		mutex.Unlock()
	}
}
