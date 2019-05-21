// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package overlay_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/require"

	"storj.io/storj/pkg/overlay"
	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/storj"
	"storj.io/storj/satellite"
	"storj.io/storj/satellite/satellitedb/satellitedbtest"
)

func BenchmarkOverlay(b *testing.B) {
	satellitedbtest.Bench(b, func(b *testing.B, db satellite.DB) {
		const (
			TotalNodeCount = 211
			OnlineCount    = 90
			OfflineCount   = 10
		)

		overlaydb := db.OverlayCache()
		ctx := context.Background()

		var all []storj.NodeID
		var check []storj.NodeID
		for i := 0; i < TotalNodeCount; i++ {
			var id storj.NodeID
			_, _ = rand.Read(id[:]) // math/rand never returns error
			all = append(all, id)
			if i < OnlineCount {
				check = append(check, id)
			}
		}

		for _, id := range all {
			err := overlaydb.UpdateAddress(ctx, &pb.Node{Id: id})
			require.NoError(b, err)
		}

		// create random offline node ids to check
		for i := 0; i < OfflineCount; i++ {
			var id storj.NodeID
			_, _ = rand.Read(id[:]) // math/rand never returns error
			check = append(check, id)
		}

		b.Run("UnreliableOrOffline", func(b *testing.B) {
			criteria := &overlay.NodeCriteria{
				AuditCount:         0,
				AuditSuccessRatio:  0.5,
				OnlineWindow:       1000 * time.Hour,
				UptimeCount:        0,
				UptimeSuccessRatio: 0.5,
			}
			for i := 0; i < b.N; i++ {
				badNodes, err := overlaydb.UnreliableOrOffline(ctx, criteria, check)
				require.NoError(b, err)
				require.Len(b, badNodes, OfflineCount)
			}
		})

		b.Run("UpdateAddress", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				id := all[i%len(all)]
				err := overlaydb.UpdateAddress(ctx, &pb.Node{Id: id})
				require.NoError(b, err)
			}
		})

		b.Run("UpdateStats", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				id := all[i%len(all)]
				_, err := overlaydb.UpdateStats(ctx, &overlay.UpdateRequest{
					NodeID:       id,
					AuditSuccess: i&1 == 0,
					IsUp:         i&2 == 0,
				})
				require.NoError(b, err)
			}
		})

		b.Run("UpdateNodeInfo", func(b *testing.B) {
			now := ptypes.TimestampNow()
			for i := 0; i < b.N; i++ {
				id := all[i%len(all)]
				_, err := overlaydb.UpdateNodeInfo(ctx, id, &pb.InfoResponse{
					Type: pb.NodeType_STORAGE,
					Operator: &pb.NodeOperator{
						Wallet: "0x0123456789012345678901234567890123456789",
						Email:  "a@example.com",
					},
					Capacity: &pb.NodeCapacity{
						FreeBandwidth: 1000,
						FreeDisk:      1000,
					},
					Version: &pb.NodeVersion{
						Version:    "1.0.0",
						CommitHash: "0",
						Timestamp:  now,
						Release:    false,
					},
				})
				require.NoError(b, err)
			}
		})

		b.Run("UpdateUptime", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				id := all[i%len(all)]
				_, err := overlaydb.UpdateUptime(ctx, id, i&1 == 0)
				require.NoError(b, err)
			}
		})
	})
}