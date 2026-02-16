package metrics

import (
	"strings"
	"unicode/utf8"

	"github.com/prometheus/client_golang/prometheus"
)

// MustNewCounterVec 新增带labels的计数器，如果注册失败会panic
func MustNewCounterVec(subsystem, name string, labels []string) *prometheus.CounterVec {
	constrainedLabels := prometheus.ConstrainedLabels{}
	for _, label := range labels {
		constrainedLabels = append(constrainedLabels, prometheus.ConstrainedLabel{Name: label,
			Constraint: constraintLabels})
	}

	counter := prometheus.V2.NewCounterVec(prometheus.CounterVecOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace:   "tccc",
			Subsystem:   subsystem,
			Name:        name,
			ConstLabels: getConstLabels(),
		}, VariableLabels: constrainedLabels})
	err := prometheus.DefaultRegisterer.Register(counter)
	if err != nil {
		// slog.Infof(context.Background(), "register counter failed, err: %v", err)
		panic(err)
	}
	return counter
}

func getConstLabels() prometheus.Labels {
	return prometheus.Labels{}
}

var maxLableLength = 256

// constraintLabels 清洗标签值
func constraintLabels(src string) string {
	// 限制长度,防止标签值过大
	if len(src) > maxLableLength {
		src = src[:maxLableLength] + "..."
	}

	// 清洗UTF-8
	if !utf8.ValidString(src) {
		src = strings.ToValidUTF8(src, "�")
	}

	return src
}
