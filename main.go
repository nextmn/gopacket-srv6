// Copyright 2019-2023 Louis Royer, Takanori Hirano, Gernot Vormayr. All rights reserved.
// Copyright 2012 Google, Inc. All rights reserved.
// Copyright 2009-2011 Andreas Krennmair. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.
// SPDX-License-Identifier: BSD-3-Clause

package gopacket_srv6

import (
	"encoding/binary"
	"fmt"
	"net"

	gopacket "github.com/google/gopacket"
	layers "github.com/google/gopacket/layers"
)

// Register our decoding function to use it instead of gopacket's one
var LayerTypeIPv6Routing = gopacket.OverrideLayerType(47, gopacket.LayerTypeMetadata{Name: "IPv6Routing", Decoder: gopacket.DecodeFunc(decodeIPv6Routing)})

// Copy of ipv6ExtensionBase from gopacket's layers/ip6.go
type gopacketIpv6ExtensionBase struct {
	layers.BaseLayer
	NextHeader   layers.IPProtocol
	HeaderLength uint8
	ActualLength int
}

// Copy of decodeIPv6ExtensionBase from gopacket's layers/ip6.go
func gopacketDecodeIPv6ExtensionBase(data []byte, df gopacket.DecodeFeedback) (i gopacketIpv6ExtensionBase, returnedErr error) {
	if len(data) < 2 {
		df.SetTruncated()
		return gopacketIpv6ExtensionBase{}, fmt.Errorf("Invalid ip6-extension header. Length %d less than 2", len(data))
	}
	i.NextHeader = layers.IPProtocol(data[0])
	i.HeaderLength = data[1]
	i.ActualLength = int(i.HeaderLength)*8 + 8
	if len(data) < i.ActualLength {
		return gopacketIpv6ExtensionBase{}, fmt.Errorf("Invalid ip6-extension header. Length %d less than specified length %d", len(data), i.ActualLength)
	}
	i.Contents = data[:i.ActualLength]
	i.Payload = data[i.ActualLength:]
	return
}

// Copy of IPv6Routing from gopacket's layers/ip6.go
// with the following modifications:
// - Added SRv6 Fields (`LastEntry`, `Flags`, `Tag`)
type IPv6Routing struct {
	gopacketIpv6ExtensionBase
	RoutingType      uint8
	SegmentsLeft     uint8
	Reserved         []byte // only for RoutingType != 4
	LastEntry        uint8
	Flags            uint8
	Tag              uint16
	SourceRoutingIPs []net.IP // Segment List
}

// Copy of LayerType method from gopacket's layers/ip6.go
func (i *IPv6Routing) LayerType() gopacket.LayerType { return LayerTypeIPv6Routing }

// Copy of DecodeFromBytes from PR 1040
func (i *IPv6Routing) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	var err error
	i.gopacketIpv6ExtensionBase, err = gopacketDecodeIPv6ExtensionBase(data, df)
	if err != nil {
		return err
	}
	if len(data) < 8 {
		return fmt.Errorf("IPv6Routing: data too short")
	}
	i.NextHeader = layers.IPProtocol(data[0])
	i.HeaderLength = data[1]
	i.RoutingType = data[2]
	i.SegmentsLeft = data[3]
	if len(data)-8 != int(i.HeaderLength)*8 {
		return fmt.Errorf("IPv6Routing: data length mismatch")
	}
	switch i.RoutingType {
	case 0:
		i.Reserved = data[4:8]
	case 4:
		i.LastEntry = data[4]
		i.Flags = data[5]
		i.Tag = binary.BigEndian.Uint16(data[6:8])
		if len(data)-8 < int(i.LastEntry)*16 {
			return fmt.Errorf("IPv6Routing: data too short")
		}
	default:
		return fmt.Errorf("IPv6Routing: RoutingType %d not supported", i.RoutingType)
	}
	i.SourceRoutingIPs = make([]net.IP, i.LastEntry)
	for j := 0; j < int(i.LastEntry); j++ {
		i.SourceRoutingIPs[j] = net.IP(data[8+j*16 : 8+j*16+16])
	}

	return nil
}

// Copy of decodeIPv6Routing from PR 1040
func decodeIPv6Routing(data []byte, p gopacket.PacketBuilder) error {
	i := &IPv6Routing{}
	err := i.DecodeFromBytes(data, p)
	p.AddLayer(i)
	if err != nil {
		return err
	}
	return p.NextDecoder(i.NextHeader)
}

// Copy of SerializeTo from PR 1040
func (i *IPv6Routing) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	bytes, err := b.PrependBytes(8 + len(i.SourceRoutingIPs)*16)
	if err != nil {
		return err
	}
	if opts.FixLengths {
		i.HeaderLength = uint8(len(i.SourceRoutingIPs) * 2)
	}
	bytes[0] = byte(i.NextHeader)
	bytes[1] = i.HeaderLength
	bytes[2] = i.RoutingType
	bytes[3] = i.SegmentsLeft
	switch i.RoutingType {
	case 0:
		copy(bytes[4:], i.Reserved)
	case 4:
		if opts.FixLengths {
			i.LastEntry = uint8(len(i.SourceRoutingIPs)) - 1
		}
		bytes[4] = i.LastEntry
		bytes[5] = i.Flags
		binary.BigEndian.PutUint16(bytes[6:], i.Tag)
	default:
		return fmt.Errorf("IPv6Routing: RoutingType %d not supported", i.RoutingType)
	}
	for j, ip := range i.SourceRoutingIPs {
		copy(bytes[8+j*16:], ip.To16())
	}
	return nil
}
