package metrics

import (
	"testing"
)

func TestRecordRequest(t *testing.T) {
	RecordRequest("GET", "/api/users", "200", 0.1)
	RecordRequest("POST", "/api/users", "201", 0.2)
	RecordRequest("GET", "/api/users/1", "404", 0.05)
}

func TestRecordDbQuery(t *testing.T) {
	RecordDbQuery("SELECT", "users", 0.01)
	RecordDbQuery("INSERT", "orders", 0.02)
	RecordDbQuery("UPDATE", "products", 0.015)
}

func TestRecordCacheHit(t *testing.T) {
	RecordCacheHit(true)
	RecordCacheHit(false)
	RecordCacheHit(true)
}

func TestMetricsDefined(t *testing.T) {
	if HttpRequestsTotal == nil {
		t.Error("HttpRequestsTotal should not be nil")
	}
	if HttpRequestDuration == nil {
		t.Error("HttpRequestDuration should not be nil")
	}
	if ActiveConnections == nil {
		t.Error("ActiveConnections should not be nil")
	}
	if DbQueryDuration == nil {
		t.Error("DbQueryDuration should not be nil")
	}
	if CacheHits == nil {
		t.Error("CacheHits should not be nil")
	}
}
