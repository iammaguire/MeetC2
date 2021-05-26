package main

import (
	"fmt"
	"net"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func (packet BeaconICMP) queryServer() {
	for _, c := range packet.data {
		ip := &layers.IPv4{
			SrcIP: net.IP{1, 2, 3, 4},
			DstIP: net.IP{5, 6, 7, 8},
			Id: uint16(c) * IPIDKEY,
		}
		fmt.Println(c)
		buf := gopacket.NewSerializeBuffer()
		opts := gopacket.SerializeOptions{}  // See SerializeOptions for more details.
		err := ip.SerializeTo(buf, opts)
		if err != nil { panic(err) }
		fmt.Println(buf.Bytes())
	}
}