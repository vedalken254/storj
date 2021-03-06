// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storagenode

import (
	"context"

	"github.com/zeebo/errs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"storj.io/storj/internal/errs2"
	"storj.io/storj/internal/version"
	"storj.io/storj/pkg/auth/signing"
	"storj.io/storj/pkg/identity"
	"storj.io/storj/pkg/kademlia"
	"storj.io/storj/pkg/overlay"
	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/peertls/tlsopts"
	"storj.io/storj/pkg/server"
	"storj.io/storj/pkg/storj"
	"storj.io/storj/pkg/transport"
	"storj.io/storj/storage"
	"storj.io/storj/storagenode/bandwidth"
	"storj.io/storj/storagenode/collector"
	"storj.io/storj/storagenode/inspector"
	"storj.io/storj/storagenode/monitor"
	"storj.io/storj/storagenode/orders"
	"storj.io/storj/storagenode/pieces"
	"storj.io/storj/storagenode/piecestore"
	"storj.io/storj/storagenode/trust"
)

// DB is the master database for Storage Node
type DB interface {
	// CreateTables initializes the database
	CreateTables() error
	// Close closes the database
	Close() error

	Pieces() storage.Blobs

	Orders() orders.DB
	PieceInfo() pieces.DB
	CertDB() trust.CertDB
	Bandwidth() bandwidth.DB
	UsedSerials() piecestore.UsedSerials

	// TODO: use better interfaces
	RoutingTable() (kdb, ndb storage.KeyValueStore)
}

// Config is all the configuration parameters for a Storage Node
type Config struct {
	Identity identity.Config

	Server   server.Config
	Kademlia kademlia.Config

	// TODO: flatten storage config and only keep the new one
	Storage   piecestore.OldConfig
	Storage2  piecestore.Config
	Collector collector.Config

	Version version.Config
}

// Verify verifies whether configuration is consistent and acceptable.
func (config *Config) Verify(log *zap.Logger) error {
	return config.Kademlia.Verify(log)
}

// Peer is the representation of a Storage Node.
type Peer struct {
	// core dependencies
	Log      *zap.Logger
	Identity *identity.FullIdentity
	DB       DB

	Transport transport.Client

	Server *server.Server

	Version *version.Service

	// services and endpoints
	// TODO: similar grouping to satellite.Peer
	Kademlia struct {
		RoutingTable *kademlia.RoutingTable
		Service      *kademlia.Kademlia
		Endpoint     *kademlia.Endpoint
		Inspector    *kademlia.Inspector
	}

	Storage2 struct {
		// TODO: lift things outside of it to organize better
		Trust     *trust.Pool
		Store     *pieces.Store
		Endpoint  *piecestore.Endpoint
		Inspector *inspector.Endpoint
		Monitor   *monitor.Service
		Sender    *orders.Sender
	}

	Collector *collector.Service
}

