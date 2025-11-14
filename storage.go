package main

import (
	"context"
	"sync"
)

type ThreatMemStorage struct {
	mu        sync.RWMutex
	ipToFeed  map[string]string // ip:feed
	feedToIPs map[string][]string
}

func NewThreatMemStorage() *ThreatMemStorage {
	return &ThreatMemStorage{
		ipToFeed:  make(map[string]string),
		feedToIPs: make(map[string][]string),
	}
}

func (ts *ThreatMemStorage) Query(ctx context.Context, ip string) (QueryResponse, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	var resp QueryResponse
	if feed, found := ts.ipToFeed[ip]; found {
		resp.Feed = feed
		resp.Threat = true
	}

	return resp, nil
}

func (ts *ThreatMemStorage) Ingest(ctx context.Context, data IngestPayload) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if oldIPs, exists := ts.feedToIPs[data.Feed]; exists {
		for _, ip := range oldIPs {
			delete(ts.ipToFeed, ip)
		}
	}

	for _, ip := range data.Entries {
		ts.ipToFeed[ip] = data.Feed
	}

	ts.feedToIPs[data.Feed] = data.Entries

	return nil
}
