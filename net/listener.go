// Kiebitz - Privacy-Friendly Appointment Scheduling
// Copyright (C) 2021-2021 The Kiebitz Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version. Additional terms
// as defined in section 7 of the license (e.g. regarding attribution)
// are specified at https://kiebitz.eu/en/docs/open-source/additional-terms.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package net

import (
	"fmt"
	"github.com/kiebitz-oss/services"
	"net"
	"sync"
	"time"
)

type RateLimitedListener struct {
	listener   net.Listener
	rateLimits []*services.RateLimit
	rates      []map[string]int64
	mutex      sync.Mutex
}

func MakeRateLimitedListener(listener net.Listener, rateLimits []*services.RateLimit) *RateLimitedListener {
	services.Log.Tracef("Creating rate-limited network listener...")
	rates := make([]map[string]int64, len(rateLimits))
	for i, _ := range rateLimits {
		rates[i] = make(map[string]int64)
	}
	return &RateLimitedListener{
		rates:      rates,
		listener:   listener,
		rateLimits: rateLimits,
	}
}

// Accept a connection, ensuring that rate limits are enforced
func (l *RateLimitedListener) Accept() (net.Conn, error) {
acceptLoop:
	for {
		if conn, err := l.listener.Accept(); err != nil {
			return nil, err
		} else {
			t := time.Now().UnixNano()
			l.mutex.Lock()
			var ip net.IP
			switch v := conn.RemoteAddr().(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			case *net.TCPAddr:
				ip = v.IP
			}
			key := ip.String()
			services.Log.Tracef("Got a connection from '%s'", key)
			for i, rateLimit := range l.rateLimits {
				services.Log.Tracef("Checking rate limit of type '%s' with limit %d", rateLimit.Type, rateLimit.Limit)
				tw := services.MakeTimeWindow(t, rateLimit.Type)
				if tw.Type == "" {
					l.mutex.Unlock()
					return nil, fmt.Errorf("invalid time window type: %s", rateLimit.Type)
				}
				if rateLimit.TimeWindow == nil {
					rateLimit.TimeWindow = &tw
				} else if !rateLimit.TimeWindow.EqualTo(&tw) {
					// this time window expired, we reset the rates
					l.rates[i] = make(map[string]int64)
					rateLimit.TimeWindow = &tw
				}
				rate, _ := l.rates[i][key]
				if rate >= rateLimit.Limit {
					services.Log.Tracef("Rate limit exceeded, closing connection...")
					if err := conn.Close(); err != nil {
						services.Log.Error(err)
					}
					l.mutex.Unlock()
					continue acceptLoop
				}
				l.rates[i][key] = rate + 1
			}
			l.mutex.Unlock()
			return conn, nil
		}
	}
}

func (l *RateLimitedListener) Close() error {
	return l.listener.Close()
}

func (l *RateLimitedListener) Addr() net.Addr {
	return l.listener.Addr()
}
