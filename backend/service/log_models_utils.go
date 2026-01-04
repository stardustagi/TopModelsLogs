package service

import "fmt"

// parseLatency 解析延迟字符串为float64
func (s *LogService) parseLatency(latency string) float64 {
	var result float64
	fmt.Sscanf(latency, "%f", &result)
	return result
}
