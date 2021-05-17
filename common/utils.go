package common

import (
	"github.com/bwmarrin/snowflake"
)

var gNode *snowflake.Node = nil

func GenGid() string {
	return gNode.Generate().Base58()
}

func init() {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	gNode = node
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func If(condition bool, trueObj interface{}, falseObj interface{}) interface{} {
	if condition {
		return trueObj
	}
	return falseObj
}
