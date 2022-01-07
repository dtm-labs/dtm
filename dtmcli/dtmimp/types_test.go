/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmimp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	err := CatchP(func() {
		idGen := BranchIDGen{BranchID: "12345678901234567890123"}
		idGen.NewSubBranchID()
	})
	assert.Error(t, err)
	err = CatchP(func() {
		idGen := BranchIDGen{subBranchID: 99}
		idGen.NewSubBranchID()
	})
	assert.Error(t, err)
}
