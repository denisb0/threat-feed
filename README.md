Threat Feed Ingestion Service (8–10 h)
**Problem** :  
Build a small HTTP service that:
1. Accepts POST `/ingest` with JSON body:  
   ```json
   { "feed": "malicious_ips", "entries": ["185.220.101.32", "185.220.101.33"] }
   ```
2. Deduplicates globally (across restarts) and stores entries in memory + persists to a file (`feed.jsonl`).
3. Exposes GET `/query?ip=1.2.3.4` → returns `{ "threat": true/false, "feed": "..." }` in <10 ms for 10 M entries.
4. Handles graceful shutdown (SIGTERM) — no data loss.
5. Rate-limits `/ingest` to 1000 req/s per source IP (token bucket).
6. All code must be **tested** (80%+ coverage) and include a `README.md` with build/run instructions.
7. Structured logging

**Bonus points** (what separates mid from senior):
- Use `sync.Pool` for JSON buffers.
- Memory-mapped file for persistence (`github.com/edsrzf/mmap-go`).
- Bloom filter to speed up negative lookups.
- `pprof` endpoint enabled.
- Dockerized.

**Files to deliver** (zip):
```
threat-feed/
├── main.go
├── ingest.go
├── query.go
├── storage.go
├── rate_limiter.go
├── go.mod
├── Dockerfile
├── README.md
└── threat_feed_test.go
```

### Calls
```bash
  curl -X POST http://localhost:8080/ingest \
    -H "Content-Type: application/json" \
    -d '{
      "feed": "malware-domains",
      "entries": ["malicious.example.com", "phishing.test.org", "bad-actor.net"]
    }'

  curl "http://localhost:8080/query?ip=bad-actor.net"
```