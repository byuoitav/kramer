package protocol3000

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byuoitav/connpool"
)

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
	var resp string
	cmd := []byte(fmt.Sprintf("#VID %s>%s\n", input, output))

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
		return err
	}

	// response looks like: ~01@VID 1>2
	split := strings.Split(resp, "VID")
	if len(split) != 2 {
		return fmt.Errorf("unexpected response: %q", resp)
	}

	mapping := strings.Split(strings.TrimSpace(split[1]), ">")
	switch {
	case len(mapping) != 2:
		return fmt.Errorf("unexpected response: %q", resp)
	case mapping[0] != input || mapping[1] != output:
		return fmt.Errorf("unexpected response: %q", resp)
	}

	return nil
}
