package query_cache

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/error_helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/semaphore"
	"google.golang.org/protobuf/proto"
)

type CacheData interface {
	proto.Message
	*sdkproto.QueryResult | *sdkproto.IndexBucket
}

const (
	// cache has a default hard TTL limit of 24 hours
	DefaultMaxTtl = 24 * time.Hour
	rowBufferSize = 1000
)

type QueryCache struct {
	Stats      *CacheStats
	pluginName string
	// map of connection name to plugin schema
	PluginSchemaMap map[string]*grpc.PluginSchema
	// map of pending cache transfers, keyed by index bucket key
	pendingData     map[string]*pendingIndexBucket
	pendingDataLock sync.RWMutex

	cache *cache.Cache[[]byte]
	// map of ongoing set requests, keyed by callId
	setRequests       map[string]*setRequest
	setRequestMapLock sync.RWMutex
	Enabled           bool
}

func NewQueryCache(pluginName string, pluginSchemaMap map[string]*grpc.PluginSchema, opts *QueryCacheOptions) (*QueryCache, error) {
	queryCache := &QueryCache{
		Stats:           &CacheStats{},
		pluginName:      pluginName,
		PluginSchemaMap: pluginSchemaMap,
		pendingData:     make(map[string]*pendingIndexBucket),
		setRequests:     make(map[string]*setRequest),
		Enabled:         opts.Enabled,
	}
	if err := queryCache.createCache(opts.MaxSizeMb, opts.Ttl); err != nil {
		return nil, err
	}
	log.Printf("[INFO] query cache created")
	return queryCache, nil
}

func (c *QueryCache) createCache(maxCacheStorageMb int, maxTtl time.Duration) error {
	cacheStore, err := c.createCacheStore(maxCacheStorageMb, maxTtl)
	if err != nil {
		return err
	}
	c.cache = cache.New[[]byte](cacheStore)
	return nil
}

func (c *QueryCache) createCacheStore(maxCacheStorageMb int, maxTtl time.Duration) (store.StoreInterface, error) {
	config := bigcache.DefaultConfig(maxTtl)
	// ensure each shard is at least 5Mb
	config.Shards = 1024
	for maxCacheStorageMb/config.Shards < 5 {
		config.Shards = config.Shards / 2
		if config.Shards == 2 {
			break
		}
	}
	config.HardMaxCacheSize = maxCacheStorageMb
	log.Printf("[TRACE] createCacheStore for plugin '%s' setting max size to %dMb, Shards: %d, max shard size: %d ", c.pluginName, maxCacheStorageMb, config.Shards, ((maxCacheStorageMb*1024*1024)/config.Shards)/(1024*1024))

	bigcacheClient, _ := bigcache.NewBigCache(config)
	bigcacheStore := store.NewBigcache(bigcacheClient)
	return bigcacheStore, nil
}

func (c *QueryCache) Get(ctx context.Context, req *CacheRequest, streamRowFunc func(row *sdkproto.Row)) error {
	if !c.Enabled {
		return fmt.Errorf("cache disabled")
	}
	cacheHit := false
	ctx, span := telemetry.StartSpan(ctx, "QueryCache.Get (%s)", req.Table)
	defer func() {
		span.SetAttributes(attribute.Bool("cache-hit", cacheHit))
		span.End()
	}()

	// set root result key
	req.resultKeyRoot = c.buildResultKey(req)

	// get the index bucket key for this table and quals
	indexBucketKey := c.buildIndexKey(req.ConnectionName, req.Table)
	log.Printf("[TRACE] QueryCache Get - indexBucketKey %s, quals: %s (%s)", indexBucketKey, req.CallId, grpc.QualMapToLogLine(req.QualMap))

	// do we have a cached result?
	err := c.getCachedQueryResult(ctx, indexBucketKey, req, streamRowFunc)
	if err == nil {
		// only set cache hit if there was no error
		cacheHit = true
		return nil
	}

	// if this is not a cache miss, just return the error
	if !IsCacheMiss(err) {
		return err
	}

	// so we have a cache miss
	// there was no cached result - is there data fetch in progress? If so, subscribe to it
	err = c.findAndSubscribeToPendingRequest(ctx, indexBucketKey, req, streamRowFunc)
	if err == nil {
		log.Printf("[INFO] subscribed to pending request")

		cacheHit = true
	}
	return err
}

