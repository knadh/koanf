//go:build go1.19
// +build go1.19

package nats

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/logger"
	"github.com/nats-io/nats-server/v2/server"
)

var testNatsURL string

func TestMain(m *testing.M) {
	gnatsd, err := server.NewServer(&server.Options{
		Port:      server.RANDOM_PORT,
		JetStream: true,
	})
	if err != nil {
		log.Fatal("failed to create gnatsd server")
	}
	gnatsd.SetLogger(
		logger.NewStdLogger(false, false, false, false, false),
		false,
		false,
	)
	go gnatsd.Start()
	defer gnatsd.Shutdown()

	if !gnatsd.ReadyForConnections(time.Second) {
		log.Fatal("failed to start the gnatsd server")
	}
	testNatsURL = "nats://" + gnatsd.Addr().String()

	os.Exit(m.Run())
}
