package sdk_pkg

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeRedisServer is a minimal, stdlib-only RESP2 server that implements just
// enough of the Redis command surface to exercise the SDK's redis wrapper
// (pkg/redis.pkg.go) without a real Redis process. It is intentionally
// simplistic: lazy expiry, no persistence, unknown commands answer +OK so
// go-redis handshakes (CLIENT SETINFO, ...) never fail.
type fakeRedisServer struct {
	ln net.Listener

	mu          sync.Mutex
	strs        map[string]string
	hashes      map[string]map[string]string
	lists       map[string][]string
	sets        map[string]map[string]struct{}
	zsets       map[string][]fakeZMember
	expiry      map[string]time.Time
	connections sync.WaitGroup

	connMu sync.Mutex
	conns  []net.Conn
}

func (s *fakeRedisServer) trackConn(c net.Conn) {
	s.connMu.Lock()
	s.conns = append(s.conns, c)
	s.connMu.Unlock()
}

func (s *fakeRedisServer) closeConns() {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	for _, c := range s.conns {
		_ = c.Close()
	}

	s.conns = nil
}

type fakeZMember struct {
	score  float64
	member string
}

func startFakeRedis(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake redis listen: %v", err)
	}

	s := &fakeRedisServer{
		ln:     ln,
		strs:   make(map[string]string),
		hashes: make(map[string]map[string]string),
		lists:  make(map[string][]string),
		sets:   make(map[string]map[string]struct{}),
		zsets:  make(map[string][]fakeZMember),
		expiry: make(map[string]time.Time),
	}

	go s.serve()

	t.Cleanup(s.Stop)

	return ln.Addr().String()
}

func (s *fakeRedisServer) Stop() {
	_ = s.ln.Close()
	s.closeConns()
	s.connections.Wait()
}

func (s *fakeRedisServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}

		s.trackConn(conn)

		s.connections.Add(1)

		go func() {
			defer s.connections.Done()
			defer conn.Close()

			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)

			for {
				args, err := readRESPCommand(reader)
				if err != nil {
					return
				}

				if len(args) == 0 {
					continue
				}

				reply := s.exec(args)

				if _, err := writer.WriteString(reply); err != nil {
					return
				}

				if err := writer.Flush(); err != nil {
					return
				}
			}
		}()
	}
}

func readRESPCommand(r *bufio.Reader) ([]string, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	if prefix != '*' {
		return nil, fmt.Errorf("expected array, got %q", prefix)
	}

	n, err := readIntLine(r)
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, n)

	for i := 0; i < n; i++ {
		bulkPrefix, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		if bulkPrefix != '$' {
			return nil, fmt.Errorf("expected bulk, got %q", bulkPrefix)
		}

		length, err := readIntLine(r)
		if err != nil {
			return nil, err
		}

		buf := make([]byte, length+2) // payload + CRLF

		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}

		args = append(args, string(buf[:length]))
	}

	return args, nil
}

func readIntLine(r *bufio.Reader) (int, error) {
	line, err := readLine(r)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(line)
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(line, "\r\n"), nil
}

// ---- reply helpers ----

func replyOK() string              { return "+OK\r\n" }
func replyPong() string            { return "+PONG\r\n" }
func replyInt(n int64) string      { return fmt.Sprintf(":%d\r\n", n) }
func replyNil() string             { return "$-1\r\n" }
func replyError(msg string) string { return "-ERR " + msg + "\r\n" }

func replyBulk(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func replyArray(items []string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "*%d\r\n", len(items))

	for _, item := range items {
		b.WriteString(replyBulk(item))
	}

	return b.String()
}

// ---- store helpers (all called with s.mu held) ----

func (s *fakeRedisServer) isExpired(key string) bool {
	exp, ok := s.expiry[key]
	if !ok {
		return false
	}

	if time.Now().Before(exp) {
		return false
	}

	s.drop(key)

	return true
}

func (s *fakeRedisServer) drop(key string) {
	delete(s.strs, key)
	delete(s.hashes, key)
	delete(s.lists, key)
	delete(s.sets, key)
	delete(s.zsets, key)
	delete(s.expiry, key)
}

func (s *fakeRedisServer) exists(key string) bool {
	if s.isExpired(key) {
		return false
	}

	_, ok := s.strs[key]
	_, okH := s.hashes[key]
	_, okL := s.lists[key]
	_, okS := s.sets[key]
	_, okZ := s.zsets[key]

	return ok || okH || okL || okS || okZ
}

