// Package redisclient is a minimal, dependency-free Redis client speaking
// RESP2 over a pooled set of TCP connections. StreamPass only needs a
// handful of commands (SET with TTL, GET, DEL, EXISTS, SCAN), so a full
// third-party client was not vendored — the same dependency-free rationale
// used for shared/config's YAML parser and the backend's JWT
// implementation (KISS/YAGNI).
package redisclient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	apperrors "streampass/shared/errors"
)

// Client is a minimal Redis client with a small connection pool.
type Client struct {
	addr     string
	password string
	dialTO   time.Duration

	mu   sync.Mutex
	pool []*conn
}

type conn struct {
	nc net.Conn
	r  *bufio.Reader
}

// Config configures the Redis client.
type Config struct {
	Addr        string
	Password    string
	DialTimeout time.Duration
}

// New builds a Client. Connections are opened lazily and pooled on first
// use, up to maxPoolSize.
func New(cfg Config) *Client {
	dialTO := cfg.DialTimeout
	if dialTO == 0 {
		dialTO = 5 * time.Second
	}
	return &Client{addr: cfg.Addr, password: cfg.Password, dialTO: dialTO}
}

// maxPoolSize bounds the idle connection pool; MVP traffic never needs
// more concurrent Redis connections than this.
const maxPoolSize = 16

func (c *Client) getConn(ctx context.Context) (*conn, error) {
	c.mu.Lock()
	if n := len(c.pool); n > 0 {
		cn := c.pool[n-1]
		c.pool = c.pool[:n-1]
		c.mu.Unlock()
		return cn, nil
	}
	c.mu.Unlock()

	d := net.Dialer{Timeout: c.dialTO}
	nc, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeUnavailable, "failed to connect to redis", err)
	}
	cn := &conn{nc: nc, r: bufio.NewReader(nc)}

	if c.password != "" {
		if _, err := cn.do("AUTH", c.password); err != nil {
			nc.Close()
			return nil, apperrors.Wrap(apperrors.CodeUnavailable, "redis auth failed", err)
		}
	}
	return cn, nil
}

func (c *Client) release(cn *conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.pool) < maxPoolSize {
		c.pool = append(c.pool, cn)
		return
	}
	cn.nc.Close()
}

func (c *Client) discard(cn *conn) {
	cn.nc.Close()
}

// Set stores key=value with an optional TTL (0 means no expiry).
func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	cn, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	var reply respValue
	if ttl > 0 {
		reply, err = cn.do("SET", key, value, "PX", strconv.FormatInt(ttl.Milliseconds(), 10))
	} else {
		reply, err = cn.do("SET", key, value)
	}
	if err != nil {
		c.discard(cn)
		return apperrors.Wrap(apperrors.CodeUnavailable, "redis SET failed", err)
	}
	c.release(cn)
	if reply.str != "OK" {
		return apperrors.New(apperrors.CodeUnavailable, "unexpected redis SET reply")
	}
	return nil
}

// Get retrieves a key's value. Returns apperrors.CodeNotFound if absent.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	cn, err := c.getConn(ctx)
	if err != nil {
		return "", err
	}
	reply, err := cn.do("GET", key)
	if err != nil {
		c.discard(cn)
		return "", apperrors.Wrap(apperrors.CodeUnavailable, "redis GET failed", err)
	}
	c.release(cn)
	if reply.isNil {
		return "", apperrors.New(apperrors.CodeNotFound, "redis key not found")
	}
	return reply.str, nil
}

// Del deletes a key. Not finding the key is not an error.
func (c *Client) Del(ctx context.Context, key string) error {
	cn, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	_, err = cn.do("DEL", key)
	if err != nil {
		c.discard(cn)
		return apperrors.Wrap(apperrors.CodeUnavailable, "redis DEL failed", err)
	}
	c.release(cn)
	return nil
}

// Exists reports whether a key is present.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	cn, err := c.getConn(ctx)
	if err != nil {
		return false, err
	}
	reply, err := cn.do("EXISTS", key)
	if err != nil {
		c.discard(cn)
		return false, apperrors.Wrap(apperrors.CodeUnavailable, "redis EXISTS failed", err)
	}
	c.release(cn)
	return reply.str == "1", nil
}

// DelPrefix scans and deletes every key matching prefix* (used for
// RevokeAll). MVP-scale key counts make a SCAN+DEL loop acceptable; if the
// session count per user grows large this should switch to a Redis SET
// tracking a user's active session keys instead.
func (c *Client) DelPrefix(ctx context.Context, prefix string) error {
	cn, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	defer c.release(cn)

	cursor := "0"
	for {
		reply, err := cn.do("SCAN", cursor, "MATCH", prefix+"*", "COUNT", "100")
		if err != nil {
			c.discard(cn)
			return apperrors.Wrap(apperrors.CodeUnavailable, "redis SCAN failed", err)
		}
		if len(reply.array) != 2 {
			return apperrors.New(apperrors.CodeUnavailable, "unexpected redis SCAN reply shape")
		}
		cursor = reply.array[0].str
		for _, keyVal := range reply.array[1].array {
			if _, err := cn.do("DEL", keyVal.str); err != nil {
				c.discard(cn)
				return apperrors.Wrap(apperrors.CodeUnavailable, "redis DEL failed during scan", err)
			}
		}
		if cursor == "0" {
			break
		}
	}
	return nil
}

// respValue is a parsed RESP2 reply: exactly one of str or array is
// meaningful, depending on isArray/isNil.
type respValue struct {
	str     string
	array   []respValue
	isArray bool
	isNil   bool
}

// do sends a command and returns its parsed reply.
func (cn *conn) do(args ...string) (respValue, error) {
	if err := cn.writeCommand(args); err != nil {
		return respValue{}, err
	}
	return cn.readValue()
}

func (cn *conn) writeCommand(args []string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(a), a)
	}
	_, err := cn.nc.Write([]byte(b.String()))
	return err
}

func (cn *conn) readLine() (string, error) {
	line, err := cn.r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// readValue recursively parses one RESP2 value of any type (simple string,
// error, integer, bulk string, or array).
func (cn *conn) readValue() (respValue, error) {
	line, err := cn.readLine()
	if err != nil {
		return respValue{}, err
	}
	if len(line) == 0 {
		return respValue{}, errors.New("redis: empty reply line")
	}

	switch line[0] {
	case '+', ':': // simple string / integer
		return respValue{str: line[1:]}, nil
	case '-': // error
		return respValue{}, errors.New("redis: " + line[1:])
	case '$': // bulk string
		n, err := strconv.Atoi(line[1:])
		if err != nil {
			return respValue{}, err
		}
		if n == -1 {
			return respValue{isNil: true}, nil
		}
		buf := make([]byte, n+2) // +2 for trailing \r\n
		if _, err := io.ReadFull(cn.r, buf); err != nil {
			return respValue{}, err
		}
		return respValue{str: string(buf[:n])}, nil
	case '*': // array
		n, err := strconv.Atoi(line[1:])
		if err != nil {
			return respValue{}, err
		}
		if n == -1 {
			return respValue{isArray: true, isNil: true}, nil
		}
		items := make([]respValue, n)
		for i := 0; i < n; i++ {
			v, err := cn.readValue()
			if err != nil {
				return respValue{}, err
			}
			items[i] = v
		}
		return respValue{isArray: true, array: items}, nil
	default:
		return respValue{}, fmt.Errorf("redis: unsupported reply prefix %q", line[0])
	}
}
