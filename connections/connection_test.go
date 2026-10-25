package sdk_con

import (
	"context"
	"testing"
	"time"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_dro "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

func TestHeartbeatDuration(t *testing.T) {
	// Zero/negative falls back to the documented default.
	if got := heartbeatDuration(0); got != time.Duration(DefaultHeartbeat)*time.Second {
		t.Fatalf("zero heartbeat got %v", got)
	}

	if got := heartbeatDuration(-5); got != time.Duration(DefaultHeartbeat)*time.Second {
		t.Fatalf("negative heartbeat got %v", got)
	}

	// Legacy raw-seconds values (< 1s as a Duration) are scaled to seconds.
	if got := heartbeatDuration(30); got != 30*time.Second {
		t.Fatalf("legacy seconds got %v", got)
	}

	// Real durations pass through untouched.
	if got := heartbeatDuration(45 * time.Second); got != 45*time.Second {
		t.Fatalf("duration passthrough got %v", got)
	}
}

func TestRabbitMQConnectionNilOptions(t *testing.T) {
	if _, err := RabbitMQConnection(nil); err == nil {
		t.Fatal("expected error for nil options")
	}
}

func TestRabbitMQConnectionEmptyURL(t *testing.T) {
	opt := &sdk_dto.RabbitClientOptions{}

	if _, err := RabbitMQConnection(opt); err == nil {
		t.Fatal("expected error for empty url")
	}
}

func TestRedisConnectionRequiresAddr(t *testing.T) {
	if _, err := RedisConnection(context.Background(), &sdk_dto.RedisClientOptions{}); err == nil {
		t.Fatal("expected error for empty redis addr")
	}
}

func TestSqlConnectionUnsupportedDriver(t *testing.T) {
	env := &sdk_dro.Environment{}

	if _, err := SqlConnection(context.Background(), "app", "oracle", env); err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

// A6 regression: a failed ping used to leak the opened *sql.DB. The fixed
// code closes it and returns the error; we assert the error path returns
// quickly with an unreachable local port.
func TestSqlConnectionPingFailure(t *testing.T) {
	env := &sdk_dro.Environment{}
	env.APP.ENV = "test"
	env.MYSQL.URL = "root:root@tcp(127.0.0.1:1)/nope?timeout=2s"
	env.DBCONFIG = sdk_dro.DatabaseConfig{
		TIMEOUT:       2,
		DIAL_TIMEOUT:  2,
		READ_TIMEOUT:  2,
		WRITE_TIMEOUT: 2,
		MAX_CONN:      1,
		MAX_IDLE:      1,
		CON_MAX:       2,
		CON_IDLE:      2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()

	if _, err := SqlConnection(ctx, "app", "mysql", env); err == nil {
		t.Fatal("expected connection error for unreachable mysql")
	}

	if time.Since(start) > 5*time.Second {
		t.Fatal("ping failure did not respect timeout")
	}
}
