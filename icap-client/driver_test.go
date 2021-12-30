package icapclient

import (
	"context"
	"testing"
	"time"
)

func TestDriver(t *testing.T) {
	if !testServerRunning() {
		go startTestServer()
	}

	t.Run("Driver Connect With Context", func(t *testing.T) {
		driver := &Driver{
			Host: "127.0.0.1",
			Port: 1344,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := driver.ConnectWithContext(ctx); err != nil {
			t.Log("Driver connection with context failed: ", err.Error())
			t.Fail()
			return
		}

		if err := driver.Close(); err != nil {
			t.Log("Driver connection close failed: ", err.Error())
			t.Fail()
		}

	})

	if testServerRunning() {
		defer stopTestServer()
	}

}
