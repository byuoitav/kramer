package protocol3000

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
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

func (d *Device) GetAudioVideoInputs(ctx context.Context) (map[string]string, error) {
	var resp string
	cmd := []byte("#VID? *\n")

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

			r = bytes.TrimSpace(r)
			return fmt.Errorf("%s", r)
		}

		resp = string(r)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// response looks like: ~01@VID 2>1 ,2>2 ,2>3 ,2>4
	split := strings.Split(resp, "VID")
	if len(split) != 2 {
		return nil, fmt.Errorf("unexpected response: %q", resp)
	}

	inputs := make(map[string]string)

	// split[1] looks like: 2>1 ,2>2 ,2>3 ,2>4
	for _, input := range strings.Split(split[1], ",") {
		// input looks like: 2>1
		split := strings.Split(strings.TrimSpace(input), ">")
		if len(split) != 2 {
			return nil, fmt.Errorf("unexpected response: %q", resp)
		}

		inputs[split[1]] = split[0]
	}

	return inputs, nil
}

func (d *Device) SetAudioVideoInput(ctx context.Context, output, input string) error {
	return errors.New("not implemented")
}
