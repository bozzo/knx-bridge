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
	"context"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var done chan os.Signal

func init() {
	format := os.Getenv("LOG_FORMAT")
	// LOG_LEVEL not set, let's default to debug
	switch format {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	// LOG_LEVEL not set, let's default to debug
	if !ok {
		lvl = "debug"
	}
	// parse string, this is built-in feature of logrus
	ll, err := logrus.ParseLevel(lvl)
	if err != nil {
		ll = logrus.DebugLevel
	}
	// set global log level
	logrus.SetLevel(ll)
}

func main() {
	gwAddress := os.Getenv("GATEWAY_ADDRESS") + ":" + os.Getenv("GATEWAY_PORT")
	multicastAddress := os.Getenv("MULTICAST_ADDRESS") + ":" + os.Getenv("MULTICAST_PORT")

	done = make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	metricsServer := runMetricsServer()
	knxBridge, err := runKnxBridge(gwAddress, multicastAddress)

	if err != nil {
		logrus.Error(err)
	} else {
		// Waiting for SIGTERM
		<-done
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		if knxBridge != nil {
			knxBridge.closeKnxBridge()
		}
		cancel()
	}()

	if err := metricsServer.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server Shutdown Failed:%+v", err)
	}
	logrus.Info("Server Exited Properly")
}
