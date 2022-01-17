/*
 * Copyright (c) 2021 yedf. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package boltdb

import (
	"path"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	bolt "go.etcd.io/bbolt"

	"github.com/dtm-labs/dtm/dtmcli/dtmimp"
	"github.com/dtm-labs/dtm/dtmsvr/storage"
)

func TestInitializeBuckets(t *testing.T) {
	t.Run("normal test", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		err = initializeBuckets(db)
		g.Expect(err).ToNot(HaveOccurred())

		actualBuckets := [][]byte{}
		err = db.View(func(t *bolt.Tx) error {
			return t.ForEach(func(name []byte, _ *bolt.Bucket) error {
				actualBuckets = append(actualBuckets, name)
				return nil
			})
		})
		g.Expect(err).ToNot(HaveOccurred())

		g.Expect(actualBuckets).To(Equal(allBuckets))
	})
}

func TestCleanupExpiredData(t *testing.T) {
	t.Run("negative expired seconds", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		err = cleanupExpiredData(-1*time.Second, db)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("nil global bucket", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		err = cleanupExpiredData(time.Second, db)
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("normal test", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		// Initialize data
		err = initializeBuckets(db)
		g.Expect(err).ToNot(HaveOccurred())

		err = db.Update(func(t *bolt.Tx) error {
			doneTime := time.Now().Add(-10 * time.Minute)

			gids := []string{"gid0", "gid1", "gid2"}
			gidDatas := []storage.TransGlobalStore{
				{}, // not finished
				{
					FinishTime: &doneTime,
				},
				{
					RollbackTime: &doneTime,
				},
			}
			bucket := t.Bucket(bucketGlobal)
			for i := 0; i < len(gids); i++ {
				err = bucket.Put([]byte(gids[i]), dtmimp.MustMarshal(gidDatas[i]))
				if err != nil {
					return err
				}
			}

			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = cleanupExpiredData(time.Minute, db)
		g.Expect(err).ToNot(HaveOccurred())

		actualGids := []string{}
		err = db.View(func(t *bolt.Tx) error {
			cursor := t.Bucket(bucketGlobal).Cursor()
			for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
				actualGids = append(actualGids, string(k))
			}
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(actualGids).To(Equal([]string{"gid0"}))
	})
}

func TestCleanupGlobalWithGids(t *testing.T) {
	t.Run("nil bucket", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		err = db.Update(func(t *bolt.Tx) error {
			cleanupGlobalWithGids(t, nil)
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("normal test", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		// Initialize data
		err = db.Update(func(t *bolt.Tx) error {
			bucket, err := t.CreateBucketIfNotExists(bucketGlobal)
			if err != nil {
				return err
			}

			keys := []string{"k1", "k2", "k3"}
			datas := []string{"data1", "data2", "data3"}
			for i := 0; i < len(keys); i++ {
				err = bucket.Put([]byte(keys[i]), []byte(datas[i]))
				if err != nil {
					return err
				}
			}

			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = db.Update(func(t *bolt.Tx) error {
			cleanupGlobalWithGids(t, map[string]struct{}{
				"k1": {},
				"k2": {},
			})
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		actualGids := []string{}
		err = db.View(func(t *bolt.Tx) error {
			cursor := t.Bucket(bucketGlobal).Cursor()
			for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
				actualGids = append(actualGids, string(k))
			}
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(actualGids).To(Equal([]string{"k3"}))
	})
}

func TestCleanupBranchWithGids(t *testing.T) {
	t.Run("nil bucket", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		err = db.Update(func(t *bolt.Tx) error {
			cleanupBranchWithGids(t, nil)
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("normal test", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		// Initialize data
		err = db.Update(func(t *bolt.Tx) error {
			bucket, err := t.CreateBucketIfNotExists(bucketBranches)
			if err != nil {
				return err
			}

			keys := []string{"a", "gid001", "gid002", "gid101", "gid201", "z"}
			datas := []storage.TransBranchStore{
				{
					Gid: "a",
				},
				{
					Gid: "gid0",
				},
				{
					Gid: "gid0",
				},
				{
					Gid: "gid1",
				},
				{
					Gid: "gid2",
				},
				{
					Gid: "z",
				},
			}
			for i := 0; i < len(keys); i++ {
				err = bucket.Put([]byte(keys[i]), dtmimp.MustMarshal(datas[i]))
				if err != nil {
					return err
				}
			}

			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = db.Update(func(t *bolt.Tx) error {
			cleanupBranchWithGids(t, map[string]struct{}{
				"gid0": {},
				"gid1": {},
			})
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		actualKeys := []string{}
		err = db.View(func(t *bolt.Tx) error {
			cursor := t.Bucket(bucketBranches).Cursor()
			for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
				actualKeys = append(actualKeys, string(k))
			}
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(actualKeys).To(Equal([]string{"a", "gid201", "z"}))
	})
}

func TestCleanupIndexWithGids(t *testing.T) {
	t.Run("nil bucket", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		err = db.Update(func(t *bolt.Tx) error {
			cleanupIndexWithGids(t, nil)
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("normal test", func(t *testing.T) {
		g := NewWithT(t)
		db, err := bolt.Open(path.Join(t.TempDir(), "./test.bolt"), 0666, &bolt.Options{Timeout: 1 * time.Second})
		g.Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		// Initialize data
		err = db.Update(func(t *bolt.Tx) error {
			bucket, err := t.CreateBucketIfNotExists(bucketIndex)
			if err != nil {
				return err
			}

			keys := []string{"a", "0-gid0", "1-gid0", "2-gid1", "3-gid2", "z"}
			datas := []storage.TransBranchStore{
				{
					Gid: "a",
				},
				{
					Gid: "gid0",
				},
				{
					Gid: "gid0",
				},
				{
					Gid: "gid1",
				},
				{
					Gid: "gid2",
				},
				{
					Gid: "z",
				},
			}
			for i := 0; i < len(keys); i++ {
				err = bucket.Put([]byte(keys[i]), dtmimp.MustMarshal(datas[i]))
				if err != nil {
					return err
				}
			}

			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = db.Update(func(t *bolt.Tx) error {
			cleanupIndexWithGids(t, map[string]struct{}{
				"gid0": {},
				"gid1": {},
			})
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		actualKeys := []string{}
		err = db.View(func(t *bolt.Tx) error {
			cursor := t.Bucket(bucketIndex).Cursor()
			for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
				actualKeys = append(actualKeys, string(k))
			}
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(actualKeys).To(Equal([]string{"3-gid2", "a", "z"}))
	})
}
