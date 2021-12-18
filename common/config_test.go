package common

import (
	"os"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestLoadFromEnv(t *testing.T) {
	assert.Equal(t, "MICRO_SERVICE_DRIVER", toUnderscoreUpper("MicroService_Driver"))

	ms := MicroService{}
	os.Setenv("T_DRIVER", "d1")
	loadFromEnv("T", &ms)
	assert.Equal(t, "d1", ms.Driver)
}
