package test

import (
	"flag"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	api "github.com/prometheus/client_golang/api/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"golang.org/x/net/context"
)

const (
	subsystem = "test_exporter"
)

var (
	namespace = flag.String("namespace", "prometheus", "Namespace of metrics.")
)

type simpleTestCase struct {
	prometheus.GaugeFunc
	name            string
	expectedValueAt func(time.Time) float64
}

func NewSimpleTestCase(name string, f func(time.Time) float64) TestCase {
	return &simpleTestCase{
		GaugeFunc: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: *namespace,
				Subsystem: subsystem,
				Name:      name,
				Help:      name,
			},
			func() float64 {
				return f(time.Now())
			},
		),
		name:            name,
		expectedValueAt: f,
	}
}

func (tc *simpleTestCase) ExpectedValueAt(t time.Time) float64 {
	return tc.expectedValueAt(t)
}

func (tc *simpleTestCase) Query(ctx context.Context, client api.QueryAPI, start time.Time, duration time.Duration) ([]model.SamplePair, error) {
	metricName := prometheus.BuildFQName(*namespace, subsystem, tc.name)
	query := metricName + fmt.Sprintf("[%dm]", duration/time.Minute)
	log.Println(query, "@", start)

	value, err := client.Query(ctx, query, start)
	if err != nil {
		return nil, err
	}
	if value.Type() != model.ValMatrix {
		return nil, fmt.Errorf("Didn't get matrix from Prom!")
	}

	ms, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("Didn't get matrix from Prom!")
	}

	result := []model.SamplePair{}
	for _, stream := range ms {
		for _, pair := range stream.Values {
			result = append(result, pair)
		}
	}
	return result, nil
}