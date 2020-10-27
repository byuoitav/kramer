package protocol3000

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/byuoitav/connpool"
	"go.uber.org/zap"
)

const (
	asciiLineFeed = 0x0d
)

type Device struct {
	pool *connpool.Pool
	Log  *zap.Logger
}

func New(addr string, opts ...Option) *Device {
	options := &options{
		ttl:   _defaultTTL,
		delay: _defaultDelay,
	}

	for _, o := range opts {
		o.apply(options)
	}

	dev := &Device{
		pool: &connpool.Pool{
			TTL:   options.ttl,
			Delay: options.delay,
			NewConnection: func(ctx context.Context) (net.Conn, error) {
				dial := net.Dialer{}

				conn, err := dial.DialContext(ctx, "tcp", addr+":5000")
				if err != nil {
					return nil, err
				}

				deadline, ok := ctx.Deadline()
				if !ok {
					deadline = time.Now().Add(5 * time.Second)
				}

				conn.SetDeadline(deadline)

				// read the first 'welcome' line from the connection
				buf := make([]byte, 64)
				for !bytes.Contains(buf, []byte{asciiLineFeed}) {
					_, err := conn.Read(buf)
					if err != nil {
						conn.Close()
						return nil, fmt.Errorf("unable to read welcome line: %w", err)
					}
				}

				return conn, nil
			},
		},
		Log: options.logger,
	}

	if options.logger != nil {
		dev.pool.Logger = options.logger.Sugar()
	}

	return dev
}

func (d *Device) sendCommand(ctx context.Context, cmd []byte) (string, error) {
	var str string

	err := d.pool.Do(ctx, func(conn connpool.Conn) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			deadline = time.Now().Add(10 * time.Second)
		}

		conn.SetDeadline(deadline)

		n, err := conn.Write(cmd)
		switch {
		case err != nil:
			return fmt.Errorf("unable to write command: %w", err)
		case n != len(cmd):
			return fmt.Errorf("unable to write command: wrote %v/%v bytes", n, len(cmd))
		}

		r, err := conn.ReadUntil(asciiLineFeed, deadline)
		if err != nil {
			return fmt.Errorf("unable to read response: %w", err)
		}

		r = bytes.TrimSpace(r)
		if len(r) == 0 {
			// read the next line, where the error is
			r, err = conn.ReadUntil(asciiLineFeed, deadline)
			if err != nil {
				return fmt.Errorf("unable to read error: %w", err)
			}

			// parse the error
			r = bytes.TrimSpace(r)
			resp := string(r)
			if !strings.HasPrefix(resp, "ERR ") {
				return errors.New(string(r))
			}

			code, err := strconv.Atoi(strings.TrimPrefix(resp, "ERR "))
			if err != nil {
				return errors.New(string(r))
			}

			return errors.New(Error(code))
		}

		str = string(r)
		return nil
	})
	if err != nil {
		return "", err
	}

	return str, nil
}
