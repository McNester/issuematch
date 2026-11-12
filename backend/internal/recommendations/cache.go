package recommendations

import (
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const cacheLifetime = 5 * time.Minute

type cacheEntry struct {
	response Response
	limit    int
	expires  time.Time
}

type cache struct {
	mutex   sync.Mutex
	entries map[bson.ObjectID]cacheEntry
}

func (c *cache) get(userID bson.ObjectID, strategy string, limit int) (Response, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	entry, found := c.entries[userID]
	// An entry made for another strategy or limit is not the answer to this request.
	if !found || time.Now().After(entry.expires) || entry.response.Strategy != strategy || entry.limit != limit {
		return Response{}, false
	}
	return entry.response, true
}

func (c *cache) put(userID bson.ObjectID, limit int, response Response) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries[userID] = cacheEntry{response, limit, time.Now().Add(cacheLifetime)}
}

func (c *cache) forget(userID bson.ObjectID) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.entries, userID)
}

func (c *cache) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	clear(c.entries)
}
