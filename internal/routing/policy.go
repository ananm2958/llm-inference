package routing

type RoutingPolicy struct {
	TenantID string

	ProviderChain []string

	ModelOverrides map[string]ModelOverride

	Strategy RoutingStrategy
}

type ModelOverride struct {
	ProviderName string
	ModelName    string
}

type RoutingStrategy string

const (
	StrategyPriority   RoutingStrategy = "priority"
	StrategyRoundRobin RoutingStrategy = "round_robin"
)
