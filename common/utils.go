package common

import (
	"encoding/json"

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

func MustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	PanicIfError(err)
	return b
}

func MustMarshalString(v interface{}) string {
	return string(MustMarshal(v))
}

func Map2Obj(m map[string]interface{}, obj interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, obj)
}
