package persistent_cache

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"runtime"
	"sync"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

// Cache handles SQLite-based persistent caching with Gob encoding
type Cache struct {
	db *sql.DB
}

// _newCache initializes the cache database and table
func _newCache(dbFile string) *Cache {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	// Create table for caching function results
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS cache (
		id INTEGER PRIMARY KEY,
		function TEXT,
		args BLOB,
		result BLOB,
		timestamp INTEGER DEFAULT (strftime('%s', 'now'))
	);`)
	if err != nil {
		log.Fatal(err)
	}
	return &Cache{db}
}

var lock = &sync.Mutex{}

var _singleInstance *Cache

func _getPersistentCache() *Cache {
	if _singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if _singleInstance == nil {
			slog.Debug("Creating single instance now.")
			// TODO: Make the cache file path configurable
			_singleInstance = _newCache("cache.db")
		}
	}
	return _singleInstance
}

// _serialize encodes a Go object using Gob
func _serialize[T any](value T) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(value)
	return buffer.Bytes(), err
}

// _cacheSet stores a value in the cache with expiration
func _cacheSet[T any](func_name string, key string, value T) {
	serializedValue, err := _serialize(value)
	if err != nil {
		slog.Debug("Serialization error", "error", err)
		return
	}

	c := _getPersistentCache()
	_, err = c.db.Exec("INSERT INTO cache (function, args, result) VALUES (?, ?, ?);", func_name, key, serializedValue)
	if err != nil {
		slog.Debug("Cache Set Error:", "error", err)
	}
}

// _deserialize decodes a Gob-encoded byte array back into an object
func _deserialize[T any](data []byte) (T, error) {
	var value T
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&value)
	return value, err
}

// _cacheGet retrieves a value from the cache and deserializes it
func _cacheGet[T any](func_name string, key string, ttl time.Duration) (T, bool) {
	var data []byte
	var timestamp int64

	var zero T

	c := _getPersistentCache()
	err := c.db.
		QueryRow("SELECT result, timestamp FROM cache WHERE function = ? AND args = ?", func_name, key).
		Scan(&data, &timestamp)
	if err != nil {
		return zero, false
	}

	expirationTime := time.Unix(timestamp, 0).UTC().Add(ttl)
	currentTime := time.Now().UTC()

	// Check if expired
	if !currentTime.Before(expirationTime) {
		_, _ = c.db.Exec("DELETE FROM cache WHERE function = ? AND args = ?", func_name, key)
		return zero, false
	}

	item, err := _deserialize[T](data)
	if err != nil {
		slog.Debug("Deserialization error:", "error", err)
		return zero, false
	}

	return item, true
}

func _funcName(fn any) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

// _generateKey hashes function arguments into a unique cache key
func _generateKey(args ...any) string {
	return fmt.Sprintf("%v", args)
}

func Memoize0[R any](ttl time.Duration, fn func() R) func() R {
	func_name := _funcName(fn)
	return func() R {
		key := _generateKey()
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit", "key", key)
			return item
		}

		result := fn()
		_cacheSet(func_name, key, result)
		slog.Debug("Cache miss", "key", key)
		return result
	}
}

func Memoize1[A, R any](ttl time.Duration, fn func(A) R) func(A) R {
	func_name := _funcName(fn)
	return func(arg A) R {
		key := _generateKey(arg)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg)
		_cacheSet(func_name, key, result)
		return result
	}
}

func Memoize2[A, B, R any](ttl time.Duration, fn func(A, B) R) func(A, B) R {
	func_name := _funcName(fn)
	return func(arg1 A, arg2 B) R {
		key := _generateKey(arg1, arg2)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg1, arg2)
		_cacheSet(func_name, key, result)
		return result
	}
}

func Memoize3[A, B, C, R any](ttl time.Duration, fn func(A, B, C) R) func(A, B, C) R {
	func_name := _funcName(fn)
	return func(arg1 A, arg2 B, arg3 C) R {
		key := _generateKey(arg1, arg2, arg3)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg1, arg2, arg3)
		_cacheSet(func_name, key, result)
		return result
	}
}

func Memoize4[A, B, C, D, R any](ttl time.Duration, fn func(A, B, C, D) R) func(A, B, C, D) R {
	func_name := _funcName(fn)
	return func(arg1 A, arg2 B, arg3 C, arg4 D) R {
		key := _generateKey(arg1, arg2, arg3, arg4)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg1, arg2, arg3, arg4)
		_cacheSet(func_name, key, result)
		return result
	}
}

func Memoize5[A, B, C, D, E, R any](ttl time.Duration, fn func(A, B, C, D, E) R) func(A, B, C, D, E) R {
	func_name := _funcName(fn)
	return func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E) R {
		key := _generateKey(arg1, arg2, arg3, arg4, arg5)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg1, arg2, arg3, arg4, arg5)
		_cacheSet(func_name, key, result)
		return result
	}
}

func Memoize6[A, B, C, D, E, F, R any](ttl time.Duration, fn func(A, B, C, D, E, F) R) func(A, B, C, D, E, F) R {
	func_name := _funcName(fn)
	return func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E, arg6 F) R {
		key := _generateKey(arg1, arg2, arg3, arg4, arg5, arg6)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg1, arg2, arg3, arg4, arg5, arg6)
		_cacheSet(func_name, key, result)
		return result
	}
}

func Memoize7[A, B, C, D, E, F, G, R any](ttl time.Duration, fn func(A, B, C, D, E, F, G) R) func(A, B, C, D, E, F, G) R {
	func_name := _funcName(fn)
	return func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E, arg6 F, arg7 G) R {
		key := _generateKey(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
		_cacheSet(func_name, key, result)
		return result
	}
}

func Memoize8[A, B, C, D, E, F, G, H, R any](ttl time.Duration, fn func(A, B, C, D, E, F, G, H) R) func(A, B, C, D, E, F, G, H) R {
	func_name := _funcName(fn)
	return func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E, arg6 F, arg7 G, arg8 H) R {
		key := _generateKey(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
		_cacheSet(func_name, key, result)
		return result
	}
}

func Memoize9[A, B, C, D, E, F, G, H, I, R any](ttl time.Duration, fn func(A, B, C, D, E, F, G, H, I) R) func(A, B, C, D, E, F, G, H, I) R {
	func_name := _funcName(fn)
	return func(arg1 A, arg2 B, arg3 C, arg4 D, arg5 E, arg6 F, arg7 G, arg8 H, arg9 I) R {
		key := _generateKey(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
		_cacheSet(func_name, key, result)
		return result
	}
}

func MemoizeN[R any](ttl time.Duration, fn func(args ...any) R) func(...any) R {
	func_name := _funcName(fn)
	return func(args ...any) R {
		key := _generateKey(args...)
		item, found := _cacheGet[R](func_name, key, ttl)

		if found {
			slog.Debug("Cache hit:", "key", key)
			return item
		}

		slog.Debug("Cache miss:", "key", key)
		result := fn(args...)
		_cacheSet(func_name, key, result)
		return result
	}
}

func DeleteFuncCache(fn any) {
	func_name := _funcName(fn)
	c := _getPersistentCache()
	_, err := c.db.Exec("DELETE FROM cache WHERE function = ?", func_name)
	if err != nil {
		slog.Debug("Cache Delete Error:", "func_name", func_name)
	}
}
