package test

import (
	"testing"

	"github.com/dtm-labs/dtm/client/dtmcli"
)

func TestSagaCover(t *testing.T) {
	dtmcli.SetPassthroughHeaders([]string{})
}
