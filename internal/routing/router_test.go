package routing
import (
 "context"; "io"; "testing"
 "github.com/ananm2958/llm-gateway/internal/providers"
)
type testProvider struct{name string; models []string}
func(p testProvider)Name()string{return p.name};func(p testProvider)SupportedModels()[]string{return p.models};func(p testProvider)Chat(context.Context,*providers.ChatRequest)(*providers.ChatResponse,error){return &providers.ChatResponse{},nil};func(p testProvider)ChatStream(context.Context,*providers.ChatRequest)(io.ReadCloser,error){return nil,nil};func(p testProvider)Healthy(context.Context)bool{return true}
func TestBuildChainRoundRobinAndOverride(t *testing.T) { r:=NewRouter(providers.NewRegistry([]providers.Provider{testProvider{name:"a"},testProvider{name:"b"}}),&RoutingPolicy{ProviderChain:[]string{"a","b"},Strategy:StrategyRoundRobin,ModelOverrides:map[string]ModelOverride{"alias":{ProviderName:"b",ModelName:"real"}}}); first,err:=r.buildChain(&providers.ChatRequest{Model:"x"});if err!=nil||first[0]!="a"{t.Fatalf("first chain: %v, %v",first,err)};second,_:=r.buildChain(&providers.ChatRequest{Model:"x"});if second[0]!="b"{t.Fatalf("second chain: %v",second)};req:=&providers.ChatRequest{Model:"alias"};chain,_:=r.buildChain(req);if req.Model!="real"||chain[0]!="b"{t.Fatalf("override failed: %#v %#v",req,chain)} }
