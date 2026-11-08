// Package authguard bounds authentication attempts across transports.
package authguard

import (
	"net"
	"sync"
	"time"
)

const (
	window         = time.Minute
	perClientLimit = 20
	globalLimit    = 300
	maxClients     = 4096
)

type counter struct {
	start time.Time
	count int
}
type Limiter struct {
	mu      sync.Mutex
	clients map[string]counter
	global  counter
	now     func() time.Time
}

func New() *Limiter { return &Limiter{clients: make(map[string]counter), now: time.Now} }

// Allow uses the direct network peer; untrusted forwarding headers are never consulted.
func (l *Limiter) Allow(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	if host == "" {
		host = "unknown"
	}
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.global.start) >= window || l.global.start.IsZero() {
		l.global = counter{start: now}
	}
	if l.global.count >= globalLimit {
		return false
	}
	c, ok := l.clients[host]
	if !ok && len(l.clients) >= maxClients {
		for k, v := range l.clients {
			if now.Sub(v.start) >= window {
				delete(l.clients, k)
			}
		}
		if len(l.clients) >= maxClients {
			return false
		}
	}
	if !ok || now.Sub(c.start) >= window {
		c = counter{start: now}
	}
	if c.count >= perClientLimit {
		return false
	}
	c.count++
	l.global.count++
	l.clients[host] = c
	return true
}
