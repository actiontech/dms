package core

// Confidence 置信度 (例如: High, Medium, Low)
// swagger:enum Confidence
type Confidence string

const (
	// 高：高度确信为敏感数据
	ConfidenceHigh Confidence = "High"
	// 中：中等确信为敏感数据
	ConfidenceMedium Confidence = "Medium"
	// 低：低确信为敏感数据
	ConfidenceLow Confidence = "Low"
)
