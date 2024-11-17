package redis_test

import (
	"fmt"
	"github.com/xuanswe/mini-redis/internal/support"
	"github.com/xuanswe/mini-redis/redis"
	"io"
	"net"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	support.SetupLogger()
	os.Exit(m.Run())
}

type testRequest struct {
	conn        net.Conn
	name        string
	data        string
	want        string
	readTimeout time.Duration
}

func TestServer(t *testing.T) {
	redisServer := prepareServer(t, 1*time.Second)
	defer redisServer.Shutdown()

	t.Run("ParallelClients", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			t.Run(fmt.Sprintf("Client %d", i), func(t *testing.T) {
				t.Parallel()

				conn, _ := net.Dial("tcp", "localhost:"+redisServer.Config().Port)
				defer conn.Close()

				for j := 0; j < 3; j++ {
					tr := testRequest{
						conn: conn,
						name: fmt.Sprintf("World %d", j),
						data: fmt.Sprintf("World %d", j),
						want: fmt.Sprintf("Hello World %d!", j),
					}

					t.Run(tr.name, func(t *testing.T) {
						verifyTestRequest(t, tr)
					})
				}
			})
		}
	})
}

func verifyTestRequest(t *testing.T, tr testRequest) {
	t.Helper()

	conn := tr.conn

	// Send request
	conn.Write([]byte(tr.data + "\n"))

	readTimeout := tr.readTimeout
	if readTimeout <= 0 {
		readTimeout = 1 * time.Second
	}

	// Read response
	actual := make([]byte, len(tr.want))
	errChan := make(chan error, 1)
	go func() {
		_, err := io.ReadFull(conn, actual)
		errChan <- err
	}()

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("error reading response: %s", err)
		}

		if string(actual) != tr.want {
			t.Errorf("actual %q, want: %q", actual, tr.want)
		}
	case <-time.After(readTimeout):
		t.Errorf("timeout reading response")
	}
}

func prepareServer(t *testing.T, ConnIdleTimeout time.Duration) redis.ServerInterface {
	t.Helper()

	port := "6379"
	redisServer, _ := redis.NewServer(redis.ServerConfig{
		Host:            "0.0.0.0",
		Port:            port,
		ConnIdleTimeout: ConnIdleTimeout,
	})
	if err := redisServer.Start(); err != nil {
		t.Fatalf("Error starting Redis server: %s", err)
	}

	return redisServer
}
