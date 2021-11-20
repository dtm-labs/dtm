/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package test

import (
	"testing"

	"github.com/yedf/dtm/examples"
)

func TestExamples(t *testing.T) {
	examples.QsStartSvr()
	for _, s := range examples.Samples {
		assertSucceed(t, s.Action())
	}
}