func (s *fakeRedisServer) holds(key string) (str, hash, list, set, zset bool) {
	_, str = s.strs[key]
	_, hash = s.hashes[key]
	_, list = s.lists[key]
	_, set = s.sets[key]
	_, zset = s.zsets[key]

	return
}

// wrongType reports an error reply when key already holds another type.
func (s *fakeRedisServer) wrongType(key, kind string) (string, bool) {
	if !s.exists(key) {
		return "", false
	}

	str, hash, list, set, zset := s.holds(key)

	mismatch := false

	switch kind {
	case "string":
		mismatch = hash || list || set || zset
	case "hash":
		mismatch = str || list || set || zset
	case "list":
		mismatch = str || hash || set || zset
	case "set":
		mismatch = str || hash || list || zset
	case "zset":
		mismatch = str || hash || list || set
	}

	if mismatch {
		return replyError("WRONGTYPE Operation against a key holding the wrong kind of value"), true
	}

	return "", false
}

// ---- command dispatch ----

func (s *fakeRedisServer) exec(args []string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := ""
	if len(args) > 1 {
		key = args[1]
	}

	switch strings.ToUpper(args[0]) {
	case "PING":
		return replyPong()

	case "HELLO":
		return replyError("unknown command 'HELLO'")

	case "SELECT", "AUTH":
		return replyOK()

	case "SET":
		return s.cmdSet(args)

	case "GET":
		if len(args) < 2 {
			return replyError("wrong number of arguments for 'get' command")
		}

		if !s.exists(key) {
			return replyNil()
		}

		if bad, ok := s.wrongType(key, "string"); ok {
			return bad
		}

		return replyBulk(s.strs[key])

	case "SETEX":
		// SETEX key seconds value
		if len(args) < 4 {
			return replyError("wrong number of arguments for 'setex' command")
		}

		secs, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return replyError("invalid expire time")
		}

		s.drop(key)
		s.strs[key] = args[3]
		s.expiry[key] = time.Now().Add(time.Duration(secs) * time.Second)

		return replyOK()

	case "PSETEX":
		// PSETEX key milliseconds value
		if len(args) < 4 {
			return replyError("wrong number of arguments for 'psetex' command")
		}

		ms, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return replyError("invalid expire time")
		}

		s.drop(key)
		s.strs[key] = args[3]
		s.expiry[key] = time.Now().Add(time.Duration(ms) * time.Millisecond)

		return replyOK()

	case "SETNX":
		if len(args) < 3 {
			return replyError("wrong number of arguments for 'setnx' command")
		}

		if s.exists(key) {
			if bad, ok := s.wrongType(key, "string"); ok {
				return bad
			}

			return replyInt(0)
		}

		s.strs[key] = args[2]

		return replyInt(1)

	case "DEL", "UNLINK":
		var count int64

		for _, k := range args[1:] {
			if s.exists(k) {
				s.drop(k)
				count++
			}
		}

		return replyInt(count)

	case "EXISTS":
		var count int64

		for _, k := range args[1:] {
			if s.exists(k) {
				count++
			}
		}

		return replyInt(count)

	case "MSET":
		for i := 1; i+1 < len(args); i += 2 {
			s.drop(args[i])
			s.strs[args[i]] = args[i+1]
		}

		return replyOK()

	case "MSETNX":
		for i := 1; i+1 < len(args); i += 2 {
			if s.exists(args[i]) {
				return replyInt(0)
			}
		}

		for i := 1; i+1 < len(args); i += 2 {
			s.strs[args[i]] = args[i+1]
		}

		return replyInt(1)

	case "MGET":
		items := make([]string, 0, len(args)-1)

		for _, k := range args[1:] {
			if s.exists(k) {
				items = append(items, "\x00v"+s.strs[k])
			} else {
				items = append(items, "\x00nil")
			}
		}

		return replyMixedArray(items)

	case "TTL":
		if !s.exists(key) {
			return replyInt(-2)
		}

		exp, ok := s.expiry[key]
		if !ok {
			return replyInt(-1)
		}

		remaining := time.Until(exp)
		if remaining < 0 {
			remaining = 0
		}

		seconds := int64(remaining.Seconds())
		if remaining.Seconds()-float64(seconds) > 0 {
			seconds++
		}

		return replyInt(seconds)

	case "EXPIRE", "PEXPIRE":
		if len(args) < 3 {
			return replyError("wrong number of arguments for 'expire' command")
		}

		ttlVal, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return replyError("value is not an integer or out of range")
		}

		if !s.exists(key) {
			return replyInt(0)
		}

		if strings.EqualFold(args[0], "pexpire") {
			s.expiry[key] = time.Now().Add(time.Duration(ttlVal) * time.Millisecond)
		} else {
			s.expiry[key] = time.Now().Add(time.Duration(ttlVal) * time.Second)
		}

		return replyInt(1)

	case "INCRBY":
		if len(args) < 3 {
			return replyError("wrong number of arguments for 'incrby' command")
		}

		delta, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return replyError("value is not an integer or out of range")
		}

		if s.exists(key) {
			if bad, ok := s.wrongType(key, "string"); ok {
				return bad
			}

			cur, err := strconv.ParseInt(s.strs[key], 10, 64)
			if err != nil {
				return replyError("value is not an integer or out of range")
			}

			cur += delta
			s.strs[key] = strconv.FormatInt(cur, 10)

			return replyInt(cur)
		}

		s.strs[key] = strconv.FormatInt(delta, 10)

		return replyInt(delta)

	case "HSET":
		return s.cmdHSet(args)

	case "HSETNX":
		if len(args) < 4 {
			return replyError("wrong number of arguments for 'hsetnx' command")
		}

		if s.exists(key) {
			if bad, ok := s.wrongType(key, "hash"); ok {
				return bad
			}
		}

		if _, ok := s.hashes[key]; !ok {
			s.hashes[key] = make(map[string]string)
		}

		if _, ok := s.hashes[key][args[2]]; ok {
			return replyInt(0)
		}

		s.hashes[key][args[2]] = args[3]

		return replyInt(1)

	case "HGET":
		if len(args) < 3 {
			return replyError("wrong number of arguments for 'hget' command")
		}

		if !s.exists(key) {
			return replyNil()
		}

		if bad, ok := s.wrongType(key, "hash"); ok {
			return bad
		}

		val, ok := s.hashes[key][args[2]]
		if !ok {
			return replyNil()
		}

		return replyBulk(val)

	case "HGETALL":
		if !s.exists(key) {
			return replyArray(nil)
		}

		if bad, ok := s.wrongType(key, "hash"); ok {
			return bad
		}

		fields := make([]string, 0, len(s.hashes[key]))

		for f := range s.hashes[key] {
			fields = append(fields, f)
		}

		sort.Strings(fields)

		items := make([]string, 0, len(fields)*2)

		for _, f := range fields {
			items = append(items, f, s.hashes[key][f])
		}

		return replyArray(items)

	case "HEXISTS":
		if len(args) < 3 {
			return replyError("wrong number of arguments for 'hexists' command")
		}

		if !s.exists(key) {
			return replyInt(0)
		}

		if bad, ok := s.wrongType(key, "hash"); ok {
			return bad
		}

		if _, ok := s.hashes[key][args[2]]; ok {
			return replyInt(1)
		}

		return replyInt(0)

	case "HDEL":
		if !s.exists(key) {
			return replyInt(0)
		}

		if bad, ok := s.wrongType(key, "hash"); ok {
			return bad
		}

		var count int64

		for _, f := range args[2:] {
			if _, ok := s.hashes[key][f]; ok {
				delete(s.hashes[key], f)
				count++
			}
		}

		if len(s.hashes[key]) == 0 {
			delete(s.hashes, key)
			delete(s.expiry, key)
		}

		return replyInt(count)

	case "HINCRBYFLOAT":
		if len(args) < 4 {
			return replyError("wrong number of arguments for 'hincrbyfloat' command")
		}

		delta, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			return replyError("value is not a valid float")
		}

		if s.exists(key) {
			if bad, ok := s.wrongType(key, "hash"); ok {
				return bad
			}
		}

		if _, ok := s.hashes[key]; !ok {
			s.hashes[key] = make(map[string]string)
		}

		cur := 0.0
		if raw, ok := s.hashes[key][args[2]]; ok {
			cur, err = strconv.ParseFloat(raw, 64)
			if err != nil {
				return replyError("value is not a valid float")
			}
		}

		cur += delta
		out := strconv.FormatFloat(cur, 'f', -1, 64)
		s.hashes[key][args[2]] = out

		return replyBulk(out)

	case "LPUSH":
		if s.exists(key) {
			if bad, ok := s.wrongType(key, "list"); ok {
				return bad
			}
		}

		for _, v := range args[2:] {
			s.lists[key] = append([]string{v}, s.lists[key]...)
		}

		return replyInt(int64(len(s.lists[key])))

	case "LRANGE":
		if len(args) < 4 {
			return replyError("wrong number of arguments for 'lrange' command")
		}

		if !s.exists(key) {
			return replyArray(nil)
		}

		if bad, ok := s.wrongType(key, "list"); ok {
			return bad
		}

		list := s.lists[key]
		n := len(list)

		start, _ := strconv.Atoi(args[2])
		stop, _ := strconv.Atoi(args[3])

		if start < 0 {
			start += n
		}

		if stop < 0 {
			stop += n
		}

		if start < 0 {
			start = 0
		}

		if stop >= n {
			stop = n - 1
		}

		if start > stop || start >= n {
			return replyArray(nil)
		}

		return replyArray(list[start : stop+1])

	case "SADD":
		if s.exists(key) {
			if bad, ok := s.wrongType(key, "set"); ok {
				return bad
			}
		}

		if _, ok := s.sets[key]; !ok {
			s.sets[key] = make(map[string]struct{})
		}

		var added int64

		for _, m := range args[2:] {
			if _, ok := s.sets[key][m]; !ok {
				s.sets[key][m] = struct{}{}
				added++
			}
		}

		return replyInt(added)

	case "SISMEMBER":
		if len(args) < 3 {
			return replyError("wrong number of arguments for 'sismember' command")
		}

		if !s.exists(key) {
			return replyInt(0)
		}

		if bad, ok := s.wrongType(key, "set"); ok {
			return bad
		}

		if _, ok := s.sets[key][args[2]]; ok {
			return replyInt(1)
		}

		return replyInt(0)

	case "SREM":
		if !s.exists(key) {
			return replyInt(0)
		}

		if bad, ok := s.wrongType(key, "set"); ok {
			return bad
		}

		var count int64

		for _, m := range args[2:] {
			if _, ok := s.sets[key][m]; ok {
				delete(s.sets[key], m)
				count++
			}
		}

		if len(s.sets[key]) == 0 {
			delete(s.sets, key)
			delete(s.expiry, key)
		}

		return replyInt(count)

	case "SMEMBERS":
		if !s.exists(key) {
			return replyArray(nil)
		}

		if bad, ok := s.wrongType(key, "set"); ok {
			return bad
		}

		members := make([]string, 0, len(s.sets[key]))

		for m := range s.sets[key] {
			members = append(members, m)
		}

		sort.Strings(members)

		return replyArray(members)

	case "ZADD":
		if s.exists(key) {
			if bad, ok := s.wrongType(key, "zset"); ok {
				return bad
			}
		}

		var added int64

		for i := 2; i+1 < len(args); i += 2 {
			score, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return replyError("value is not a valid float")
			}

			member := args[i+1]

			replaced := false

			for idx, z := range s.zsets[key] {
				if z.member == member {
					s.zsets[key][idx].score = score
					replaced = true

					break
				}
			}

			if !replaced {
				s.zsets[key] = append(s.zsets[key], fakeZMember{score: score, member: member})
				added++
			}
		}

		sort.SliceStable(s.zsets[key], func(a, b int) bool {
			if s.zsets[key][a].score != s.zsets[key][b].score {
				return s.zsets[key][a].score < s.zsets[key][b].score
			}

			return s.zsets[key][a].member < s.zsets[key][b].member
		})

		return replyInt(added)

	case "ZRANGE":
		if !s.exists(key) {
			return replyArray(nil)
		}

		if bad, ok := s.wrongType(key, "zset"); ok {
			return bad
		}

		zs := s.zsets[key]
		n := len(zs)

		start, _ := strconv.Atoi(args[2])
		stop, _ := strconv.Atoi(args[3])

		if start < 0 {
			start += n
		}

		if stop < 0 {
			stop += n
		}

		if start < 0 {
			start = 0
		}

		if stop >= n {
			stop = n - 1
		}

		if start > stop || start >= n {
			return replyArray(nil)
		}

		members := make([]string, 0, stop-start+1)

		for _, z := range zs[start : stop+1] {
			members = append(members, z.member)
		}

		return replyArray(members)

	case "HEXPIRE":
		// HEXPIRE key <ex|px> <t> [NX|XX|GT|LT] FIELDS n f1 ...
		fieldsIdx := -1

		for i, a := range args {
			if strings.EqualFold(a, "FIELDS") {
				fieldsIdx = i

				break
			}
		}

		if fieldsIdx < 0 || fieldsIdx+1 >= len(args) {
			return replyError("syntax error")
		}

		count, _ := strconv.Atoi(args[fieldsIdx+1])

		if !s.exists(key) {
			items := make([]string, 0, count)
			for i := 0; i < count; i++ {
				items = append(items, "-2")
			}

			return replyIntArray(items)
		}

		unit := "ex"
		if fieldsIdx >= 3 {
			unit = strings.ToLower(args[fieldsIdx-2])
		}

		var ttl time.Duration

		switch unit {
		case "px":
			ms, _ := strconv.ParseInt(args[fieldsIdx-1], 10, 64)
			ttl = time.Duration(ms) * time.Millisecond
		default:
			secs, _ := strconv.ParseInt(args[fieldsIdx-1], 10, 64)
			ttl = time.Duration(secs) * time.Second
		}

		s.expiry[key] = time.Now().Add(ttl)

		items := make([]string, 0, count)
		for i := 0; i < count; i++ {
			items = append(items, "1")
		}

		return replyIntArray(items)

	default:
		// CLIENT SETINFO and any other handshake chatter.
		return replyOK()
	}
}

