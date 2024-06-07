package manager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/timer"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"
	fconfig "github.com/nuklai/nuklai-feed/config"
	"github.com/nuklai/nuklai-feed/database"
	"github.com/nuklai/nuklaivm/actions"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nrpc "github.com/nuklai/nuklaivm/rpc"
	"go.uber.org/zap"
)

type FeedContent struct {
	Message string `json:"message"`
	URL     string `json:"url"`
}

type FeedObject struct {
	SubnetID  string `json:"subnetID"`
	ChainID   string `json:"chainID"`
	Address   string `json:"address"`
	TxID      ids.ID `json:"txID"`
	Timestamp int64  `json:"timestamp"`
	Fee       uint64 `json:"fee"`

	Content *FeedContent `json:"content"`
}

type Manager struct {
	log    logging.Logger
	config *fconfig.Config

	ncli     *nrpc.JSONRPCClient
	subnetID ids.ID
	chainID  ids.ID

	l             sync.RWMutex
	t             *timer.Timer
	epochStart    int64
	epochMessages int
	feeAmount     uint64

	feed       []*FeedObject
	cancelFunc context.CancelFunc

	db *database.DB
}

func New(logger logging.Logger, config *fconfig.Config, db *sql.DB) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cli := rpc.NewJSONRPCClient(config.NuklaiRPC)
	networkID, subnetID, chainID, err := cli.Network(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	ncli := nrpc.NewJSONRPCClient(config.NuklaiRPC, networkID, chainID)

	dbInstance, err := database.NewDB(db)
	if err != nil {
		cancel()
		return nil, err
	}
	m := &Manager{log: logger, config: config, ncli: ncli, subnetID: subnetID, chainID: chainID, feed: []*FeedObject{}, cancelFunc: cancel, db: dbInstance}
	m.epochStart = time.Now().Unix()
	m.feeAmount = m.config.MinFee
	m.t = timer.NewTimer(m.updateFee)
	m.log.Info("feed initialized",
		zap.Uint32("network ID", networkID),
		zap.String("subnet ID", subnetID.String()),
		zap.String("chain ID", chainID.String()),
		zap.String("address", m.config.Recipient),
		zap.String("fee", utils.FormatBalance(m.feeAmount, nconsts.Decimals)),
	)

	return m, nil
}

func (m *Manager) saveFeed(feed *FeedObject) error {
	content, err := json.Marshal(feed.Content)
	if err != nil {
		m.log.Error("Failed to marshal feed content", zap.Error(err))
		return fmt.Errorf("failed to marshal feed content: %w", err)
	}
	err = m.db.SaveFeed(&database.FeedObject{
		TxID:      feed.TxID.String(),
		SubnetID:  feed.SubnetID,
		ChainID:   feed.ChainID,
		Address:   feed.Address,
		Timestamp: feed.Timestamp,
		Fee:       feed.Fee,
		Content:   string(content),
	})
	if err != nil {
		m.log.Error("Failed to save feed to database", zap.Error(err))
	}
	return err
}

func (m *Manager) getLastFeeds(n int) ([]*FeedObject, error) {
	feeds, err := m.db.GetLastFeeds(n)
	if err != nil {
		m.log.Error("Failed to get last feeds from database", zap.Error(err))
		return nil, err
	}
	var feedObjects []*FeedObject
	for _, feed := range feeds {
		var content FeedContent
		if err := json.Unmarshal([]byte(feed.Content), &content); err != nil {
			m.log.Error("Failed to unmarshal feed content", zap.Error(err))
			return nil, err
		}
		txID, err := ids.FromString(feed.TxID)
		if err != nil {
			m.log.Error("Failed to parse TxID from string", zap.Error(err))
			return nil, err
		}
		feedObjects = append(feedObjects, &FeedObject{
			SubnetID:  feed.SubnetID,
			ChainID:   feed.ChainID,
			Address:   feed.Address,
			TxID:      txID,
			Timestamp: feed.Timestamp,
			Fee:       feed.Fee,
			Content:   &content,
		})
	}
	return feedObjects, nil
}

func (m *Manager) appendFeed(feed *FeedObject) {
	m.log.Info("Appending new feed", zap.String("TxID", feed.TxID.String()))
	if err := m.saveFeed(feed); err != nil {
		m.log.Error("Failed to save feed", zap.Error(err))
	}
}

func (m *Manager) updateFee() {
	m.l.Lock()
	defer m.l.Unlock()

	now := time.Now().Unix()
	if now-m.epochStart < m.config.TargetDurationPerEpoch/2 {
		return
	}

	if m.feeAmount > m.config.MinFee && m.epochMessages == 0 {
		m.feeAmount -= m.config.FeeDelta
		m.log.Info("Decreasing message fee", zap.Uint64("fee", m.feeAmount))
	}
	m.epochMessages = 0
	m.epochStart = time.Now().Unix()
	m.t.SetTimeoutIn(time.Duration(m.config.TargetDurationPerEpoch) * time.Second)
	m.log.Info("Fee updated", zap.Int64("epochStart", m.epochStart), zap.Uint64("feeAmount", m.feeAmount))
}

