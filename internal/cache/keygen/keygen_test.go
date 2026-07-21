package keygen

import (
 "testing"
 "github.com/ananm2958/llm-gateway/internal/providers"
)
func TestExactKeyDeterministic(t *testing.T) { req:=&providers.ChatRequest{Model:"m",Messages:[]providers.Message{{Role:"user",Content:"hello"}},Temperature:.2,MaxTokens:10}; first,err:=ExactKey("tenant",req);if err!=nil{t.Fatal(err)}; second,err:=ExactKey("tenant",req);if err!=nil{t.Fatal(err)};if first!=second{t.Fatalf("keys differ: %s != %s",first,second)} }