func (c *QueryCache) findAndSubscribeToPendingRequest(ctx context.Context, indexBucketKey string, req *CacheRequest, streamRowFunc func(row *sdkproto.Row)) (err error) {
	log.Printf("[INFO] getCachedQueryResult returned CACHE MISS - checking for pending transfers (%s)", req.CallId)

	// try to get pending item within a read lock
	c.pendingDataLock.RLock()
	pendingItem := c.getPendingResultItem(indexBucketKey, req)
	c.pendingDataLock.RUnlock()

	if pendingItem == nil {
		// return a cache miss
		err = CacheMissError{}
	} else {
		// so we have a pending item - subscribe to it
		log.Printf("[INFO] found pending item [%s] - subscribing to its data (%s)", pendingItem.callId, req.CallId)
		err = c.subscribeToPendingRequest(ctx, pendingItem, req, streamRowFunc)
	}

	// success?
	if err == nil {
		return
	}

	// if we fail for any reason, this will be treated as a cache miss - add a pending item
	log.Printf("[INFO] findAndSubscribeToPendingRequest returning error %s - this will be treated as a cache miss, so add pending item (%s)", err.Error(), req.CallId)

	// get a write lock in preparation for adding a pending item
	c.pendingDataLock.Lock()

	// if the error WAS a cache miss error,  before adding a pending result, try again for a pending item
	// this is to allow for the race condition where 2 threads are both making a concurrent cache request
	// - one will create a pending item first
	if IsCacheMiss(err) {
		if pendingItem := c.getPendingResultItem(indexBucketKey, req); pendingItem != nil {
			// unlock before subscribeToPendingRequest
			c.pendingDataLock.Unlock()

			// ok NOW there is a pending item - just subscribe to it
			return c.subscribeToPendingRequest(ctx, pendingItem, req, streamRowFunc)
		}
	}

	// add a pending result so anyone else asking for this data will wait the fetch to complete
	c.addPendingResult(ctx, indexBucketKey, req)

	// unlock the write lock
	c.pendingDataLock.Unlock()

	return err

}

func (c *QueryCache) subscribeToPendingRequest(ctx context.Context, pendingItem *pendingIndexItem, req *CacheRequest, streamRowFunc func(row *sdkproto.Row)) error {
	// get the ongoing set request
	pendingSetRequest := pendingItem.pendingSetRequest

	// create a subscriber
	subscriber := newSetRequestSubscriber(streamRowFunc, req.CallId)

	// now lock the set request
	pendingSetRequest.mut.Lock()
	// subscribe with the wrapped func
	pendingSetRequest.subscribe(subscriber)
	// get keys of all data already written to cache
	prevKeys := pendingSetRequest.getPrevPageResultKeys()
	bufferedRows := pendingSetRequest.getRows()

	// now unlock the set request
	pendingSetRequest.mut.Unlock()

	// stream all data already cached
	log.Printf("[INFO] stream all data already cached")
	for _, pageKey := range prevKeys {
		var cacheResult = &sdkproto.QueryResult{}
		if err := doGet[*sdkproto.QueryResult](ctx, pageKey, c.cache, cacheResult); err != nil {
			return err
		}
		for _, row := range cacheResult.Rows {
			subscriber.streamRowFunc(row)
		}
	}
	// now stream all data currently in the set request buffer
	for _, row := range bufferedRows {
		subscriber.streamRowFunc(row)
	}

	// wait for all rows tro be streamed (or an error)
	err := subscriber.waitUntilDone()
	if err != nil {
		log.Printf("[WARN] set request we are subscribed to failed: %s", err.Error())
	} else {
		log.Printf("[INFO] All rows streamed")
	}
	return err
}

