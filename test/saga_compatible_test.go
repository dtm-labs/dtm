/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"fmt"
	"testing"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/examples"
	"github.com/stretchr/testify/assert"
)

func TestSagaCompatibleNormal(t *testing.T) { // compatible with old http, which put payload in steps.data
	gid := dtmimp.GetFuncName()
	body := fmt.Sprintf(`{"gid":"%s","trans_type":"saga","steps":[{"action":"%s/TransOut","compensate":"%s/TransOutRevert","data":"{\"amount\":30,\"transInResult\":\"SUCCESS\",\"transOutResult\":\"SUCCESS\"}"},{"action":"%s/TransIn","compensate":"%s/TransInRevert","data":"{\"amount\":30,\"transInResult\":\"SUCCESS\",\"transOutResult\":\"SUCCESS\"}"}]}`,
		gid, examples.Busi, examples.Busi, examples.Busi, examples.Busi)
	dtmimp.RestyClient.R().SetBody(body).Post(fmt.Sprintf("%s/submit", examples.DtmHttpServer))
	waitTransProcessed(gid)
	assert.Equal(t, []string{StatusPrepared, StatusSucceed, StatusPrepared, StatusSucceed}, getBranchesStatus(gid))
	assert.Equal(t, StatusSucceed, getTransStatus(gid))
}
