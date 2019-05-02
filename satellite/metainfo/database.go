// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package metainfo

import (
	"context"
	"time"

	"github.com/skyrings/skyring-common/tools/uuid"

	"storj.io/storj/pkg/storj"
)

// Bucket defines internal implementation of buckets
type Bucket struct {
	ID uuid.UUID

	ProjectID  uuid.UUID
	Name       string
	PathCipher storj.CipherSuite

	AttributionID uuid.UUID // []byte?
	CreatedAt     time.Time

	// do we need "Default" prefix here?
	DefaultSegmentSize int64
	DefaultRedundancy  storj.RedundancyScheme
	DefaultEncryption  storj.EncryptionParameters
}

// BucketListOptions lists objects
type BucketListOptions struct {
	Cursor    string
	Direction storj.ListDirection
	Limit     int
}

// BucketList is a list of buckets
type BucketList struct {
	More  bool
	Items []*Bucket // TODO: does this need to be a pointer?
}

// NextPage returns options for listing the next page
func (opts BucketListOptions) NextPage(list BucketList) BucketListOptions {
	if !list.More || len(list.Items) == 0 {
		return BucketListOptions{}
	}

	switch opts.Direction {
	case storj.Before, storj.Backward:
		return BucketListOptions{
			Cursor:    list.Items[0].Name,
			Direction: storj.Before,
			Limit:     opts.Limit,
		}
	case storj.After, storj.Forward:
		return BucketListOptions{
			Cursor:    list.Items[len(list.Items)-1].Name,
			Direction: storj.After,
			Limit:     opts.Limit,
		}
	}

	return BucketListOptions{}
}

// TODO create interface metainfo.DB.Buckets()

// Buckets interface for managing buckets
type Buckets interface {
	Create(ctx context.Context, bucket *Bucket) error
	Get(ctx context.Context, projectID uuid.UUID, name string) (*Bucket, error)
	Delete(ctx context.Context, projectID uuid.UUID, name string) error
	List(ctx context.Context, projectID uuid.UUID, opts BucketListOptions) (BucketList, error)
}

type ObjectStatus byte

const (
	Partial ObjectStatus = iota
	Committing
	Committed
	Deleting
)

type ObjectVersion int64

const LastObjectVersion = ObjectVersion(-1)

type Object struct {
	BucketID      uuid.UUID
	EncryptedPath storj.Path
	Version       ObjectVersion
	Status        ObjectStatus

	StreamID uuid.UUID

	EncryptedMetadataNonce []byte
	EncryptedMetadata      []byte

	TotalSize  int64
	InlineSize int64
	RemoteSize int64

	CreatedAt time.Time
	ExpiresAt time.Time

	FixedSegmentSize int64
	Redundancy       storj.RedundancyScheme
	Encryption       storj.EncryptionParameters
}

type Segment struct {
	StreamID     uuid.UUID
	SegmentIndex uint64

	RootPieceID storj.PieceID

	EncryptedKeyNonce storj.Nonce
	EncryptedKey      storj.EncryptedPrivateKey

	SegmentChecksum uint64
	SegmentSize     int64

	EncryptedInlineData []byte
	Nodes               []storj.NodeID
}

// ListOptions lists objects
type ListOptions struct {
	EncryptedPrefix storj.Path
	EncryptedCursor storj.Path // Cursor is relative to Prefix, full path is Prefix + Cursor
	Recursive       bool
	Direction       storj.ListDirection
	Limit           int

	AllVersions bool
}

// ObjectList is a list of objects
type ObjectList struct {
	EncryptedPrefix storj.Path
	More            bool

	// Items paths are relative to Prefix
	// To get the full path use list.Prefix + list.Items[0].Path
	Items []*Object
}

// NextPage returns options for listing the next page
func (opts ListOptions) NextPage(list ObjectList) ListOptions {
	if !list.More || len(list.Items) == 0 {
		return ListOptions{}
	}

	switch opts.Direction {
	case storj.Before, storj.Backward:
		return ListOptions{
			EncryptedPrefix: opts.EncryptedPrefix,
			EncryptedCursor: list.Items[0].EncryptedPath,
			Direction:       storj.Before,
			Limit:           opts.Limit,
		}
	case storj.After, storj.Forward:
		return ListOptions{
			EncryptedPrefix: opts.EncryptedPrefix,
			EncryptedCursor: list.Items[len(list.Items)-1].EncryptedPath,
			Direction:       storj.After,
			Limit:           opts.Limit,
		}
	}

	return ListOptions{}
}

// Objects interface for managing objects
type Objects interface {
	Create(ctx context.Context, object *Object) error
	Commit(ctx context.Context, object *Object) (*Object, error)

	Get(ctx context.Context, bucket uuid.UUID, encryptedPath storj.Path, version uint32) (*Object, error)
	List(ctx context.Context, bucket uuid.UUID, opts ListOptions) (ObjectList, error)
	Delete(ctx context.Context, bucket uuid.UUID, encryptedPath storj.Path, version uint32) error

	// Should these be "Partial" or "Uncommitted" or "Any"?
	GetPartial(ctx context.Context, bucket uuid.UUID, encryptedPath storj.Path, version uint32) (*Object, error)
	ListPartial(ctx context.Context, bucket uuid.UUID, opts ListOptions) (ObjectList, error)
	DeletePartial(ctx context.Context, bucket uuid.UUID, encryptedPath storj.Path, version uint32) error
}

// Segments interface for managing segments
type Segments interface {
	Create(ctx context.Context, segment *Segment) error
	Commit(ctx context.Context, segment *Segment) error
	Delete(ctx context.Context, streamID uuid.UUID, segmentIndex int64) error
	List(ctx context.Context, streamID uuid.UUID, segmentIndex int64, limit int64) ([]*Segment, error)
}
