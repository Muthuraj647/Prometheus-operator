package handler

import (
	prometheus "github.com/prometheus/client_golang/prometheus"
)

var (
	validate_requests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_validate_requests_total",
			Help: "Count of Validate Requests from API Server",
		},
		[]string{"verb", "code"},
	)
	validate_requests_failed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pod_identity_webhook_mutation_failed",
			Help: "Count of Pod Identity Webhook Mutation Failed",
		},
	)
	validate_requests_status = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pod_identity_mutation_validation_details",
			Help: "Details of Pod Identity Webhook Mutation",
		},
		[]string{"namespace", "pod", "status"},
	)
)

func init() {
	register()
}

func register() {
	prometheus.MustRegister(validate_requests)
	prometheus.MustRegister(validate_requests_status)
	prometheus.MustRegister(validate_requests_failed)
}

//record the samples

func RecordRequests(verb, code string) {
	validate_requests.WithLabelValues(verb, code).Inc()
}

func RecordValidationFailures() {
	validate_requests_failed.Inc()
}

func RecordValidationGauge(namespace, pod, status string) {
	validate_requests_status.WithLabelValues(namespace, pod, status).Set(1)
}

func Reset() {
	validate_requests_status.Reset()
}
