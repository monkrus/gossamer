// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package network

import (
	"context"

	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	ma "github.com/multiformats/go-multiaddr"
)

// ConnManager implements connmgr.ConnManager
type ConnManager struct{}

// Notifee is used to monitor changes to a connection
func (cm *ConnManager) Notifee() network.Notifiee {
	nb := new(network.NotifyBundle)

	nb.ListenF = cm.Listen
	nb.ListenCloseF = cm.ListenClose
	nb.ConnectedF = cm.Connected
	nb.DisconnectedF = cm.Disconnected
	nb.OpenedStreamF = cm.OpenedStream
	nb.ClosedStreamF = cm.ClosedStream

	return nb
}

// TagPeer peer
func (*ConnManager) TagPeer(peer.ID, string, int) {}

// UntagPeer peer
func (*ConnManager) UntagPeer(peer.ID, string) {}

// UpsertTag peer
func (*ConnManager) UpsertTag(peer.ID, string, func(int) int) {}

// GetTagInfo peer
func (*ConnManager) GetTagInfo(peer.ID) *connmgr.TagInfo { return &connmgr.TagInfo{} }

// TrimOpenConns peer
func (*ConnManager) TrimOpenConns(ctx context.Context) {}

// Protect peer
func (*ConnManager) Protect(peer.ID, string) {}

// Unprotect peer
func (*ConnManager) Unprotect(peer.ID, string) bool { return false }

// Close peer
func (*ConnManager) Close() error { return nil }

// Listen is called when network starts listening on an address
func (cm *ConnManager) Listen(n network.Network, addr ma.Multiaddr) {
	logger.Trace(
		"[network] Started listening",
		"host", n.LocalPeer(),
		"address", addr,
	)
}

// ListenClose is called when network stops listening on an address
func (cm *ConnManager) ListenClose(n network.Network, addr ma.Multiaddr) {
	logger.Trace(
		"[network] Stopped listening",
		"host", n.LocalPeer(),
		"address", addr,
	)
}

// Connected is called when a connection opened
func (cm *ConnManager) Connected(n network.Network, c network.Conn) {
	logger.Trace(
		"[network] Connected to peer",
		"host", c.LocalPeer(),
		"peer", c.RemotePeer(),
	)
}

// Disconnected is called when a connection closed
func (cm *ConnManager) Disconnected(n network.Network, c network.Conn) {
	logger.Trace(
		"[network] Disconnected from peer",
		"host", c.LocalPeer(),
		"peer", c.RemotePeer(),
	)
}

// OpenedStream is called when a stream opened
func (cm *ConnManager) OpenedStream(n network.Network, s network.Stream) {
	logger.Trace(
		"[network] Opened stream",
		"host", s.Conn().LocalPeer(),
		"peer", s.Conn().RemotePeer(),
		"protocol", s.Protocol(),
	)
}

// ClosedStream is called when a stream closed
func (cm *ConnManager) ClosedStream(n network.Network, s network.Stream) {
	logger.Trace(
		"[network] Closed stream",
		"host", s.Conn().LocalPeer(),
		"peer", s.Conn().RemotePeer(),
		"protocol", s.Protocol(),
	)
}
