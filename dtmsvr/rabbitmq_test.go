package dtmsvr

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties/assert"
	"github.com/sirupsen/logrus"
	"github.com/yedf/dtm/common"
)

func init() {
	LoadConfig()
}

func TestRabbitConfig(t *testing.T) {
	assert.Matches(t, ServerConfig.Rabbitmq.KeyCommited, "key_committed")
}

func TestRabbitmq1Msg(t *testing.T) {
	rb := RabbitmqNew(&ServerConfig.Rabbitmq)
	err := rb.SendAndConfirm(RabbitmqConstPrepared, gin.H{
		"gid": common.GenGid(),
	})
	assert.Equal(t, nil, err)
	queue := rb.QueueNew(RabbitmqConstPrepared)
	queue.WaitAndHandle(func(data map[string]interface{}) {
		logrus.Printf("processed msg: %v in queue1", data)
	})
	assert.Equal(t, 0, 1)
}
