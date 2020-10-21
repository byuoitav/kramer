package protocol3000

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/byuoitav/connpool"
)

const (
	asciiLineFeed = 0x0d
)

type Device struct {
	pool *connpool.Pool
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
	}

	if options.logger != nil {
		dev.pool.Logger = options.logger.Sugar()
	}

	return dev
}
