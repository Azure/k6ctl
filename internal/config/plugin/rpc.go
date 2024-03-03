package plugin

import (
	"encoding/gob"
	"net/rpc"
)

func init() {
	gob.Register([]interface{}{})
}

type rpcServer struct {
	Impl Interface
}

func (g *rpcServer) GetNames(args interface{}, resp *[]string) error {
	r, err := g.Impl.GetNames()
	*resp = r
	return err
}

func (g *rpcServer) Resolve(req ResolveRequest, resp *string) error {
	r, err := g.Impl.Resolve(req)
	*resp = r
	return err
}

type rpcClient struct{ client *rpc.Client }

func (g *rpcClient) GetNames() ([]string, error) {
	var resp []string
	err := g.client.Call("Plugin.GetNames", new(interface{}), &resp)
	return resp, err
}

func (g *rpcClient) Resolve(req ResolveRequest) (string, error) {
	var resp string
	err := g.client.Call("Plugin.Resolve", req, &resp)
	return resp, err
}
