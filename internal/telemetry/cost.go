package telemetry

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CostCalculator struct {
	pool *pgxpool.Pool
}

func NewCostCalculator(pool *pgxpool.Pool) *CostCalculator {
	return &CostCalculator{pool: pool}
}

func (c *CostCalculator) Calculate(
	ctx context.Context,
	provider, model string,
	promptTokens, completionTokens int,
) (float64, error) {
	var promptCost, completionCost float64

	err := c.pool.QueryRow(ctx, `
        SELECT prompt_cost_per_1k, completion_cost_per_1k
        FROM   model_pricing
        WHERE  provider = $1 AND model = $2
        ORDER  BY effective_from DESC
        LIMIT  1
    `, provider, model).Scan(&promptCost, &completionCost)

	if err != nil {
		return 0, nil
	}

	cost := (float64(promptTokens)/1000)*promptCost +
		(float64(completionTokens)/1000)*completionCost

	return cost, nil
}
