package rpc

import (
	"context"
	"encoding/json"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklai-feed/manager"
	"github.com/nuklai/nuklaivm/consts"
)

type JSONRPCServer struct {
	m *manager.Manager
}

func NewJSONRPCServer(m *manager.Manager) *JSONRPCServer {
	return &JSONRPCServer{m}
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *jsonrpcError `json:"error,omitempty"`
	ID      interface{}   `json:"id"`
}

type jsonrpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (j *JSONRPCServer) HandleRequest(req JSONRPCRequest) JSONRPCResponse {
	var result interface{}
	var jsonErr *jsonrpcError

	switch req.Method {
	case "feedInfo":
		var params struct{}
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			jsonErr = &jsonrpcError{Code: -32602, Message: "Invalid params"}
			break
		}
		var reply FeedInfoReply
		jsonErr = j.FeedInfo(params, &reply)
		result = reply

	case "feed":
		var params FeedArgs
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			jsonErr = &jsonrpcError{Code: -32602, Message: "Invalid params"}
			break
		}
		var reply FeedReply
		jsonErr = j.Feed(params, &reply)
		result = reply

	case "updateNuklaiRPC":
		var params UpdateNuklaiRPCArgs
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			jsonErr = &jsonrpcError{Code: -32602, Message: "Invalid params"}
			break
		}
		var reply UpdateNuklaiRPCReply
		jsonErr = j.UpdateNuklaiRPC(params, &reply)
		result = reply

	default:
		jsonErr = &jsonrpcError{Code: -32601, Message: "Method not found"}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		Error:   jsonErr,
		ID:      req.ID,
	}
}

type FeedInfoReply struct {
	Address string `json:"address"`
	Fee     uint64 `json:"fee"`
}

func (j *JSONRPCServer) FeedInfo(_ struct{}, reply *FeedInfoReply) *jsonrpcError {
	addr, fee, err := j.m.GetFeedInfo(context.Background())
	if err != nil {
		return &jsonrpcError{Code: -32000, Message: err.Error()}
	}
	reply.Address = codec.MustAddressBech32(consts.HRP, addr)
	reply.Fee = fee
	return nil
}

type FeedArgs struct {
	SubnetID string `json:"subnetID"`
	ChainID  string `json:"chainID"`
	Limit    int    `json:"limit"`
}

type FeedReply struct {
	Feed []*manager.FeedObject `json:"feed"`
}

func (j *JSONRPCServer) Feed(args FeedArgs, reply *FeedReply) *jsonrpcError {
	feed, err := j.m.GetFeed(context.Background(), args.SubnetID, args.ChainID, args.Limit)
	if err != nil {
		return &jsonrpcError{Code: -32000, Message: err.Error()}
	}
	reply.Feed = feed
	return nil
}

type UpdateNuklaiRPCArgs struct {
	NuklaiRPCUrl string `json:"nuklaiRPCUrl"`
	AdminToken   string `json:"adminToken"`
}

type UpdateNuklaiRPCReply struct {
	Success bool `json:"success"`
}

func (j *JSONRPCServer) UpdateNuklaiRPC(args UpdateNuklaiRPCArgs, reply *UpdateNuklaiRPCReply) *jsonrpcError {
	if args.AdminToken != j.m.Config().AdminToken {
		return &jsonrpcError{Code: -32000, Message: "unauthorized user"}
	}
	err := j.m.UpdateNuklaiRPC(context.Background(), args.NuklaiRPCUrl)
	if err != nil {
		return &jsonrpcError{Code: -32000, Message: err.Error()}
	}
	reply.Success = true
	return nil
}