// startSet begins a streaming cache Set operation.
// NOTE: this mutates req
func (c *QueryCache) startSet(_ context.Context, req *CacheRequest) *setRequest {
	// should never happen
	if !c.Enabled {
		return nil
	}

	log.Printf("[TRACE] startSet (%s)", req.CallId)

	// set root result key
	req.resultKeyRoot = c.buildResultKey(req)
	// create rows buffer
	req.rows = make([]*sdkproto.Row, rowBufferSize)

	setRequest := &setRequest{CacheRequest: req}

	// lock the set request map
	c.setRequestMapLock.Lock()
	c.setRequests[req.CallId] = setRequest
	c.setRequestMapLock.Unlock()

	return setRequest
}

func (c *QueryCache) IterateSet(ctx context.Context, row *sdkproto.Row, callId string) error {
	// should never happen
	if !c.Enabled {
		return nil
	}
	// get the ongoing request
	c.setRequestMapLock.RLock()
	req, ok := c.setRequests[callId]
	c.setRequestMapLock.RUnlock()
	if !ok {
		// not expected
		return fmt.Errorf("IterateSet failed - no set request with call id %s", callId)
	}
	// lock access to set request
	req.mut.Lock()
	defer req.mut.Unlock()

	if !ok {
		return fmt.Errorf("IterateSet called for callId %s but there is no in-progress 'set' operation", callId)
	}
	// was there an error in a previous iterate
	if req.err != nil {
		return req.err
	}

	req.rows[req.rowIndex] = row
	req.rowIndex++

	// write this row to all subscribers
	req.streamToSubscribers(row)

	// if we have buffered a page, write to cache
	if req.rowIndex == rowBufferSize {
		// reset index and update page count
		log.Printf("[TRACE] IterateSet written 1 page of %d rows. Page count %d (%s)", rowBufferSize, req.pageCount, req.CallId)
		req.err = c.writePageToCache(ctx, req)
	}

	return nil
}

func (c *QueryCache) EndSet(ctx context.Context, callId string) (err error) {
	// should never happen
	if !c.Enabled {
		return nil
	}

	c.setRequestMapLock.RLock()
	// get the ongoing request
	req, ok := c.setRequests[callId]
	c.setRequestMapLock.RUnlock()
	if !ok {
		log.Printf("[WARN] EndSet called for callId %s but there is no in progress set operation", callId)
		return fmt.Errorf("EndSet called for callId %s but there is no in progress set operation", callId)
	}

	log.Printf("[TRACE] EndSet (%s) table %s root key %s, pages: %d", callId, req.Table, req.resultKeyRoot, req.pageCount)

	// lock the set request
	req.mut.Lock()
	defer req.mut.Unlock()

	defer func() {
		log.Printf("[TRACE] EndSet DEFER (%s) table %s", callId, req.Table)
		if r := recover(); r != nil {
			log.Printf("[WARN] QueryCache EndSet suffered a panic: %v", helpers.ToError(r))
			err = helpers.ToError(r)
		}
		// remove entry from the map
		c.setRequestMapLock.Lock()
		delete(c.setRequests, callId)
		c.setRequestMapLock.Unlock()

		log.Printf("[TRACE] calling pendingItemComplete (%s)", callId)

		// clear the corresponding pending item - we have completed the transfer
		// (we need to do this even if the cache set fails)
		log.Printf("[TRACE] QueryCache EndSet table: %s, marking pending item complete (%s)", req.Table, req.CallId)
		c.pendingItemComplete(req.CacheRequest)

		// mark the request as complete
		req.complete = true
	}()

	// write the remainder to the result cache
	err = c.writePageToCache(ctx, req)
	if err != nil {
		log.Printf("[WARN] QueryCache EndSet - result Set failed: %v", err)
		return err
	} else {
		log.Printf("[TRACE] QueryCache EndSet - result written")
	}

	// now update the index
	// get the index bucket for this table and connection
	indexBucketKey := c.buildIndexKey(req.ConnectionName, req.Table)
	log.Printf("[TRACE] QueryCache EndSet indexBucketKey %s", indexBucketKey)

	indexBucket, err := c.getCachedIndexBucket(ctx, indexBucketKey)
	if err != nil {
		if IsCacheMiss(err) {
			log.Printf("[TRACE] getCachedIndexBucket returned cache miss (%s)", callId)
		} else {
			// if there is an error fetching the index bucket, log it and return
			// we do not want to risk overwriting an existing index bucket
			log.Printf("[WARN] getCachedIndexBucket failed: %v", err)
			return
		}
	}

	indexItem := NewIndexItem(req.CacheRequest)

	if indexBucket == nil {
		// create new index bucket
		indexBucket = newIndexBucket()
	}
	indexBucket.Append(indexItem)
	log.Printf("[TRACE] QueryCache EndSet - Added index item (%p) to bucket (%p), page count %d,  key root %s (%s)", indexItem, indexBucket, req.pageCount, req.resultKeyRoot, callId)

	// write index bucket back to cache
	err = c.cacheSetIndexBucket(ctx, indexBucketKey, indexBucket, req.CacheRequest)
	if err != nil {
		log.Printf("[WARN] cache Set failed for index bucket: %v", err)
	} else {
		log.Printf("[TRACE] QueryCache EndSet - IndexBucket written (%s)", callId)
	}

	// write empty row to all subscribers
	req.streamToSubscribers(nil)

	return err
}

