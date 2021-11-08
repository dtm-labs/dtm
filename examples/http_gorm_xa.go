package examples

import (
	"github.com/go-resty/resty/v2"
	"github.com/yedf/dtm/dtmcli"
	"github.com/yedf/dtm/dtmcli/dtmimp"
)

func init() {
	addSample("xa_gorm", func() string {
		gid := dtmcli.MustGenGid(DtmServer)
		err := XaClient.XaGlobalTransaction(gid, func(xa *dtmcli.Xa) (*resty.Response, error) {
			resp, err := xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransOutXaGorm")
			if err != nil {
				return resp, err
			}
			return xa.CallBranch(&TransReq{Amount: 30}, Busi+"/TransInXa")
		})
		dtmimp.FatalIfError(err)
		return gid
	})

}
