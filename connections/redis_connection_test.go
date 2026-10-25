package sdk_con

import (
	"bufio"
	"context"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

// startMiniRedis runs a RESP2 server that answers PING with +PONG and every
// other command with +OK. It is enough to drive redisClient's happy path
// (handshake + ping) without a real Redis.
func startMiniRedis(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	var wg sync.WaitGroup

	var connMu sync.Mutex

	var conns []net.Conn

	closeAll := func() {
		connMu.Lock()
		defer connMu.Unlock()

		for _, c := range conns {
			_ = c.Close()
		}

		conns = nil
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			connMu.Lock()
			conns = append(conns, conn)
			connMu.Unlock()

			wg.Add(1)

			go func() {
				defer wg.Done()
				defer conn.Close()

				r := bufio.NewReader(conn)
				w := bufio.NewWriter(conn)

				for {
					cmd, err := readMiniCommand(r)
					if err != nil {
						return
					}

					if len(cmd) > 0 && strings.EqualFold(cmd[0], "PING") {
						_, _ = w.WriteString("+PONG\r\n")
					} else if len(cmd) > 0 && strings.EqualFold(cmd[0], "HELLO") {
						// Force the client onto the RESP2 code path.
						_, _ = w.WriteString("-ERR unknown command 'HELLO'\r\n")
					} else {
						_, _ = w.WriteString("+OK\r\n")
					}

					if err := w.Flush(); err != nil {
						return
					}
				}
			}()
		}
	}()

	t.Cleanup(func() {
		_ = ln.Close()
		closeAll()
		wg.Wait()
	})

	return ln.Addr().String()
}

func readMiniCommand(r *bufio.Reader) ([]string, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	if prefix != '*' {
		return nil, io.ErrUnexpectedEOF
	}

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	n, err := strconv.Atoi(strings.TrimRight(line, "\r\n"))
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, n)

	for i := 0; i < n; i++ {
		if _, err := r.ReadByte(); err != nil { // '$'
			return nil, err
		}

		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		length, err := strconv.Atoi(strings.TrimRight(line, "\r\n"))
		if err != nil {
			return nil, err
		}

		buf := make([]byte, length+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}

		args = append(args, string(buf[:length]))
	}

	return args, nil
}

func TestRedisClientHappyPathAgainstFakeServer(t *testing.T) {
	addr := startMiniRedis(t)

	client, err := RedisConnection(context.Background(), &sdk_dto.RedisClientOptions{
		Addr:        "redis://" + addr,
		Protocol:    2,
		PoolFIFO:    true,
		MaxRetries:  1,
		DialTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("RedisConnection: %v", err)
	}

	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestRedisClientParseURLFailure(t *testing.T) {
	_, err := RedisConnection(context.Background(), &sdk_dto.RedisClientOptions{
		Addr: "not a valid url",
	})
	if err == nil {
		t.Fatal("expected ParseURL error")
	}
}

func TestRedisSentinelClientFailure(t *testing.T) {
	// Nothing listens on port 1; the failover client must surface the ping
	// failure instead of hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := RedisConnection(ctx, &sdk_dto.RedisClientOptions{
		Addr:          "redis://127.0.0.1:6379",
		MasterName:    "primary",
		SentinelAddrs: []string{"127.0.0.1:1"},
		MaxRetries:    1,
		DialTimeout:   300 * time.Millisecond,
		ReadTimeout:   300 * time.Millisecond,
		WriteTimeout:  300 * time.Millisecond,
		PoolTimeout:   300 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected sentinel connection error")
	}
}

func TestRabbitConnectionClusterRequiresURLs(t *testing.T) {
	_, err := rabbitConnectionCluster(&sdk_dto.RabbitClientOptions{})
	if err == nil {
		t.Fatal("expected error for empty cluster urls")
	}
}

func TestRabbitConnectionClusterDialFailure(t *testing.T) {
	_, err := rabbitConnectionCluster(&sdk_dto.RabbitClientOptions{
		Urls:    []string{"amqp://guest:guest@127.0.0.1:1"},
		Timeout: 1,
	})
	if err == nil {
		t.Fatal("expected dial error for unreachable cluster")
	}
}
