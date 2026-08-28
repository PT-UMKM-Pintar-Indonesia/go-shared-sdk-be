package sdk_cons

import "time"

// ============================================================================
// SINGLE OPERATION TIMEOUTS
// ============================================================================
const (
	RedisGetTimeout  = 2 * time.Second
	RedisSetTimeout  = 2 * time.Second
	RedisDelTimeout  = 2 * time.Second
	RedisMGetTimeout = 3 * time.Second

	DBGetTimeout    = 3 * time.Second
	DBInsertTimeout = 5 * time.Second
	DBUpdateTimeout = 8 * time.Second
	DBDeleteTimeout = 5 * time.Second

	PubTimeout = 3 * time.Second

	HTTPDialTimeout    = 10 * time.Second
	HTTPRequestTimeout = 30 * time.Second
)

// ============================================================================
// BULK OPERATION TIMEOUTS
// ============================================================================
const (
	RedisBulkGetTimeout  = 5 * time.Second
	RedisBulkSetTimeout  = 5 * time.Second
	RedisPipelineTimeout = 8 * time.Second

	DBBulkInsertTimeout = 15 * time.Second
	DBBulkUpdateTimeout = 20 * time.Second
	DBBulkDeleteTimeout = 15 * time.Second

	PubBulkTimeout  = 8 * time.Second
	ConsumerTimeout = 60 * time.Second
)

// ============================================================================
// BULK OPERATION LIMITS
// ============================================================================
const (
	RedisBulkGetLimit  = 500
	RedisBulkSetLimit  = 500
	RedisPipelineLimit = 1000

	DBBulkInsertLimit = 1000
	DBBulkUpdateLimit = 1000
	DBBulkDeleteLimit = 1000

	PubBulkLimit = 500
)

// ============================================================================
// CONTEXT HIERARCHY TIMEOUTS
// ============================================================================
const (
	HTTPHandlerTimeout   = 30 * time.Second
	DBTransactionTimeout = 20 * time.Second
	BackgroundJobTimeout = 5 * time.Minute
)
