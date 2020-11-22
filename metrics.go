/*
 *
 *    Copyright 2020 Boris Barnier <bozzo@users.noreply.github.com>
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"syscall"
)

var (
	// Metric gauge
	tunnelEventCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "knx_bridge_tunnel_event",
		Help: "The event number received on the Tunnel interface",
	})
	// Metric gauge
	routerEventCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "knx_bridge_router_event",
		Help: "The event number received on the Router interface",
	})
)

func runMetricsServer() *http.Server {
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PROMETHEUS_PORT")
	if port == "" {
		port = "13671"
	}

	server := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: nil}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logrus.Error(err)
			done <- syscall.SIGTERM
		}
	}()

	logrus.Infof("Metrics server listening on port %s", port)
	return server
}
