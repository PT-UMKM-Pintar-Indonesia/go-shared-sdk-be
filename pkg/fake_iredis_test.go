package sdk_pkg

import (
	"context"
	"errors"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"

	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
)

// errFakeNotImplemented marks interface methods the JWT flows never call.
var errFakeNotImplemented = errors.New("fake iredis: not implemented")

// fakeIRedis is an in-memory implementation of sdk_inf.IRedis used to test
// the JWT sign/verify flows without any Redis server. Only the handful of
// commands the flows actually touch are functional.
type fakeIRedis struct {
	mu      sync.Mutex
	strs    map[string][]byte
	hashes  map[string]map[string][]byte
	expired map[string]bool // keys forced to look missing
}

var _ sdk_inf.IRedis = (*fakeIRedis)(nil)

func newFakeIRedis() *fakeIRedis {
	return &fakeIRedis{
		strs:    make(map[string][]byte),
		hashes:  make(map[string]map[string][]byte),
		expired: make(map[string]bool),
	}
}

func (f *fakeIRedis) Client(context.Context) *goredis.Client { return nil }

func (f *fakeIRedis) Set(_ context.Context, key string, value any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.strs[key] = toBytes(value)

	return nil
}

func (f *fakeIRedis) SetEx(_ context.Context, key string, _ time.Duration, value any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.strs[key] = toBytes(value)

	return nil
}

func (f *fakeIRedis) SetNX(ctx context.Context, key string, exp time.Duration, value any) error {
	f.mu.Lock()

	if _, ok := f.strs[key]; ok {
		f.mu.Unlock()

		return nil
	}

	f.mu.Unlock()

	return f.Set(ctx, key, value)
}

func (f *fakeIRedis) Get(_ context.Context, key string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.expired[key] {
		return nil, errors.New("redis get failed for key " + key + ": redis: nil")
	}

	val, ok := f.strs[key]
	if !ok {
		return nil, errors.New("redis get failed for key " + key + ": redis: nil")
	}

	return val, nil
}

func (f *fakeIRedis) Del(_ context.Context, key string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.strs[key]; ok {
		delete(f.strs, key)

		return 1, nil
	}

	return 0, nil
}

func (f *fakeIRedis) Exists(_ context.Context, key string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.expired[key] {
		return 0, nil
	}

	if _, ok := f.strs[key]; ok {
		return 1, nil
	}

	return 0, nil
}

func (f *fakeIRedis) HSet(_ context.Context, key string, values ...any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.hashes[key]; !ok {
		f.hashes[key] = make(map[string][]byte)
	}

	var flat []any

	for _, v := range values {
		switch typed := v.(type) {
		case []any:
			flat = append(flat, typed...)
		case []string:
			for i := range typed {
				flat = append(flat, typed[i])
			}
		case map[string]string:
			for field, val := range typed {
				f.hashes[key][field] = []byte(val)
			}
		case map[string]any:
			for field, val := range typed {
				f.hashes[key][field] = toBytes(val)
			}
		default:
			flat = append(flat, v)
		}
	}

	for i := 0; i+1 < len(flat); i += 2 {
		field, ok := flat[i].(string)
		if !ok {
			return errors.New("fake iredis: non-string field")
		}

		f.hashes[key][field] = toBytes(flat[i+1])
	}

	return nil
}

func (f *fakeIRedis) HSetEx(ctx context.Context, key string, _ time.Duration, values ...any) error {
	return f.HSet(ctx, key, values...)
}

func (f *fakeIRedis) HSetNX(ctx context.Context, key, field string, value any) error {
	f.mu.Lock()
	_, exists := f.hashes[key][field]
	f.mu.Unlock()

	if exists {
		return nil
	}

	return f.HSet(ctx, key, []any{field, value})
}

func (f *fakeIRedis) HGet(_ context.Context, key, field string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	h, ok := f.hashes[key]
	if !ok {
		return nil, errors.New("redis hget failed for key " + key + " field " + field + ": redis: nil")
	}

	val, ok := h[field]
	if !ok {
		return nil, errors.New("redis hget failed for key " + key + " field " + field + ": redis: nil")
	}

	return val, nil
}

func (f *fakeIRedis) MSet(context.Context, ...any) (string, error) { return "", errFakeNotImplemented }

func (f *fakeIRedis) MSetNX(context.Context, ...any) (bool, error) {
	return false, errFakeNotImplemented
}

func (f *fakeIRedis) MGet(context.Context, string) ([]any, error) { return nil, errFakeNotImplemented }

func (f *fakeIRedis) HGetAll(context.Context, string) (map[string]string, error) {
	return nil, errFakeNotImplemented
}

func (f *fakeIRedis) HExists(context.Context, string, string) (bool, error) {
	return false, errFakeNotImplemented
}

func (f *fakeIRedis) HDel(context.Context, string, ...string) (int64, error) {
	return 0, errFakeNotImplemented
}

func (f *fakeIRedis) HIncrByFloat(context.Context, string, string, float64) (float64, error) {
	return 0, errFakeNotImplemented
}

func (f *fakeIRedis) LPush(context.Context, string, ...any) ([]string, error) {
	return nil, errFakeNotImplemented
}

func (f *fakeIRedis) SAdd(context.Context, string, ...any) (int64, error) {
	return 0, errFakeNotImplemented
}

func (f *fakeIRedis) SIsMember(context.Context, string, any) (bool, error) {
	return false, errFakeNotImplemented
}

func (f *fakeIRedis) SRem(context.Context, string, any) (int64, error) {
	return 0, errFakeNotImplemented
}

func (f *fakeIRedis) SMembers(context.Context, string) ([]string, error) {
	return nil, errFakeNotImplemented
}

func (f *fakeIRedis) ZAdd(context.Context, string, ...goredis.Z) (int64, error) {
	return 0, errFakeNotImplemented
}

func (f *fakeIRedis) ZRange(context.Context, string) ([]string, error) {
	return nil, errFakeNotImplemented
}

func (f *fakeIRedis) IncrBy(context.Context, string, int) (int, error) {
	return 0, errFakeNotImplemented
}

func (f *fakeIRedis) TTL(context.Context, string) (int, error) { return 0, errFakeNotImplemented }

func (f *fakeIRedis) Unlink(context.Context, string) (int64, error) {
	return 0, errFakeNotImplemented
}

func (f *fakeIRedis) Expire(context.Context, string, time.Duration) (bool, error) {
	return false, errFakeNotImplemented
}

func toBytes(v any) []byte {
	switch t := v.(type) {
	case string:
		return []byte(t)
	case []byte:
		return t
	default:
		return []byte("x")
	}
}

func toAnySlice(v any) []any {
	switch t := v.(type) {
	case []any:
		return t
	case []string:
		out := make([]any, len(t))
		for i := range t {
			out[i] = t[i]
		}

		return out
	default:
		return []any{v}
	}
}
