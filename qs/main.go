/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package main

import (
	"time"

	"github.com/dtm-labs/dtm/test/busi"
)

func main() {
	busi.QsStartSvr()
	busi.QsFireRequest()
	time.Sleep(1 * time.Second)
}
