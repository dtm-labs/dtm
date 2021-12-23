package common

import (
	"errors"
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

func TestCheckConfig(t *testing.T) {
	config := &Config
	retryIntervalErr := checkConfig()
	retryIntervalExpect := errors.New("RetryInterval should not be less than 10")
	assert.Equal(t, retryIntervalErr, retryIntervalExpect)

	config.RetryInterval = 10
	timeoutToFailErr := checkConfig()
	timeoutToFailExpect := errors.New("TimeoutToFail should not be less than RetryInterval")
	assert.Equal(t, timeoutToFailErr, timeoutToFailExpect)

	config.TimeoutToFail = 20
	driverErr := checkConfig()
	assert.Equal(t, driverErr, nil)

	config.Store = Store{Driver: Mysql}
	hostErr := checkConfig()
	hostExpect := errors.New("Db host not valid ")
	assert.Equal(t, hostErr, hostExpect)

	config.Store = Store{Driver: Mysql, Host: "127.0.0.1"}
	portErr := checkConfig()
	portExpect := errors.New("Db port not valid ")
	assert.Equal(t, portErr, portExpect)

	config.Store = Store{Driver: Mysql, Host: "127.0.0.1", Port: 8686}
	userErr := checkConfig()
	userExpect := errors.New("Db user not valid ")
	assert.Equal(t, userErr, userExpect)
}
