package dtmcli

import (
	"fmt"

	"github.com/yedf/dtm/dtmcli/dtmimp"
)

// MustGenGid generate a new gid
func MustGenGid(server string) string {
	res := map[string]string{}
	resp, err := dtmimp.RestyClient.R().SetResult(&res).Get(server + "/newGid")
	if err != nil || res["gid"] == "" {
		panic(fmt.Errorf("newGid error: %v, resp: %s", err, resp))
	}
	return res["gid"]
}

// DB interface
type DB = dtmimp.DB

// Tx interface
type Tx = dtmimp.Tx

// SetCurrentDBType set currentDBType
func SetCurrentDBType(dbType string) {
	dtmimp.SetCurrentDBType(dbType)
}

// GetCurrentDBType get currentDBType
func GetCurrentDBType() string {
	return dtmimp.GetCurrentDBType()
}
