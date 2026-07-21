package routing
import (
 "context"; "errors"; "io"; "testing"
 "github.com/ananm2958/llm-gateway/internal/providers"
)
type fallbackProvider struct{name string; fail bool; calls int}
func(p *fallbackProvider)Name()string{return p.name};func(p *fallbackProvider)SupportedModels()[]string{return []string{"m"}};func(p *fallbackProvider)Chat(context.Context,*providers.ChatRequest)(*providers.ChatResponse,error){p.calls++;if p.fail{return nil,errors.New("failed")};return &providers.ChatResponse{Model:"m"},nil};func(p *fallbackProvider)ChatStream(context.Context,*providers.ChatRequest)(io.ReadCloser,error){return nil,nil};func(p *fallbackProvider)Healthy(context.Context)bool{return true}
func TestFallbackUsesNextProviderAndSkipsOpenBreaker(t *testing.T){ bad:=&fallbackProvider{name:"bad",fail:true};good:=&fallbackProvider{name:"good"};f:=NewFallbackExecutor(providers.NewRegistry([]providers.Provider{bad,good}));for i:=0;i<7;i++ {resp,name,err:=f.Execute(context.Background(),&providers.ChatRequest{Model:"m"},[]string{"bad","good"});if err!=nil||resp==nil||name!="good"{t.Fatalf("unexpected fallback result: %v %q %v",resp,name,err)}};if bad.calls!=6||good.calls!=7{t.Fatalf("breaker did not skip failed provider: bad=%d good=%d",bad.calls,good.calls)} }
