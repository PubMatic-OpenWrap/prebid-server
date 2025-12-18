package gocache

// LockAndLoad calls DB only once for same requests
func (c *cache) LockAndLoad(key string, dbFunc func() error) (err error) {
	waitCh := make(chan struct{})
	lockCh, present := c.LoadOrStore(key, waitCh)
	if !present {
		// fetch db data and save in cache
		err = dbFunc()
		// delete and let other requests take the lock (ideally only 1 per hour per pod)
		c.Delete(key)
		// unblock waiting requests
		close(waitCh)
	}

	// requests that did not get lock will wait here until the one that reterives the data closes the channel
	<-lockCh.(chan struct{})
	return
}
