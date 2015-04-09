package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/quipo/statsd"
	"github.com/vektra/templar"
)

var fDebug = flag.Bool("debug", false, "show debugging info")
var fStatsd = flag.String("statsd", "", "address to sends statsd stats")
var fExpire = flag.Duration("expire", 5*time.Minute, "how long to use cached values")

var fMemcache = flag.String("memcache", "", "memcache servers to use for caching")
var fRedis = flag.String("redis", "", "redis server to use for caching")
var fRedisPassword = flag.String("redis-password", "", "password to redis server")
var fGroupCacheThisPeer = flag.String("groupcache-this-peer", "", "groupcache peer url to use for this peer")
var fGroupCacheOtherPeers = flag.String("groupcache-other-peers", "", "groupcache peer url set to use for caching (comma separated)")

var fListen = flag.String("listen", "0.0.0.0:9224", "address to listen on")

func main() {
	flag.Parse()

	var stats templar.MultiStats

	if *fDebug {
		stats = append(stats, &templar.DebugStats{})
	}

	if *fStatsd != "" {
		client := statsd.NewStatsdClient(*fStatsd, "")
		err := client.CreateSocket()
		if err != nil {
			panic(err)
		}

		stats = append(stats, templar.NewStatsdOutput(client))
	}

	categorizer := templar.NewCategorizer()

	transport := templar.NewHTTPTransport()

	var cache templar.CacheBackend

	switch {
	case *fMemcache != "":
		cache = templar.NewMemcacheCache(strings.Split(*fMemcache, ":"), *fExpire)
	case *fRedis != "":
		cache = templar.NewRedisCache(*fRedis, *fRedisPassword, *fExpire)
	case *fGroupCacheThisPeer != "" && *fGroupCacheOtherPeers != "":
		cache = templar.NewGroupCacheCache(*fGroupCacheThisPeer, *fGroupCacheOtherPeers, *fExpire, transport)
	case *fGroupCacheThisPeer != "" && *fGroupCacheOtherPeers == "":
		panic(errors.New("templar: passed --groupcache-this-peer without passing --groupcache-other-peers. You have to set both of them to use the group cache backend"))
	case *fGroupCacheThisPeer == "" && *fGroupCacheOtherPeers != "":
		panic(errors.New("templar: passed --groupcache-other-peers without passing --groupcache-this-peer. You have to set both of them to use the group cache backend"))
	default:
		cache = templar.NewMemoryCache(*fExpire)
	}

	fallback := templar.NewFallbackCacher(cache, transport, categorizer)
	eager := templar.NewEagerCacher(cache, fallback, categorizer)

	upstream := templar.NewUpstream(eager, stats)
	collapse := templar.NewCollapser(upstream, categorizer)

	proxy := templar.NewProxy(collapse, stats)

	log.Fatal(http.ListenAndServe(*fListen, proxy))
}
