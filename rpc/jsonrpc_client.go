// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"
	"strings"

	"github.com/ava-labs/hypersdk/requester"
	"github.com/nuklai/nuklai-feed/manager"
)

const (
	JSONRPCEndpoint = "/feed"
)

type JSONRPCClient struct {
	requester *requester.EndpointRequester
}

// New creates a new client object.
func NewJSONRPCClient(uri string) *JSONRPCClient {
	uri = strings.TrimSuffix(uri, "/")
	uri += JSONRPCEndpoint
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

// UpdateNuklaiRPC updates the RPC url for Nuklai
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
