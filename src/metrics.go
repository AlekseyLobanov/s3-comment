package main

import "github.com/penglongli/gin-metrics/ginmetrics"

func GetPrometheusHandler() *ginmetrics.Monitor {
	m := ginmetrics.GetMonitor()

	m.SetMetricPath("/metrics")
	m.SetSlowTime(2)
	m.SetDuration([]float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10})
	return m
}
