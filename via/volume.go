package via

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"go.uber.org/zap"
)

// Volumes: opening a connection with the VIAs and then return the volume for the device
func (v *Via) Volumes(ctx context.Context, block []string) (map[string]int, error) {
	toReturn := make(map[string]int)
	var cmd command
	cmd.Command = "Vol"
	cmd.Param1 = "Get"

	v.Log.Info("Sending command to get VIA Volume", zap.String("address", v.Address))
	// Note: Volume Get command in VIA API doesn't have any error handling so it only returns Vol|Get|XX or nothing
	// Checking for errors during execution of command
	vollevel, err := v.sendCommand(ctx, cmd)
	if err != nil {
		v.Log.Error("Failed to get volume", zap.Error(err))
		return nil, err
	}

	volume, err := v.volumeParse(vollevel)
	if err != nil {
		v.Log.Error("failed to parse volume response", zap.Error(err))
		return nil, err
	}

	toReturn[""] = volume
	return toReturn, nil
}

// volumeParse parser to pull out the volume level from the VIA API returned string
func (v *Via) volumeParse(vollevel string) (int, error) {
	re := regexp.MustCompile("[0-9]+")
	vol := re.FindString(vollevel)

	vfin, err := strconv.Atoi(vol)
	if err != nil {
		v.Log.Error("error converting response", zap.Error(err))
		return 0, err
	}

	return vfin, nil
}

// SetVolume: Connect and set the volume on the VIA
func (v *Via) SetVolume(ctx context.Context, block string, volume int) error {
	var cmd command
	cmd.Command = "Vol"
	cmd.Param1 = "Set"
	cmd.Param2 = strconv.Itoa(volume)

	v.Log.Info("Sending volume set command to %s", zap.String("address", v.Address))

	_, err := v.sendCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("Error in setting volume on %s", v.Address)
	}

	return nil

}
