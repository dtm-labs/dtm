/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package dtmsvr

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/dtm-labs/dtm/client/dtmcli/dtmimp"
	"github.com/dtm-labs/logger"
)

type transMsgProcessor struct {
	*TransGlobal
}

func init() {
	registorProcessorCreator("msg", func(trans *TransGlobal) transProcessor { return &transMsgProcessor{TransGlobal: trans} })
}

func (t *transMsgProcessor) GenBranches() []TransBranch {
	branches := []TransBranch{}
	for i, step := range t.Steps {
		mayTopic := strings.TrimPrefix(step[dtmimp.OpAction], dtmimp.MsgTopicPrefix)
		urls := dtmimp.If(mayTopic == step[dtmimp.OpAction], []string{mayTopic}, topic2urls(mayTopic)).([]string)
		if len(urls) == 0 {
			e2p(errors.New("topic not found"))
		}
		for j, url := range urls {
			b := TransBranch{
				Gid:      t.Gid,
				BranchID: fmt.Sprintf("%02d%s", i+1, dtmimp.If(len(urls) == 1, "", fmt.Sprintf("-%02d", j+1)).(string)),
				BinData:  t.BinPayloads[i],
				URL:      url,
				Op:       dtmimp.OpAction,
				Status:   dtmcli.StatusPrepared,
			}
			branches = append(branches, b)
		}
	}
	return branches
}

type cMsgCustom struct {
	Delay uint64 //delay call branch, unit second
}

func (t *TransGlobal) mayQueryPrepared(ctx context.Context) {
	if !t.needProcess() || t.Status == dtmcli.StatusSubmitted {
		return
	}
	err := t.getURLResult(ctx, t.QueryPrepared, "00", "msg", nil)
	if err == nil {
		t.changeStatus(dtmcli.StatusSubmitted)
	} else if errors.Is(err, dtmcli.ErrFailure) {
		t.changeStatus(dtmcli.StatusFailed)
	} else if errors.Is(err, dtmcli.ErrOngoing) {
		t.touchCronTime(cronReset, 0)
	} else {
		logger.Errorf("getting result failed for %s. error: %v", t.QueryPrepared, err)
		t.touchCronTime(cronBackoff, 0)
	}
}

func (t *transMsgProcessor) ProcessOnce(ctx context.Context, branches []TransBranch) error {
	t.mayQueryPrepared(ctx)
	if !t.needProcess() || t.Status == dtmcli.StatusPrepared {
		return nil
	}
	cmc := cMsgCustom{Delay: 0}
	if t.CustomData != "" {
		dtmimp.MustUnmarshalString(t.CustomData, &cmc)
	}

	if cmc.Delay > 0 && t.needDelay(cmc.Delay) {
		t.touchCronTime(cronKeep, cmc.Delay)
		return nil
	}
	var started int
	resultsChan := make(chan error, len(branches))
	var err error
	for i := range branches {
		b := &branches[i]
		if b.Op != dtmimp.OpAction || b.Status != dtmcli.StatusPrepared {
			continue
		}
		if t.Concurrent {
			copyCtx := NewAsyncContext(ctx)
			started++
			go func(ctx context.Context, pos int) {
				resultsChan <- t.execBranch(ctx, b, pos)
			}(copyCtx, i)
		} else {
			err = t.execBranch(ctx, b, i)
			if err != nil {
				break
			}
		}
	}
	for i := 0; i < started && err == nil; i++ {
		err = <-resultsChan
	}
	if err == dtmcli.ErrOngoing {
		return nil
	} else if err != nil {
		return err
	}
	t.changeStatus(dtmcli.StatusSucceed)
	return nil
}
