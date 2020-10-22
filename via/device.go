package via

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"regexp"
	"time"

	"go.uber.org/zap"
)

const (
	viaReboot = "Reboot"
	viaReset  = "Reset"
)

// VIA Struct that defines general parameters needed for any VIA
type Via struct {
	Address  string
	Username string
	Password string
	Log      *zap.Logger
}

func getConnection(address string) (*net.TCPConn, error) {
	radder, err := net.ResolveTCPAddr("tcp", address+":9982")
	if err != nil {
		err = fmt.Errorf("error resolving address : %s", err.Error())
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, radder)
	if err != nil {
		err = fmt.Errorf("error dialing address : %s", err.Error())
		return nil, err
	}

	return conn, nil
}

// SendCommand opens a connection with <addr> and sends the <command> to the via, returning the response from the via, or an error if one occured.
func (v *Via) sendCommand(ctx context.Context, cmd command) (string, error) {
	// get the connection
	v.Log.Info("Opening telnet connection with", zap.String("address", v.Address))
	conn, err := getConnection(v.Address)
	if err != nil {
		return "", err
	}

	timeoutDuration := 7 * time.Second

	// Set Read Connection Duration
	conn.SetReadDeadline(time.Now().Add(timeoutDuration))

	// login
	err = v.login(ctx, conn)
	if err != nil {
		v.Log.Error("Houston, we have a problem logging in. The login failed", zap.Error(err))
		return "", err
	}

	// write command
	if len(cmd.Command) > 0 {
		cmd.Username = v.Username
		b, err := xml.Marshal(cmd)
		if err != nil {
			return "", err
		}

		_, err = conn.Write(b)
		if err != nil {
			return "", err
		}
	}

	reader := bufio.NewReader(conn)
	resp, err := reader.ReadBytes('\n')
	if err != nil {
		v.Log.Error("error reading from system", zap.Error(err))
		return "", err
	}

	if len(string(resp)) > 0 {
		v.Log.Info("Response from device", zap.String("resp", string(resp)))
	}

	return string(resp), nil
}

func (v *Via) login(ctx context.Context, conn *net.TCPConn) error {
	var cmd command

	cmd.Username = v.Username
	cmd.Password = v.Password
	cmd.Command = "Login"

	// read welcome message (Only Important when we first open a connection and login)
	reader := bufio.NewReader(conn)
	_, err := reader.ReadBytes('\n')
	if err != nil {
		v.Log.Error("error reading from system", zap.Error(err))
		return err
	}

	v.Log.Info("Logging in...")
	v.Log.Debug("Username", zap.String("username", v.Username))

	b, err := xml.Marshal(cmd)
	if err != nil {
		return err
	}

	_, err = conn.Write(b)
	if err != nil {
		return err
	}

	resp, err := reader.ReadBytes('\n')
	if err != nil {
		err = fmt.Errorf("error reading from system", err.Error())
		v.Log.Error("error reading from system", zap.Error(err))
		return err
	}

	s := string(resp)

	errRx := regexp.MustCompile(`Error`)
	SuccessRx := regexp.MustCompile(`Successful`)
	respRx := errRx.MatchString(s)
	SuccessResp := SuccessRx.MatchString(s)

	if respRx == true {
		v.Log.Info("Response from device", zap.String("response", s))
		return fmt.Errorf("Unable to login due to an error: %s", s)
	}

	if SuccessResp == true {
		v.Log.Debug("Connection is successful, We are connected", zap.String("response", s))
	}

	return nil
}
