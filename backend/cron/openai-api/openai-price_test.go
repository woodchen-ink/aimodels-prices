package openai_api

import (
	"fmt"
	"testing"
)

func TestFetchOpenAIPrices(t *testing.T) {
	prices, err := fetchOpenAIPrices()
	if err != nil {
		t.Fatalf("fetchOpenAIPrices 失败: %v", err)
	}

	if len(prices) == 0 {
		t.Fatal("未解析到任何价格数据")
	}

	fmt.Printf("共解析到 %d 个模型价格:\n\n", len(prices))
	fmt.Printf("%-35s %12s %12s %12s\n", "Model", "Input", "Cached", "Output")
	fmt.Printf("%-35s %12s %12s %12s\n", "-----", "-----", "------", "------")

	for _, p := range prices {
		cachedStr := "-"
		if p.CachedPrice > 0 {
			cachedStr = fmt.Sprintf("$%.3f", p.CachedPrice)
		}
		fmt.Printf("%-35s %12s %12s %12s\n",
			p.Model,
			fmt.Sprintf("$%.3f", p.InputPrice),
			cachedStr,
			fmt.Sprintf("$%.3f", p.OutputPrice),
		)
	}
}
