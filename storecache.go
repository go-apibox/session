package session

type cookieStoreCache struct {
	stores map[string]*CookieStore
}

func (c *cookieStoreCache) has(keyPairsFile string) bool {
	_, exists := c.stores[keyPairsFile]
	return exists
}

func (c *cookieStoreCache) get(keyPairsFile string) *CookieStore {
	store, exists := c.stores[keyPairsFile]
	if exists {
		return store
	} else {
		return nil
	}
}

func (c *cookieStoreCache) set(keyPairsFile string, store *CookieStore) {
	c.stores[keyPairsFile] = store
}
