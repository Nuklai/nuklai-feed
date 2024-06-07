package rpc

import (
	"context"

	"github.com/ava-labs/hypersdk/requester"
	"github.com/nuklai/nuklai-feed/manager"
)

type JSONRPCClient struct {
	requester *requester.EndpointRequester
}

func NewJSONRPCClient(uri string) *JSONRPCClient {
	req := requester.New(uri, "feed")
	return &JSONRPCClient{
		requester: req,
	}
}

func (cli *JSONRPCClient) FeedInfo(ctx context.Context) (string, uint64, error) {
	resp := new(FeedInfoReply)
	err := cli.requester.SendRequest(
		ctx,
		"feedInfo",
		nil,
		resp,
	)
	return resp.Address, resp.Fee, err
}

func (cli *JSONRPCClient) Feed(ctx context.Context, subnetID, chainID string, limit int) ([]*manager.FeedObject, error) {
	resp := new(FeedReply)
	err := cli.requester.SendRequest(
		ctx,
		"feed",
		&FeedArgs{
			SubnetID: subnetID,
			ChainID:  chainID,
			Limit:    limit,
		},
		resp,
	)
	return resp.Feed, err
}

func (cli *JSONRPCClient) UpdateNuklaiRPC(ctx context.Context, newNuklaiRPCUrl, adminToken string) (bool, error) {
	resp := new(UpdateNuklaiRPCReply)
	err := cli.requester.SendRequest(
		ctx,
		"updateNuklaiRPC",
		&UpdateNuklaiRPCArgs{
			NuklaiRPCUrl: newNuklaiRPCUrl,
			AdminToken:   adminToken,
		},
		resp,
	)
	return resp.Success, err
}
