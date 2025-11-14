package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	tests := []struct {
		name        string
		data        []IngestPayload
		query       string
		expected    QueryResponse
		expectError bool
	}{
		{
			name: "empty",
		},
		{
			name: "basic",
			data: []IngestPayload{
				{
					Feed:    "test-feed",
					Entries: []string{"abcd"},
				},
			},
			query: "abcd",
			expected: QueryResponse{
				Threat: true,
				Feed:   "test-feed",
			},
		},
		{
			name: "multi",
			data: []IngestPayload{
				{
					Feed:    "test-feed2",
					Entries: []string{"abcd", "efgh"},
				},
				{
					Feed:    "test-feed3",
					Entries: []string{"foo", "bar"},
				},
			},
			query: "bar",
			expected: QueryResponse{
				Threat: true,
				Feed:   "test-feed3",
			},
		},
		{
			name: "miss",
			data: []IngestPayload{
				{
					Feed:    "test-feed2",
					Entries: []string{"abcd", "efgh"},
				},
				{
					Feed:    "test-feed3",
					Entries: []string{"foo", "bar"},
				},
			},
			query: "baz",
			expected: QueryResponse{
				Threat: false,
			},
		},
		{
			name: "ingest-existing",
			data: []IngestPayload{
				{
					Feed:    "test-feed2",
					Entries: []string{"abcd", "efgh"},
				},
				{
					Feed:    "test-feed3",
					Entries: []string{"foo", "bar"},
				},
				{
					Feed:    "test-feed2",
					Entries: []string{"aaa", "bbb"},
				},
			},
			query: "aaa",
			expected: QueryResponse{
				Threat: true,
				Feed:   "test-feed2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storage := NewThreatMemStorage()
			for _, d := range tt.data {
				require.NoError(t, storage.Ingest(context.Background(), d))
			}
			result, err := storage.Query(context.Background(), tt.query)
			assert.Equal(t, tt.expectError, err != nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkQuery(b *testing.B) {
	storage := NewThreatMemStorage()
	ctx := context.Background()

	// Ingest multiple feeds with many IPs
	numFeeds := 100
	entriesPerFeed := 1000
	for i := 0; i < numFeeds; i++ {
		entries := make([]string, entriesPerFeed)
		for j := 0; j < entriesPerFeed; j++ {
			// Generate IP-like strings
			entries[j] = generateIP(i, j)
		}
		payload := IngestPayload{
			Feed:    generateFeedName(i),
			Entries: entries,
		}
		if err := storage.Ingest(ctx, payload); err != nil {
			b.Fatal(err)
		}
	}

	b.Run("QueryHit", func(b *testing.B) {
		// Query for IPs that exist in the storage
		testIP := generateIP(50, 500) // Middle of the dataset
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := storage.Query(ctx, testIP)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("QueryMiss", func(b *testing.B) {
		// Query for IPs that don't exist
		testIP := "999.999.999.999"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := storage.Query(ctx, testIP)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("QueryParallel", func(b *testing.B) {
		testIP := generateIP(50, 500)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := storage.Query(ctx, testIP)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkIngest(b *testing.B) {
	ctx := context.Background()
	entriesPerFeed := 1000

	b.Run("IngestSmallFeed", func(b *testing.B) {
		storage := NewThreatMemStorage()
		entries := make([]string, 100)
		for j := 0; j < 100; j++ {
			entries[j] = generateIP(0, j)
		}
		payload := IngestPayload{
			Feed:    "test-feed",
			Entries: entries,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := storage.Ingest(ctx, payload); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("IngestLargeFeed", func(b *testing.B) {
		storage := NewThreatMemStorage()
		entries := make([]string, entriesPerFeed)
		for j := 0; j < entriesPerFeed; j++ {
			entries[j] = generateIP(0, j)
		}
		payload := IngestPayload{
			Feed:    "test-feed",
			Entries: entries,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := storage.Ingest(ctx, payload); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("IngestReplaceExisting", func(b *testing.B) {
		storage := NewThreatMemStorage()
		// Pre-populate with data
		oldEntries := make([]string, entriesPerFeed)
		for j := 0; j < entriesPerFeed; j++ {
			oldEntries[j] = generateIP(0, j)
		}
		if err := storage.Ingest(ctx, IngestPayload{
			Feed:    "test-feed",
			Entries: oldEntries,
		}); err != nil {
			b.Fatal(err)
		}

		newEntries := make([]string, entriesPerFeed)
		for j := 0; j < entriesPerFeed; j++ {
			newEntries[j] = generateIP(1, j)
		}
		payload := IngestPayload{
			Feed:    "test-feed",
			Entries: newEntries,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := storage.Ingest(ctx, payload); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Helper functions
func generateIP(feedNum, entryNum int) string {
	// Generate pseudo-IP addresses for testing
	return string([]byte{
		byte(feedNum >> 8),
		byte(feedNum & 0xFF),
		byte(entryNum >> 8),
		byte(entryNum & 0xFF),
	})
}

func generateFeedName(feedNum int) string {
	return "feed-" + string(rune('A'+feedNum%26)) + string(rune('0'+feedNum/26))
}