func (c *QueryCache) AbortSet(ctx context.Context, callId string, err error) {
	if !c.Enabled {
		return
	}
	c.setRequestMapLock.Lock()
	// get the ongoing request
	req, ok := c.setRequests[callId]
	// remove set request item
	delete(c.setRequests, callId)
	c.setRequestMapLock.Unlock()
	if !ok {
		return
	}
	// tell request to send error to all it's subscribers
	req.abort(err)

	// clear the corresponding pending item
	c.pendingItemComplete(req.CacheRequest)

	log.Printf("[TRACE] QueryCache AbortSet - deleting %d pages from the cache", req.pageCount)
	// remove all pages that have already been written
	for i := int64(0); i < req.pageCount; i++ {
		pageKey := getPageKey(req.resultKeyRoot, i)
		c.cache.Delete(ctx, pageKey)
	}
}

// ClearForConnection removes all cache entries for the given connection
func (c *QueryCache) ClearForConnection(ctx context.Context, connectionName string) {
	if !c.Enabled {
		return
	}
	c.cache.Invalidate(ctx, store.WithInvalidateTags([]string{connectionName}))
}

// write a page of rows to the cache
func (c *QueryCache) writePageToCache(ctx context.Context, req *setRequest) error {
	// ask the request for it's currently buffered rows
	rows := req.getRows()
	// reset the row buffer index and increment the page count
	// (BEFORE building pageKey)
	req.pageCount++
	req.rowIndex = 0

	// build a cache key for this page
	pageKey := req.getPageResultKey()

	log.Printf("[TRACE] QueryCache writePageToCache: %d rows, pageCount %d, page key %s", len(rows), req.pageCount, pageKey)
	// write to cache - construct result key
	result := &sdkproto.QueryResult{Rows: rows}

	// put connection name in tags
	tags := []string{req.ConnectionName}
	err := doSet(ctx, pageKey, result, req.ttl(), c.cache, tags)
	if err != nil {
		log.Printf("[WARN] writePageToCache cache Set failed: %v", err)
	} else {
		log.Printf("[TRACE] writePageToCache Set - result written")
	}

	return err
}

func (c *QueryCache) getCachedIndexBucket(ctx context.Context, key string) (*IndexBucket, error) {
	var indexBucket = &sdkproto.IndexBucket{}
	if err := doGet(ctx, key, c.cache, indexBucket); err != nil {
		if IsCacheMiss(err) {
			c.Stats.Misses++
			log.Printf("[TRACE] getCachedIndexBucket - no item retrieved for cache key %s", key)
		} else {
			log.Printf("[WARN] cacheGetResult Get failed %v", err)
		}
		return nil, err
	}

	log.Printf("[TRACE] getCachedIndexBucket cache hit ")
	var res = IndexBucketfromProto(indexBucket)
	return res, nil
}

