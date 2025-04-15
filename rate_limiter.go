package main

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	LoginLimiter   = make(map[string]*Limiter, 0)
	SignUpLimiter  = make(map[string]*Limiter, 0)
	generalLimiter = make(map[string]*Limiter, 0)
	mu             sync.Mutex
)

func getLimiter(ip string, route string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	var limiters map[string]*Limiter

	switch route {
	case "login":
		limiters = LoginLimiter
	case "signup":
		limiters = SignUpLimiter
	case "general":
		limiters = generalLimiter
	}

	if c, exists := limiters[ip]; exists {
		c.lastSeen = time.Now()
		return c.limiter
	}

	limiter := rate.NewLimiter(5, 10)

	limiters[ip] = &Limiter{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

func cleanUpOldTimers() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()

		for ip, client := range LoginLimiter {
			if time.Since(client.lastSeen) > 5*time.Minute {
				delete(LoginLimiter, ip)
			}
		}

		for ip, client := range SignUpLimiter {
			if time.Since(client.lastSeen) > 5*time.Minute {
				delete(LoginLimiter, ip)
			}
		}

		mu.Unlock()
	}
}

func rateLimit(w http.ResponseWriter, r *http.Request, route string) error {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip == "" {
		ip = r.RemoteAddr
	}
	limiter := getLimiter(ip, route)

	if !limiter.Allow() {
		return errors.New("user sent too many requests action denied")
	}

	return nil
}
