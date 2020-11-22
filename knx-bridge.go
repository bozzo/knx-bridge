// Copyright 2017 Ole Krüger.
// Licensed under the MIT license which can be found in the LICENSE file.

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/knxnet"
	"github.com/vapourismo/knx-go/knx/util"
	"net"
	"syscall"
)

type relay interface {
	relay(data cemi.LData) error

	Inbound() <-chan cemi.Message
	Close()
}

type reqRelay struct {
	*knx.Tunnel
}

func (relay reqRelay) relay(data cemi.LData) error {
	return relay.Send(&cemi.LDataReq{LData: data})
}

type indRelay struct {
	*knx.Router
}

func (relay indRelay) relay(data cemi.LData) error {
	return relay.Send(&cemi.LDataInd{LData: data})
}

func runKnxBridge(gatewayAddr string, otherAddr string) (br *bridge, err error) {
	util.Logger = logrus.StandardLogger()
	br, err = newBridge(gatewayAddr, otherAddr)
	if err != nil {
		return nil, err
	}

	logrus.Debug("Bridge created")
	go br.serve()

	return br, nil
}

type bridge struct {
	tunnel *knx.Tunnel
	other  relay
}

func newBridge(gatewayAddr, otherAddr string) (*bridge, error) {
	// Instantiate tunnel connection.
	tunnel, err := knx.NewTunnel(gatewayAddr, knxnet.TunnelLayerData, knx.DefaultTunnelConfig)
	if err != nil {
		return nil, err
	}

	var other relay

	addr, err := net.ResolveUDPAddr("udp4", otherAddr)
	if err != nil {
		tunnel.Close()
		return nil, err
	}

	if addr.IP.IsMulticast() {
		// Instantiate routing facilities.
		router, err := knx.NewRouter(otherAddr, knx.DefaultRouterConfig)
		if err != nil {
			tunnel.Close()
			return nil, err
		}

		other = indRelay{router}
	} else {
		// Instantiate tunnel connection.
		otherTunnel, err := knx.NewTunnel(otherAddr, knxnet.TunnelLayerData, knx.DefaultTunnelConfig)
		if err != nil {
			tunnel.Close()
			return nil, err
		}

		other = reqRelay{otherTunnel}
	}

	return &bridge{tunnel, other}, nil
}

func (br *bridge) serve() {
	for {
		logrus.Debug("Waiting in loop ...")

		select {
		// Receive message from gateway.
		case msg, open := <-br.tunnel.Inbound():
			if !open {
				logrus.Error("tunnel channel closed")
				done <- syscall.SIGTERM
			}

			tunnelEventCounter.Inc()

			if ind, ok := msg.(*cemi.LDataInd); ok {
				logrus.Debugf("tunnel: %+v", ind.LData)
				if err := br.other.relay(ind.LData); err != nil {
					logrus.Error(err)
					done <- syscall.SIGTERM
				}
			} else {
				logrus.Debugf("tunnel (not parsed): %+v", msg)
			}

		// Receive message from router.
		case msg, open := <-br.other.Inbound():
			if !open {
				logrus.Error("Router channel closed")
				done <- syscall.SIGTERM
			}

			routerEventCounter.Inc()

			if ind, ok := msg.(*cemi.LDataInd); ok {
				logrus.Debugf("router: %+v", ind)
				if err := br.tunnel.Send(&cemi.LDataReq{LData: ind.LData}); err != nil {
					logrus.Error(err)
					done <- syscall.SIGTERM
				}
			} else {
				logrus.Debugf("router (not parsed): %+v", msg)
			}
		}
	}
}

func (br *bridge) closeKnxBridge() {
	logrus.Debug("Closing bridge ...")
	br.tunnel.Close()
	br.other.Close()
}
