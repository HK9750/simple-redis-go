package handler

import (
	"redis/resp"
	"sync"
)

var (
	store  = make(map[string]string)
	hashes = make(map[string]map[string]string)
	mu     sync.RWMutex
)

var Handlers = map[string]func([]resp.Value) resp.Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"DEL":     del,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
}

func ping(args []resp.Value) resp.Value {
	return resp.Value{Type: "string", Str: "PONG"}
}

func set(args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'set'"}
	}
	key := string(args[0].Bulk)
	value := string(args[1].Bulk)

	mu.Lock()
	store[key] = value
	mu.Unlock()

	return resp.Value{Type: "string", Str: "OK"}
}

func get(args []resp.Value) resp.Value {
	if len(args) < 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'get'"}
	}
	key := string(args[0].Bulk)

	mu.RLock()
	val, ok := store[key]
	mu.RUnlock()

	if !ok {
		return resp.Value{Type: "null"}
	}

	return resp.Value{Type: "bulk", Bulk: []byte(val)}
}

func del(args []resp.Value) resp.Value {
	if len(args) < 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'del'"}
	}

	mu.Lock()
	deleted := 0
	for i := 0; i < len(args); i++ {
		key := string(args[i].Bulk)
		if _, ok := store[key]; ok {
			delete(store, key)
			deleted++
		}
		if _, ok := hashes[key]; ok {
			delete(hashes, key)
			deleted++
		}
	}
	mu.Unlock()

	return resp.Value{Type: "integer", Num: int64(deleted)}
}

func hset(args []resp.Value) resp.Value {
	if len(args) < 3 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'hset'"}
	}

	key := string(args[0].Bulk)
	field := string(args[1].Bulk)
	value := string(args[2].Bulk)

	mu.Lock()
	if _, ok := hashes[key]; !ok {
		hashes[key] = make(map[string]string)
	}
	_, exists := hashes[key][field]
	hashes[key][field] = value
	mu.Unlock()

	if exists {
		return resp.Value{Type: "integer", Num: 0}
	}
	return resp.Value{Type: "integer", Num: 1}
}

func hget(args []resp.Value) resp.Value {
	if len(args) < 2 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'hget'"}
	}

	key := string(args[0].Bulk)
	field := string(args[1].Bulk)

	mu.RLock()
	defer mu.RUnlock()

	if h, ok := hashes[key]; ok {
		if val, ok := h[field]; ok {
			return resp.Value{Type: "bulk", Bulk: []byte(val)}
		}
	}

	return resp.Value{Type: "null"}
}

func hgetall(args []resp.Value) resp.Value {
	if len(args) < 1 {
		return resp.Value{Type: "error", Str: "ERR wrong number of arguments for 'hgetall'"}
	}

	key := string(args[0].Bulk)

	mu.RLock()
	defer mu.RUnlock()

	h, ok := hashes[key]
	if !ok {
		return resp.Value{Type: "array", Array: []resp.Value{}}
	}

	result := []resp.Value{}
	for k, v := range h {
		result = append(result,
			resp.Value{Type: "bulk", Bulk: []byte(k)},
			resp.Value{Type: "bulk", Bulk: []byte(v)},
		)
	}

	return resp.Value{Type: "array", Array: result}
}
