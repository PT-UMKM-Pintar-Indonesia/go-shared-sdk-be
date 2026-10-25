package sdk_pkg

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

// newTestRedis spins up the stdlib fake RESP server and returns the SDK
// wrapper bound to it. Protocol=2 keeps go-redis away from HELLO/RESP3.
func newTestRedis(t *testing.T) (sdk_inf.IRedis, context.Context) {
	t.Helper()

	addr := startFakeRedis(t)
	ctx := context.Background()

	rds, client, err := NewRedis(ctx, &sdk_dto.RedisClientOptions{
		Addr:        "redis://" + addr,
		Protocol:    2,
		PoolSize:    2,
		PoolFIFO:    true,
		MaxRetries:  1,
		DialTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("NewRedis: %v", err)
	}

	t.Cleanup(func() { _ = client.Close() })

	return rds, ctx
}

func TestNewRedisFailure(t *testing.T) {
	// Unreachable port must surface an error, not a half-open client.
	_, _, err := NewRedis(context.Background(), &sdk_dto.RedisClientOptions{
		Addr:         "redis://127.0.0.1:1",
		Protocol:     2,
		MaxRetries:   1,
		DialTimeout:  300 * time.Millisecond,
		ReadTimeout:  300 * time.Millisecond,
		WriteTimeout: 300 * time.Millisecond,
		PoolTimeout:  300 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error for unreachable redis")
	}
}

func TestRedisStringCommands(t *testing.T) {
	rds, ctx := newTestRedis(t)

	if rds.Client(ctx) == nil {
		t.Fatal("Client(ctx) returned nil")
	}

	if err := rds.Set(ctx, "k", "v"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, err := rds.Get(ctx, "k")
	if err != nil || string(val) != "v" {
		t.Fatalf("Get got %q err %v", val, err)
	}

	if _, err := rds.Get(ctx, "missing"); err == nil {
		t.Fatal("Get on missing key must error")
	}

	// SetNX on an existing key must not overwrite (SDK swallows the nil
	// reply and only surfaces transport errors).
	if err := rds.SetNX(ctx, "k", time.Minute, "other"); err != nil {
		t.Fatalf("SetNX: %v", err)
	}

	val, _ = rds.Get(ctx, "k")
	if string(val) != "v" {
		t.Fatalf("SetNX overwrote existing key: %q", val)
	}

	if err := rds.SetNX(ctx, "fresh", time.Minute, "f"); err != nil {
		t.Fatalf("SetNX fresh: %v", err)
	}

	if err := rds.SetEx(ctx, "tmp", 2*time.Minute, "x"); err != nil {
		t.Fatalf("SetEx: %v", err)
	}

	ttl, err := rds.TTL(ctx, "tmp")
	if err != nil || ttl < 110 || ttl > 120 {
		t.Fatalf("TTL got %d err %v", ttl, err)
	}

	// Sub-second expiry actually expires (PX path in go-redis).
	if err := rds.SetEx(ctx, "blink", 1200*time.Millisecond, "gone"); err != nil {
		t.Fatalf("SetEx blink: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)

	if _, err := rds.Get(ctx, "blink"); err == nil {
		t.Fatal("expired key must not be readable")
	}

	exists, err := rds.Exists(ctx, "k")
	if err != nil || exists != 1 {
		t.Fatalf("Exists got %d err %v", exists, err)
	}

	if _, err := rds.Expire(ctx, "k", time.Minute); err != nil {
		t.Fatalf("Expire: %v", err)
	}

	if ok, err := rds.Expire(ctx, "ghost", time.Minute); err != nil || ok {
		t.Fatalf("Expire on missing key got %v err %v", ok, err)
	}

	n, err := rds.Del(ctx, "k")
	if err != nil || n != 1 {
		t.Fatalf("Del got %d err %v", n, err)
	}

	n, err = rds.Unlink(ctx, "fresh")
	if err != nil || n != 1 {
		t.Fatalf("Unlink got %d err %v", n, err)
	}

	// Wrong-type access must surface an error.
	if _, err := rds.HGet(ctx, "tmp", "f"); err == nil {
		t.Fatal("HGet on a string key must error")
	}
}

func TestRedisCounters(t *testing.T) {
	rds, ctx := newTestRedis(t)

	if err := rds.Set(ctx, "ctr", "10"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, err := rds.IncrBy(ctx, "ctr", 5)
	if err != nil || val != 15 {
		t.Fatalf("IncrBy got %d err %v", val, err)
	}

	val, err = rds.IncrBy(ctx, "ctr-new", 3)
	if err != nil || val != 3 {
		t.Fatalf("IncrBy new got %d err %v", val, err)
	}

	if err := rds.Set(ctx, "nan", "abc"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if _, err := rds.IncrBy(ctx, "nan", 1); err == nil {
		t.Fatal("IncrBy on non-integer must error")
	}
}

func TestRedisMultiKey(t *testing.T) {
	rds, ctx := newTestRedis(t)

	res, err := rds.MSet(ctx, "a", "1", "b", "2")
	if err != nil || res != "OK" {
		t.Fatalf("MSet got %q err %v", res, err)
	}

	ok, err := rds.MSetNX(ctx, "a", "9", "c", "3")
	if err != nil || ok {
		t.Fatalf("MSetNX with existing key got %v err %v", ok, err)
	}

	ok, err = rds.MSetNX(ctx, "x", "1", "y", "2")
	if err != nil || !ok {
		t.Fatalf("MSetNX fresh got %v err %v", ok, err)
	}

	vals, err := rds.MGet(ctx, "a")
	if err != nil || len(vals) != 1 || vals[0] != "1" {
		t.Fatalf("MGet got %v err %v", vals, err)
	}
}

func TestRedisHashCommands(t *testing.T) {
	rds, ctx := newTestRedis(t)

	if err := rds.HSet(ctx, "h", "f1", "v1", "f2", "v2"); err != nil {
		t.Fatalf("HSet: %v", err)
	}

	val, err := rds.HGet(ctx, "h", "f1")
	if err != nil || string(val) != "v1" {
		t.Fatalf("HGet got %q err %v", val, err)
	}

	if _, err := rds.HGet(ctx, "h", "missing"); err == nil {
		t.Fatal("HGet on missing field must error")
	}

	if err := rds.HSetNX(ctx, "h", "f1", "dup"); err != nil {
		t.Fatalf("HSetNX existing: %v", err)
	}

	val, _ = rds.HGet(ctx, "h", "f1")
	if string(val) != "v1" {
		t.Fatalf("HSetNX overwrote field: %q", val)
	}

	if err := rds.HSetNX(ctx, "h", "f3", "v3"); err != nil {
		t.Fatalf("HSetNX fresh: %v", err)
	}

	// HSetEx takes field/value pairs via a slice and sets a TTL on the field.
	if err := rds.HSetEx(ctx, "he", time.Minute, "field", "value"); err != nil {
		t.Fatalf("HSetEx: %v", err)
	}

	if err := rds.HSetEx(ctx, "he-empty", time.Minute); err == nil {
		t.Fatal("HSetEx without values must error")
	}

	all, err := rds.HGetAll(ctx, "h")
	if err != nil || len(all) != 3 {
		t.Fatalf("HGetAll got %v err %v", all, err)
	}

	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if keys[0] != "f1" || keys[1] != "f2" || keys[2] != "f3" {
		t.Fatalf("HGetAll keys %v", keys)
	}

	ok, err := rds.HExists(ctx, "h", "f2")
	if err != nil || !ok {
		t.Fatalf("HExists got %v err %v", ok, err)
	}

	if ok, err := rds.HExists(ctx, "h", "nope"); err != nil || ok {
		t.Fatalf("HExists missing got %v err %v", ok, err)
	}

	f, err := rds.HIncrByFloat(ctx, "h", "counter", 1.5)
	if err != nil || f != 1.5 {
		t.Fatalf("HIncrByFloat got %v err %v", f, err)
	}

	f, err = rds.HIncrByFloat(ctx, "h", "counter", 2.25)
	if err != nil || f != 3.75 {
		t.Fatalf("HIncrByFloat accumulate got %v err %v", f, err)
	}

	n, err := rds.HDel(ctx, "h", "f2", "f3")
	if err != nil || n != 2 {
		t.Fatalf("HDel got %d err %v", n, err)
	}
}

func TestRedisListSetZSet(t *testing.T) {
	rds, ctx := newTestRedis(t)

	items, err := rds.LPush(ctx, "l", "a", "b")
	if err != nil || len(items) != 2 {
		t.Fatalf("LPush got %v err %v", items, err)
	}

	// LPUSH prepends: b ends up first.
	if items[0] != "b" || items[1] != "a" {
		t.Fatalf("LPush order %v", items)
	}

	added, err := rds.SAdd(ctx, "s", "m1", "m2")
	if err != nil || added != 2 {
		t.Fatalf("SAdd got %d err %v", added, err)
	}

	added, err = rds.SAdd(ctx, "s", "m1")
	if err != nil || added != 0 {
		t.Fatalf("SAdd dup got %d err %v", added, err)
	}

	isMember, err := rds.SIsMember(ctx, "s", "m2")
	if err != nil || !isMember {
		t.Fatalf("SIsMember got %v err %v", isMember, err)
	}

	if isMember, err := rds.SIsMember(ctx, "s", "ghost"); err != nil || isMember {
		t.Fatalf("SIsMember ghost got %v err %v", isMember, err)
	}

	members, err := rds.SMembers(ctx, "s")
	if err != nil || len(members) != 2 || members[0] != "m1" || members[1] != "m2" {
		t.Fatalf("SMembers got %v err %v", members, err)
	}

	removed, err := rds.SRem(ctx, "s", "m1")
	if err != nil || removed != 1 {
		t.Fatalf("SRem got %d err %v", removed, err)
	}

	zAdded, err := rds.ZAdd(ctx, "z", goredis.Z{Score: 2, Member: "b"}, goredis.Z{Score: 1, Member: "a"})
	if err != nil || zAdded != 2 {
		t.Fatalf("ZAdd got %d err %v", zAdded, err)
	}

	zRange, err := rds.ZRange(ctx, "z")
	if err != nil || len(zRange) != 2 || zRange[0] != "a" || zRange[1] != "b" {
		t.Fatalf("ZRange got %v err %v", zRange, err)
	}

	// Re-adding an existing member updates the score without growing.
	if n, err := rds.ZAdd(ctx, "z", goredis.Z{Score: 3, Member: "a"}); err != nil || n != 0 {
		t.Fatalf("ZAdd update got %d err %v", n, err)
	}

	if r, err := rds.ZRange(ctx, "z"); err != nil || r[0] != "b" || r[1] != "a" {
		t.Fatalf("ZRange after score update got %v err %v", r, err)
	}
}

func TestRedisConcurrentAccess(t *testing.T) {
	rds, ctx := newTestRedis(t)

	const writers = 16

	done := make(chan struct{}, writers)

	for i := 0; i < writers; i++ {
		go func(i int) {
			defer func() { done <- struct{}{} }()

			key := fmt.Sprintf("conc-%d", i)
			if err := rds.Set(ctx, key, i); err != nil {
				t.Errorf("Set %s: %v", key, err)
			}
			if _, err := rds.IncrBy(ctx, "shared-counter", 1); err != nil {
				t.Errorf("IncrBy: %v", err)
			}
		}(i)
	}

	for i := 0; i < writers; i++ {
		<-done
	}

	val, err := rds.IncrBy(ctx, "shared-counter", 0)
	if err != nil || val != writers {
		t.Fatalf("counter got %d err %v", val, err)
	}
}