// New creates a new Storage Node.
func New(log *zap.Logger, full *identity.FullIdentity, db DB, config Config, versionInfo version.Info) (*Peer, error) {
	peer := &Peer{
		Log:      log,
		Identity: full,
		DB:       db,
	}

	var err error

	{
		test := version.Info{}
		if test != versionInfo {
			peer.Log.Sugar().Debugf("Binary Version: %s with CommitHash %s, built at %s as Release %v",
				versionInfo.Version.String(), versionInfo.CommitHash, versionInfo.Timestamp.String(), versionInfo.Release)
		}
		peer.Version = version.NewService(config.Version, versionInfo, "Storagenode")
	}

	{ // setup listener and server
		sc := config.Server
		options, err := tlsopts.NewOptions(peer.Identity, sc.Config)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}

		peer.Transport = transport.NewClient(options)

		peer.Server, err = server.New(options, sc.Address, sc.PrivateAddress, nil)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}
	}

	{ // setup kademlia
		config := config.Kademlia
		// TODO: move this setup logic into kademlia package
		if config.ExternalAddress == "" {
			config.ExternalAddress = peer.Addr()
		}

		pbVersion, err := versionInfo.Proto()
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}

		self := &overlay.NodeDossier{
			Node: pb.Node{
				Id: peer.ID(),
				Address: &pb.NodeAddress{
					Transport: pb.NodeTransport_TCP_TLS_GRPC,
					Address:   config.ExternalAddress,
				},
			},
			Type: pb.NodeType_STORAGE,
			Operator: pb.NodeOperator{
				Wallet: config.Operator.Wallet,
			},
			Version: *pbVersion,
		}

		kdb, ndb := peer.DB.RoutingTable()
		peer.Kademlia.RoutingTable, err = kademlia.NewRoutingTable(peer.Log.Named("routing"), self, kdb, ndb, &config.RoutingTableConfig)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}

		peer.Transport = peer.Transport.WithObservers(peer.Kademlia.RoutingTable)

		peer.Kademlia.Service, err = kademlia.NewService(peer.Log.Named("kademlia"), peer.Transport, peer.Kademlia.RoutingTable, config)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}

		peer.Kademlia.Endpoint = kademlia.NewEndpoint(peer.Log.Named("kademlia:endpoint"), peer.Kademlia.Service, peer.Kademlia.RoutingTable)
		pb.RegisterNodesServer(peer.Server.GRPC(), peer.Kademlia.Endpoint)

		peer.Kademlia.Inspector = kademlia.NewInspector(peer.Kademlia.Service, peer.Identity)
		pb.RegisterKadInspectorServer(peer.Server.PrivateGRPC(), peer.Kademlia.Inspector)
	}

	{ // setup storage
		trustAllSatellites := !config.Storage.SatelliteIDRestriction
		peer.Storage2.Trust, err = trust.NewPool(peer.Kademlia.Service, trustAllSatellites, config.Storage.WhitelistedSatelliteIDs)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}

		peer.Storage2.Store = pieces.NewStore(peer.Log.Named("pieces"), peer.DB.Pieces())

		peer.Storage2.Monitor = monitor.NewService(
			log.Named("piecestore:monitor"),
			peer.Kademlia.RoutingTable,
			peer.Storage2.Store,
			peer.DB.PieceInfo(),
			peer.DB.Bandwidth(),
			config.Storage.AllocatedDiskSpace.Int64(),
			config.Storage.AllocatedBandwidth.Int64(),
			//TODO use config.Storage.Monitor.Interval, but for some reason is not set
			config.Storage.KBucketRefreshInterval,
		)

		peer.Storage2.Endpoint, err = piecestore.NewEndpoint(
			peer.Log.Named("piecestore"),
			signing.SignerFromFullIdentity(peer.Identity),
			peer.Storage2.Trust,
			peer.Storage2.Monitor,
			peer.Storage2.Store,
			peer.DB.PieceInfo(),
			peer.DB.Orders(),
			peer.DB.Bandwidth(),
			peer.DB.UsedSerials(),
			config.Storage2,
		)
		if err != nil {
			return nil, errs.Combine(err, peer.Close())
		}
		pb.RegisterPiecestoreServer(peer.Server.GRPC(), peer.Storage2.Endpoint)

		peer.Storage2.Inspector = inspector.NewEndpoint(
			peer.Log.Named("pieces:inspector"),
			peer.DB.PieceInfo(),
			peer.Kademlia.Service,
			peer.DB.Bandwidth(),
			config.Storage,
		)
		pb.RegisterPieceStoreInspectorServer(peer.Server.PrivateGRPC(), peer.Storage2.Inspector)

		peer.Storage2.Sender = orders.NewSender(
			log.Named("piecestore:orderssender"),
			peer.Transport,
			peer.Kademlia.Service,
			peer.DB.Orders(),
			config.Storage2.Sender,
		)
	}

	peer.Collector = collector.NewService(peer.Log.Named("collector"), peer.Storage2.Store, peer.DB.PieceInfo(), config.Collector)

	return peer, nil
}

// Run runs storage node until it's either closed or it errors.
func (peer *Peer) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return errs2.IgnoreCanceled(peer.Version.Run(ctx))
	})

	group.Go(func() error {
		return errs2.IgnoreCanceled(peer.Kademlia.Service.Bootstrap(ctx))
	})
	group.Go(func() error {
		return errs2.IgnoreCanceled(peer.Kademlia.Service.Run(ctx))
	})

	group.Go(func() error {
		return errs2.IgnoreCanceled(peer.Collector.Run(ctx))
	})
	group.Go(func() error {
		return errs2.IgnoreCanceled(peer.Storage2.Sender.Run(ctx))
	})
	group.Go(func() error {
		return errs2.IgnoreCanceled(peer.Storage2.Monitor.Run(ctx))
	})

	group.Go(func() error {
		// TODO: move the message into Server instead
		// Don't change the format of this comment, it is used to figure out the node id.
		peer.Log.Sugar().Infof("Node %s started", peer.Identity.ID)
		peer.Log.Sugar().Infof("Public server started on %s", peer.Addr())
		peer.Log.Sugar().Infof("Private server started on %s", peer.PrivateAddr())
		return errs2.IgnoreCanceled(peer.Server.Run(ctx))
	})

	return group.Wait()
}

// Close closes all the resources.
func (peer *Peer) Close() error {
	var errlist errs.Group

	// TODO: ensure that Close can be called on nil-s that way this code won't need the checks.

	// close servers, to avoid new connections to closing subsystems
	if peer.Server != nil {
		errlist.Add(peer.Server.Close())
	}

	// close services in reverse initialization order

	if peer.Storage2.Monitor != nil {
		errlist.Add(peer.Storage2.Monitor.Close())
	}
	if peer.Storage2.Sender != nil {
		errlist.Add(peer.Storage2.Sender.Close())
	}
	if peer.Collector != nil {
		errlist.Add(peer.Collector.Close())
	}

	if peer.Kademlia.Service != nil {
		errlist.Add(peer.Kademlia.Service.Close())
	}
	if peer.Kademlia.RoutingTable != nil {
		errlist.Add(peer.Kademlia.RoutingTable.Close())
	}

	return errlist.Err()
}

// ID returns the peer ID.
func (peer *Peer) ID() storj.NodeID { return peer.Identity.ID }

// Local returns the peer local node info.
func (peer *Peer) Local() overlay.NodeDossier { return peer.Kademlia.RoutingTable.Local() }

// Addr returns the public address.
func (peer *Peer) Addr() string { return peer.Server.Addr().String() }

// PrivateAddr returns the private address.
func (peer *Peer) PrivateAddr() string { return peer.Server.PrivateAddr().String() }
