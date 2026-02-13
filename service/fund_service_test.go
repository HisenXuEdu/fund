package service

import (
	"testing"
	"time"
)

// TestFetchBatchFundsForRealtime 测试批量获取基金实时数据
func TestFetchBatchFundsForRealtime(t *testing.T) {
	service := NewFundService()

	// 获取第一页数据
	data, err := service.FetchBatchFundsForRealtime(1, 200)

	// 基本验证
	if err != nil {
		t.Fatalf("❌ 获取失败: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("❌ 返回数据为空")
	}

	t.Logf("✅ 成功获取 %d 只基金数据", len(data))

	// 检查数据结构
	count := 0
	validCount := 0
	for code, fundData := range data {
		if count >= 5 { // 只检查前5只基金
			break
		}

		// 验证基金代码
		if code == "" {
			t.Errorf("❌ 基金代码为空")
			continue
		}

		// 验证必要字段
		name, hasName := fundData["name"].(string)
		if !hasName || name == "" {
			t.Errorf("❌ 基金 %s 名称缺失", code)
			continue
		}

		netValue, hasNet := fundData["netValue"].(string)
		dayGrowth := fundData["dayGrowth"]

		if hasNet && netValue != "" && netValue != "---" {
			t.Logf("  ✓ [%s] %s | 净值: %s | 涨跌: %s",
				code, name, netValue, dayGrowth)
			validCount++
		}

		count++
	}

	if validCount == 0 {
		t.Error("❌ 没有找到有效数据")
	} else {
		t.Logf("✅ 数据结构验证通过，有效数据: %d/%d", validCount, count)
	}
}

// TestBatchFetchAndMemoryWrite 测试批量获取和内存写入性能
func TestBatchFetchAndMemoryWrite(t *testing.T) {
	t.Log("🧪 开始压力测试：每500ms获取一次，共10次")
	t.Log("=" + "==============================================")

	intradayService := NewIntradayService()

	// 获取次数
	rounds := 10
	interval := 500 * time.Millisecond

	var totalFunds int
	var totalTime time.Duration

	for i := 1; i <= rounds; i++ {
		t.Logf("\n📊 第 %d/%d 轮采集", i, rounds)

		startTime := time.Now()

		// 模拟批量获取并写入内存
		intradayService.fetchAllFundsRealtimeBatch()

		elapsed := time.Since(startTime)
		totalTime += elapsed

		// 检查内存中的数据数量
		dataCount := intradayService.GetDataCount()
		totalFunds = dataCount

		t.Logf("  ⏱️  耗时: %v", elapsed)
		t.Logf("  💾 内存数据量: %d 只基金", dataCount)

		// 最后一次不等待
		if i < rounds {
			time.Sleep(interval)
		}
	}

	t.Log("\n" + "=" + "==============================================")
	t.Logf("✅ 压力测试完成")
	t.Logf("📈 总轮数: %d 次", rounds)
	t.Logf("⏱️  总耗时: %v", totalTime)
	t.Logf("📊 平均耗时: %v/次", totalTime/time.Duration(rounds))
	t.Logf("💾 最终数据: %d 只基金", totalFunds)

	// 验证数据
	if totalFunds == 0 {
		t.Error("❌ 内存中没有数据")
	}
}
