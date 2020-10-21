package protocol3000

import (
	"context"
	"fmt"
	"strings"
)

func (d *Device) Health(ctx context.Context) error {
	resp, err := d.sendCommand(ctx, []byte("#VID? *\n"))
	if err != nil {
		return err
	}

	if !strings.Contains(resp, "OK") {
		return fmt.Errorf("unexpected response: %q", resp)
	}

	return nil
}
