package ws

import (
	"errors"
	"sync"
)

// HINT: concurrency safe maps

// safePool connection pool
type safePool struct {
	sync.RWMutex
	// List of the connections alive
	conn map[string]IClient
}

func (p *safePool) set(ws IClient) {
	p.Lock()
	p.conn[ws.GetID()] = ws
	p.Unlock()
}

func (p *safePool) all() map[string]IClient {
	p.RLock()
	ret := make(map[string]IClient, 0)
	for wsUUID, kws := range p.conn {
		ret[wsUUID] = kws
	}
	p.RUnlock()
	return ret
}

func (p *safePool) get(key string) (IClient, error) {
	p.RLock()
	ret, ok := p.conn[key]
	p.RUnlock()
	if !ok {
		return nil, errors.New("invalid conn")
	}
	return ret, nil
}

func (p *safePool) contains(key string) bool {
	p.RLock()
	_, ok := p.conn[key]
	p.RUnlock()
	return ok
}

func (p *safePool) delete(key string) {
	p.Lock()
	delete(p.conn, key)
	p.Unlock()
}

//nolint:all
func (p *safePool) reset() {
	p.Lock()
	p.conn = make(map[string]IClient)
	p.Unlock()
}

// safeListeners hub event handlers
type safeListeners struct {
	sync.RWMutex
	list map[string][]HubHandlerFn
}

func (l *safeListeners) set(event string, callback HubHandlerFn) {
	l.Lock()
	l.list[event] = append(l.list[event], callback)
	l.Unlock()
}

func (l *safeListeners) get(event string) []HubHandlerFn {
	l.RLock()
	defer l.RUnlock()
	if _, ok := l.list[event]; !ok {
		return make([]HubHandlerFn, 0)
	}

	ret := make([]HubHandlerFn, 0)
	ret = append(ret, l.list[event]...)
	return ret
}
