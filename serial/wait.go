package serial

import (
	"fmt"
	"os"
	"time"
)

// WaitForSerial polls for the presence of the given serial port path until it
// becomes available or the timeout expires. It returns an error if the timeout
// is reached before the device exists.
func WaitForSerial(port string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(port); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("serial port %s not found after %v", port, timeout)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
