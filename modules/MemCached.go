package modules

import (
	"github.com/bradfitz/gomemcache/memcache"
)

func InitiateMemCachedClient() *memcache.Client {
	memcachedHost := MapConfig["memcachedHost"]
	memcachedPort := MapConfig["memcachedPort"]

	memCachedClient := memcache.New("127.0.0.1:11211")
	DoLog("INFO", "", "MemCached", "InitiateMemCachedClient",
		"Initiate memcached to host "+memcachedHost+", port "+memcachedPort, true, nil)

	return memCachedClient
}

func SetMemCached(memcachedClient *memcache.Client, memKey string, memVal string) bool {
	err := memcachedClient.Set(&memcache.Item{Key: memKey, Value: []byte(memVal)})
	if err != nil {
		return false
	} else {
		return true
	}
}

func GetMemCached(memcachedClient *memcache.Client, memKey string) string {
	val, err := memcachedClient.Get(memKey)

	if err != nil {
		return ""
	} else {
		return string(val.Value)
	}
}

func DelMemCached(memcachedClient *memcache.Client, memKey string) bool {
	err := memcachedClient.Delete(memKey)
	if err != nil {
		return false
	} else {
		return true
	}
}