func replyIntArray(items []string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "*%d\r\n", len(items))

	for _, item := range items {
		fmt.Fprintf(&b, ":%s\r\n", item)
	}

	return b.String()
}

func (s *fakeRedisServer) cmdSet(args []string) string {
	if len(args) < 3 {
		return replyError("wrong number of arguments for 'set' command")
	}

	key, val := args[1], args[2]

	var (
		nx, xx     bool
		expiration time.Duration
	)

	for i := 3; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "EX":
			if i+1 < len(args) {
				secs, err := strconv.ParseInt(args[i+1], 10, 64)
				if err != nil {
					return replyError("invalid expire time")
				}

				expiration = time.Duration(secs) * time.Second

				i++
			}
		case "PX":
			if i+1 < len(args) {
				ms, err := strconv.ParseInt(args[i+1], 10, 64)
				if err != nil {
					return replyError("invalid expire time")
				}

				expiration = time.Duration(ms) * time.Millisecond

				i++
			}
		case "NX":
			nx = true
		case "XX":
			xx = true
		}
	}

	existed := s.exists(key)

	if nx && existed {
		return replyNil()
	}

	if xx && !existed {
		return replyNil()
	}

	if existed {
		if bad, ok := s.wrongType(key, "string"); ok {
			return bad
		}

		s.drop(key)
	}

	s.strs[key] = val

	if expiration > 0 {
		s.expiry[key] = time.Now().Add(expiration)
	}

	return replyOK()
}

func (s *fakeRedisServer) cmdHSet(args []string) string {
	if len(args) < 4 || (len(args)-2)%2 != 0 {
		return replyError("wrong number of arguments for 'hset' command")
	}

	key := args[1]

	if s.exists(key) {
		if bad, ok := s.wrongType(key, "hash"); ok {
			return bad
		}
	}

	if _, ok := s.hashes[key]; !ok {
		s.hashes[key] = make(map[string]string)
	}

	var added int64

	for i := 2; i+1 < len(args); i += 2 {
		if _, ok := s.hashes[key][args[i]]; !ok {
			added++
		}

		s.hashes[key][args[i]] = args[i+1]
	}

	return replyInt(added)
}

// replyMixedArray renders an array where sentinel-prefixed items are emitted
// as RESP null bulk strings (MGET misses) while hits are written as bulks
// with the sentinel stripped.
func replyMixedArray(items []string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "*%d\r\n", len(items))

	for _, item := range items {
		if item == "\x00nil" {
			b.WriteString(replyNil())

			continue
		}

		b.WriteString(replyBulk(strings.TrimPrefix(item, "\x00v")))
	}

	return b.String()
}