func (m *Manager) Run(ctx context.Context) error {
	m.log.Info("Manager run started")
	m.t.SetTimeoutIn(time.Duration(m.config.TargetDurationPerEpoch) * time.Second)
	go m.t.Dispatch()
	defer m.t.Stop()

	var scli *rpc.WebSocketClient
	currentRPCURL := m.config.NuklaiRPC

	reconnect := func() error {
		var err error
		if scli != nil {
			scli.Close()
		}
		scli, err = rpc.NewWebSocketClient(m.config.NuklaiRPC, rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
		if err != nil {
			m.log.Warn("Failed to connect to RPC", zap.String("uri", m.config.NuklaiRPC), zap.Error(err))
			return fmt.Errorf("failed to connect to RPC: %w", err)
		}
		if err = scli.RegisterBlocks(); err != nil {
			m.log.Warn("Failed to register for blocks", zap.String("uri", m.config.NuklaiRPC), zap.Error(err))
			return fmt.Errorf("failed to register for blocks: %w", err)
		}
		m.log.Info("Connected to RPC and registered for blocks", zap.String("uri", m.config.NuklaiRPC))
		return nil
	}

	if err := reconnect(); err != nil {
		m.log.Error("Initial RPC connection failed", zap.Error(err))
		return err
	}

	for ctx.Err() == nil {
		if m.config.NuklaiRPC != currentRPCURL {
			m.log.Info("Detected RPC URL change, reconnecting", zap.String("newURL", m.config.NuklaiRPC))
			if err := reconnect(); err != nil {
				m.log.Error("Reconnection failed", zap.Error(err))
				continue
			}
			currentRPCURL = m.config.NuklaiRPC
		}

		parser, err := m.ncli.Parser(ctx)
		if err != nil {
			m.log.Error("Failed to create parser", zap.Error(err))
			return err
		}

		blk, results, _, err := scli.ListenBlock(ctx, parser)
		if err != nil {
			m.log.Warn("Unable to listen for blocks", zap.Error(err))
			time.Sleep(10 * time.Second)
			continue
		}

		for i, tx := range blk.Txs {
			result := results[i]
			if result.Success {
				for _, act := range tx.Actions {
					action, ok := act.(*actions.Transfer)

					recipientAddr, err := m.config.RecipientAddress()
					if err != nil {
						m.log.Error("Failed to get recipient address", zap.Error(err))
						return err
					}
					if !ok || action.To != recipientAddr {
						continue
					}

					fromStr := codec.MustAddressBech32(nconsts.HRP, tx.Auth.Actor())
					if !result.Success || action.Value < m.feeAmount {
						m.log.Info("Incoming message failed or did not pay enough", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value), zap.Uint64("required", m.feeAmount))
						continue
					}

					var content FeedContent
					if err := json.Unmarshal(action.Memo, &content); err != nil || len(content.Message) == 0 {
						m.log.Info("Incoming message could not be parsed or was empty", zap.String("from", fromStr), zap.String("memo", string(action.Memo)), zap.Uint64("payment", action.Value), zap.Error(err))
						continue
					}

					m.appendFeed(&FeedObject{
						SubnetID:  m.subnetID.String(),
						ChainID:   m.chainID.String(),
						Address:   fromStr,
						TxID:      tx.ID(),
						Timestamp: blk.Tmstmp,
						Fee:       action.Value,
						Content:   &content,
					})
				}
			}
		}

		time.Sleep(1 * time.Second)
	}

	m.log.Info("Manager run completed", zap.Error(ctx.Err()))
	return ctx.Err()
}

func (m *Manager) GetFeedInfo(_ context.Context) (codec.Address, uint64, error) {
	m.l.RLock()
	defer m.l.RUnlock()

	addr, err := m.config.RecipientAddress()
	if err != nil {
		m.log.Error("Failed to get recipient address", zap.Error(err))
	}
	return addr, m.feeAmount, err
}

func (m *Manager) GetFeed(_ context.Context, subnetID, chainID string, limit int) ([]*FeedObject, error) {
	return m.getLastFeeds(limit)
}

func (m *Manager) UpdateNuklaiRPC(ctx context.Context, newNuklaiRPCUrl string) error {
	m.l.Lock()
	defer m.l.Unlock()

	m.log.Info("Updating Nuklai RPC URL", zap.String("oldURL", m.config.NuklaiRPC), zap.String("newURL", newNuklaiRPCUrl))

	m.config.NuklaiRPC = newNuklaiRPCUrl

	cli := rpc.NewJSONRPCClient(newNuklaiRPCUrl)
	networkID, subnetID, chainID, err := cli.Network(ctx)
	if err != nil {
		m.log.Error("Failed to fetch network details", zap.Error(err))
		return fmt.Errorf("failed to fetch network details: %w", err)
	}

	m.ncli = nrpc.NewJSONRPCClient(newNuklaiRPCUrl, networkID, chainID)

	m.subnetID = subnetID
	m.chainID = chainID
	m.epochStart = time.Now().Unix()
	m.feeAmount = m.config.MinFee
	m.t = timer.NewTimer(m.updateFee)

	m.log.Info("RPC client has been updated and manager reinitialized",
		zap.String("new RPC URL", newNuklaiRPCUrl),
		zap.Uint32("network ID", networkID),
		zap.String("subnet ID", subnetID.String()),
		zap.String("chain ID", chainID.String()),
		zap.String("address", m.config.Recipient),
		zap.String("fee", utils.FormatBalance(m.feeAmount, nconsts.Decimals)),
	)

	return nil
}

// Config returns the configuration of the manager
func (m *Manager) Config() *fconfig.Config {
	return m.config
}