func (c *QueryCache) getCachedQueryResult(ctx context.Context, indexBucketKey string, req *CacheRequest, streamRowFunc func(row *sdkproto.Row)) error {
	log.Printf("[TRACE] QueryCache getCachedQueryResult - table %s, connectionName %s", req.Table, req.ConnectionName)
	keyColumns := c.getKeyColumnsForTable(req.Table, req.ConnectionName)

	log.Printf("[TRACE] index bucket key: %s ttlSeconds %d limit: %d\n", indexBucketKey, req.TtlSeconds, req.Limit)
	indexBucket, err := c.getCachedIndexBucket(ctx, indexBucketKey)
	if err != nil {
		return err
	}

	// now check whether we have a cache entry that covers the required quals and columns - check the index
	indexItem := indexBucket.Get(req, keyColumns)
	if indexItem == nil {
		limitString := "NONE"
		if req.Limit != -1 {
			limitString = fmt.Sprintf("%d", req.Limit)
		}
		c.Stats.Misses++
		log.Printf("[TRACE] getCachedQueryResult - no cached data covers columns %v, limit %s\n", req.Columns, limitString)
		return new(CacheMissError)
	}

	return c.getCachedQueryResultFromIndexItem(ctx, indexItem, streamRowFunc)
}

func (c *QueryCache) getCachedQueryResultFromIndexItem(ctx context.Context, indexItem *IndexItem, streamRowFunc func(row *sdkproto.Row)) error {
	// so we have a cache index, retrieve the item
	log.Printf("[TRACE] got an index item - try to retrieve rows from cache")

	cacheHit := true
	var errors []error
	errorChan := make(chan (error), indexItem.PageCount)
	var wg sync.WaitGroup
	const maxReadThreads = 5
	var maxReadSem = semaphore.NewWeighted(maxReadThreads)
	log.Printf("[TRACE] %d pages", indexItem.PageCount)

	// define streaming function
	streamRows := func(cacheResult *sdkproto.QueryResult) {
		for _, r := range cacheResult.Rows {
			// check for context cancellation
			if error_helpers.IsContextCancelledError(ctx.Err()) {
				log.Printf("[INFO] getCachedQueryResult context cancelled - returning")
				return
			}
			streamRowFunc(r)
		}
	}
	// ok so we have an index item - we now stream
	// ensure the first page exists (evictions start with oldest item so if first page exists, they all exist)
	pageIdx := int64(0)
	pageKey := getPageKey(indexItem.Key, pageIdx)
	var cacheResult = &sdkproto.QueryResult{}
	if err := doGet[*sdkproto.QueryResult](ctx, pageKey, c.cache, cacheResult); err != nil {
		return err
	}
	// ok it's there, stream rows
	streamRows(cacheResult)
	// update page index
	pageIdx++

	// now fetch the rest (if any), in parallel maxReadThreads at a time
	for ; pageIdx < indexItem.PageCount; pageIdx++ {
		maxReadSem.Acquire(ctx, 1)
		wg.Add(1)
		// construct the page key, _using the index item key as the root_
		p := getPageKey(indexItem.Key, pageIdx)

		go func(pageKey string) {
			defer wg.Done()
			defer maxReadSem.Release(1)

			log.Printf("[TRACE] fetching key: %s", pageKey)
			var cacheResult = &sdkproto.QueryResult{}
			if err := doGet[*sdkproto.QueryResult](ctx, pageKey, c.cache, cacheResult); err != nil {
				if IsCacheMiss(err) {
					// This is not expected
					log.Printf("[WARN] getCachedQueryResult - no item retrieved for cache key %s", pageKey)
				} else {
					log.Printf("[WARN] cacheGetResult Get failed %v", err)
				}
				errorChan <- err
				return
			}

			log.Printf("[TRACE] got result: %d rows", len(cacheResult.Rows))

			streamRows(cacheResult)
		}(p)
	}
	doneChan := make(chan bool)
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	for {
		select {
		case err := <-errorChan:
			log.Printf("[WARN] getCachedQueryResult received error: %s", err.Error())
			if IsCacheMiss(err) {
				cacheHit = false
			} else {
				errors = append(errors, err)
			}
		case <-doneChan:
			// any real errors return them
			if len(errors) > 0 {
				return helpers.CombineErrors(errors...)
			}
			if cacheHit {
				// this was a hit - return
				c.Stats.Hits++
				return nil
			} else {
				c.Stats.Misses++
				return CacheMissError{}
			}
		}
	}
}

