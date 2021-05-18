package common

import (
	"encoding/json"

	"github.com/bwmarrin/snowflake"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
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

var RestyClient = resty.New()

func init() {
	RestyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		logrus.Printf("requesting: %s %s %v", r.Method, r.URL, r.Body)
		return nil
	})
	RestyClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		r := resp.Request
		logrus.Printf("requested: %s %s %s", r.Method, r.URL, resp.String())
		return nil
	})
}
