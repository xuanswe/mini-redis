package redis_test

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xuanswe/mini-redis/internal/resp/consts/commands"
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
	conn           net.Conn
	name           string
	waitBeforeSend time.Duration
	command        string
	want           string
	readTimeout    time.Duration
}

func TestServer_ValidCommand_ShouldGetValidResponse(t *testing.T) {
	redisServer := prepareServer(1 * time.Second)
	defer redisServer.Shutdown()

	for i := range 5 {
		nClients := i + 1
		t.Run(fmt.Sprintf("%d clients", nClients), func(t *testing.T) {
			conns := make([]net.Conn, nClients)
			for i := range conns {
				conn, _ := net.Dial("tcp", "localhost:"+redisServer.Config().Port)
				conns[i] = conn
			}
			defer func() {
				for _, conn := range conns {
					conn.Close()
				}
			}()

			var testRequests = make([]testRequest, 0)
			appendForAllClients := func(command string, want string) {
				for j, conn := range conns {
					testRequests = append(testRequests, testRequest{
						conn:    conn,
						name:    fmt.Sprintf("client %d: %q", j, command),
						command: command,
						want:    want,
					})
				}
			}

			appendForAllClients("*1\r\n$4\r\nPING\r\n", "+PONG\r\n")
			appendForAllClients("*2\r\n$4\r\nPING\r\n$7\r\nMessage\r\n", "+Message\r\n")
			appendForAllClients("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n", "$3\r\nhey\r\n")
			appendForAllClients("*2\r\n$3\r\nGET\r\n$2\r\nkb\r\n", "$-1\r\n")

			testRequests = append(testRequests, testRequest{
				conn:    conns[0],
				name:    fmt.Sprintf("client %d: SET ka=va", 0),
				command: "*3\r\n$3\r\nSET\r\n$2\r\nka\r\n$2\r\nva\r\n",
				want:    "+OK\r\n",
			})

			appendForAllClients("*2\r\n$3\r\nGET\r\n$2\r\nka\r\n", "$2\r\nva\r\n")

			testRequests = append(testRequests, testRequest{
				conn:    conns[0],
				name:    fmt.Sprintf("client %d: SET kb=vb, expired in 100ms", 0),
				command: "*5\r\n$3\r\nSET\r\n$2\r\nkb\r\n$2\r\nvb\r\n$2\r\nPX\r\n$3\r\n100\r\n",
				want:    "+OK\r\n",
			})

			appendForAllClients("*2\r\n$3\r\nGET\r\n$2\r\nkb\r\n", "$2\r\nvb\r\n")

			n := len(testRequests)
			appendForAllClients("*2\r\n$3\r\nGET\r\n$2\r\nkb\r\n", "$-1\r\n")
			testRequests[n].waitBeforeSend = 100 * time.Millisecond

			for _, tr := range testRequests {
				t.Run(tr.name, func(t *testing.T) {
					verifyTestRequest(t, tr)
				})
			}
		})
	}
}

func TestServer_PartialCommandTimeout_ShouldCloseConnectionFromServerSide(t *testing.T) {
	redisServer := prepareServer(500 * time.Millisecond)
	defer redisServer.Shutdown()

	conn, _ := net.Dial("tcp", "localhost:"+redisServer.Config().Port)
	defer conn.Close()

	partialCommand := "*2\r\n$4\r\nPING\r\n"
	conn.Write([]byte(partialCommand))

	// wait until the server closes the connection
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := io.ReadFull(conn, make([]byte, 1))
	log.Debug().Msgf("n: %d, err: %v", n, err)
	if err != io.EOF {
		t.Errorf("got: %v, want: %v", err, io.EOF)
	}
}

func TestServer_CommandIsNotAnArray_ShouldCloseConnectionFromServerSide(t *testing.T) {
	redisServer := prepareServer(1 * time.Second)
	defer redisServer.Shutdown()

	conn, _ := net.Dial("tcp", "localhost:"+redisServer.Config().Port)
	defer conn.Close()

	nonArrayCommand := "$4\r\nPING\r\n"
	conn.Write([]byte(nonArrayCommand))

	// wait until the server closes the connection
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := io.ReadFull(conn, make([]byte, 1))
	log.Debug().Msgf("n: %d, err: %v", n, err)
	if err != io.EOF {
		t.Errorf("got: %v, want: %v", err, io.EOF)
	}
}

func TestServer_InvalidCommand_ShouldGetErrorResponse(t *testing.T) {
	redisServer := prepareServer(1 * time.Second)
	defer redisServer.Shutdown()

	var testRequests = []testRequest{
		{
			name:    "Unknown command: UKCMD",
			command: "*1\r\n$5\r\nUKCMD\r\n",
			want:    "-ERR unsupported command UKCMD\r\n",
		},
		{
			name:    "PING: 2 arguments",
			command: "*3\r\n$4\r\nPING\r\n$3\r\nMSG\r\n$3\r\nMSG\r\n",
			want:    fmt.Sprintf("-ERR wrong number of arguments for '%s' command\r\n", commands.Ping),
		},
		{
			name:    "ECHO: no arguments",
			command: "*1\r\n$4\r\nECHO\r\n",
			want:    fmt.Sprintf("-ERR wrong number of arguments for '%s' command\r\n", commands.Echo),
		},
		{
			name:    "ECHO: 2 arguments",
			command: "*3\r\n$4\r\nECHO\r\n$3\r\nMSG\r\n$3\r\nMSG\r\n",
			want:    fmt.Sprintf("-ERR wrong number of arguments for '%s' command\r\n", commands.Echo),
		},
	}

	for _, tr := range testRequests {
		t.Run(tr.name, func(t *testing.T) {
			conn, _ := net.Dial("tcp", "localhost:"+redisServer.Config().Port)
			defer conn.Close()

			tr.conn = conn
			verifyTestRequest(t, tr)
		})
	}
}

func verifyTestRequest(t *testing.T, tr testRequest) {
	conn := tr.conn

	// Wait
	if tr.waitBeforeSend > 0 {
		time.Sleep(tr.waitBeforeSend)
	}

	// Send request
	conn.Write([]byte(tr.command))

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

func prepareServer(ConnIdleTimeout time.Duration) redis.ServerInterface {
	port := "6379"
	redisServer, _ := redis.NewServer(redis.ServerConfig{
		Host:            "0.0.0.0",
		Port:            port,
		ConnIdleTimeout: ConnIdleTimeout,
	})
	go redisServer.Start()

	// TODO: wait for server ready event, don't guess to sleep
	// time.Sleep(3 * time.Second)
	return redisServer
}