func (c *QueryCache) buildIndexKey(connectionName, table string) string {
	str := c.sanitiseKey(fmt.Sprintf("index__%s_%s",
		connectionName,
		table))
	return str
}

// build a result key, using connection, table, quals, columns and limit
func (c *QueryCache) buildResultKey(req *CacheRequest) string {
	qualString := ""
	if len(req.QualMap) > 0 {
		qualString = fmt.Sprintf("_%s", c.formatQualMapForKey(req.QualMap))
	}
	str := c.sanitiseKey(fmt.Sprintf("%s_%s%s_%s_%d",
		req.ConnectionName,
		req.Table,
		qualString,
		strings.Join(req.Columns, ","),
		req.Limit))
	return str
}

func (c *QueryCache) formatQualMapForKey(qualMap map[string]*sdkproto.Quals) string {
	if len(qualMap) == 0 {
		return ""
	}

	var strs = make([]string, len(qualMap))
	// first build list of keys, then sort them
	keys := make([]string, len(qualMap))
	idx := 0
	for key := range qualMap {
		keys[idx] = key
		idx++
	}
	sort.Strings(keys)
	log.Printf("[TRACE] formatQualMapForKey sorted keys %v\n", keys)

	// now construct cache key from ordered quals
	for i, key := range keys {
		for _, q := range qualMap[key].Quals {
			strs = append(strs, fmt.Sprintf("%s-%s-%v", q.FieldName, q.GetStringValue(), grpc.GetQualValue(q.Value)))
		}
		strs[i] = strings.Join(strs, "-")
	}
	return strings.Join(strs, "-")
}

// return a map of key column for the given table
func (c *QueryCache) getKeyColumnsForTable(table string, connectionName string) map[string]*sdkproto.KeyColumn {
	res := make(map[string]*sdkproto.KeyColumn)
	schema, ok := c.PluginSchemaMap[connectionName]
	if !ok {
		return res
	}
	// build a list of all key columns
	tableSchema, ok := schema.Schema[table]
	if ok {
		for _, k := range append(tableSchema.ListCallKeyColumnList, tableSchema.GetCallKeyColumnList...) {
			res[k.Name] = k
		}
	} else {
		log.Printf("[TRACE] getKeyColumnsForTable found no schema for table '%s'", table)
	}
	return res
}

func (c *QueryCache) sanitiseKey(str string) string {
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	return str
}

// write index bucket back to cache
func (c *QueryCache) cacheSetIndexBucket(ctx context.Context, indexBucketKey string, indexBucket *IndexBucket, req *CacheRequest) error {
	log.Printf("[TRACE] cacheSetIndexBucket %s", indexBucketKey)

	// put connection name in tags
	tags := []string{req.ConnectionName}
	return doSet(ctx, indexBucketKey, indexBucket.AsProto(), req.ttl(), c.cache, tags)
}

func doGet[T CacheData](ctx context.Context, key string, cache *cache.Cache[[]byte], target T) error {
	// get the bytes from the cache
	getRes, err := cache.Get(ctx, key)
	if err != nil {
		if IsCacheMiss(err) {
			log.Printf("[TRACE] doGet cache miss ")
		} else {
			log.Printf("[WARN] cache.Get returned error %s", err.Error())
		}
		//  return the error
		return err
	}

	// unmarshall into the correct type
	err = proto.Unmarshal(getRes, target)
	if err != nil {
		log.Printf("[WARN] error unmarshalling result: %s", err.Error())
		return err
	}

	return nil
}

func doSet[T CacheData](ctx context.Context, key string, value T, ttl time.Duration, cache *cache.Cache[[]byte], tags []string) error {
	bytes, err := proto.Marshal(value)
	if err != nil {
		log.Printf("[WARN] doSet - marshal failed: %v", err)
		return err
	}

	err = cache.Set(ctx,
		key,
		bytes,
		store.WithExpiration(ttl),
		store.WithTags(tags),
	)
	if err != nil {
		log.Printf("[WARN] doSet cache.Set failed: %v", err)
	}

	return err
}
