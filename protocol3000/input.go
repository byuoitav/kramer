package protocol3000

import (
	"context"
	"fmt"
	"strings"
)

func (d *Device) AudioVideoInputs(ctx context.Context) (map[string]string, error) {
	resp, err := d.sendCommand(ctx, []byte("#VID? *\n"))
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
	cmd := []byte(fmt.Sprintf("#VID %s>%s\n", input, output))

	resp, err := d.sendCommand(ctx, cmd)
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
