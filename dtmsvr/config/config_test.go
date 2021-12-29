package config

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestConfig(t *testing.T) {
	testConfigStringField(&Config.Store.Driver, "", t)
	testConfigStringField(&Config.Store.User, "", t)
	testConfigIntField(&Config.RetryInterval, 9, t)
	testConfigIntField(&Config.TimeoutToFail, 9, t)
}

func testConfigStringField(fd *string, val string, t *testing.T) {
	old := *fd
	*fd = val
	str := checkConfig()
	assert.NotEqual(t, "", str)
	*fd = old
}

func testConfigIntField(fd *int64, val int64, t *testing.T) {
	old := *fd
	*fd = val
	str := checkConfig()
	assert.NotEqual(t, "", str)
	*fd = old
}
