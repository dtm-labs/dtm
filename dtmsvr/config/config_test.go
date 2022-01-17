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

func TestLoadConfig(t *testing.T) {
	MustLoadConfig("../../conf.sample.yml")
}
func TestCheckConfig(t *testing.T) {
	conf := Config
	conf.RetryInterval = 1
	retryIntervalErr := checkConfig(&conf)
	retryIntervalExpect := errors.New("RetryInterval should not be less than 10")
	assert.Equal(t, retryIntervalErr, retryIntervalExpect)

	conf.RetryInterval = 10
	conf.TimeoutToFail = 5
	timeoutToFailErr := checkConfig(&conf)
	timeoutToFailExpect := errors.New("TimeoutToFail should not be less than RetryInterval")
	assert.Equal(t, timeoutToFailErr, timeoutToFailExpect)

	conf.TimeoutToFail = 20
	driverErr := checkConfig(&conf)
	assert.Equal(t, driverErr, nil)

	conf.Store = Store{Driver: Mysql}
	hostErr := checkConfig(&conf)
	hostExpect := errors.New("Db host not valid ")
	assert.Equal(t, hostErr, hostExpect)

	conf.Store = Store{Driver: Mysql, Host: "127.0.0.1"}
	portErr := checkConfig(&conf)
	portExpect := errors.New("Db port not valid ")
	assert.Equal(t, portErr, portExpect)

	conf.Store = Store{Driver: Mysql, Host: "127.0.0.1", Port: 8686}
	userErr := checkConfig(&conf)
	userExpect := errors.New("Db user not valid ")
	assert.Equal(t, userErr, userExpect)

	conf.Store = Store{Driver: Redis, Host: "", Port: 8686}
	assert.Equal(t, errors.New("Redis host not valid"), checkConfig(&conf))

	conf.Store = Store{Driver: Redis, Host: "127.0.0.1", Port: 0}
	assert.Equal(t, errors.New("Redis port not valid"), checkConfig(&conf))

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
	str := checkConfig(&Config)
	assert.NotEqual(t, "", str)
	*fd = old
}

func testConfigIntField(fd *int64, val int64, t *testing.T) {
	old := *fd
	*fd = val
	str := checkConfig(&Config)
	assert.NotEqual(t, "", str)
	*fd = old
}
