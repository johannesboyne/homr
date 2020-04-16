package main

import (
	"flag"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/labstack/gommon/log"
)

const wifiInterface = "mon0"

var bluetoothFlag bool

func init() {
	flag.BoolVar(&bluetoothFlag, "bluetooth scanning", false, "")
}

type Packet struct {
	Mac       string    `json:"mac"`
	RSSI      int       `json:"rssi"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	flag.Parse()
	packets := []Packet{}
	// Open device
	handle, err := pcap.OpenLive(wifiInterface, 2048, false, pcap.BlockForever)
	if err != nil {
		return
	}
	defer handle.Close()
	startTime := time.Now()

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if time.Since(startTime).Seconds() > 30 {
			handle.Close()
			return
		}
		// Process packet here
		// fmt.Println(packet.String())
		address := ""
		rssi := 0
		transmitter := ""
		receiver := ""
		for _, layer := range packet.Layers() {
			switch layer.LayerType() {
			case layers.LayerTypeRadioTap:
				rt := layer.(*layers.RadioTap)
				rssi = int(rt.DBMAntennaSignal)
			case layers.LayerTypeDot11:
				dot11 := layer.(*layers.Dot11)
				receiver = dot11.Address1.String()
				transmitter = dot11.Address2.String()
				// if doIgnoreRandomizedMacs && utils.IsMacRandomized(transmitter) {
				// 	break
				// }
				// if doAllPackets || receiver == "ff:ff:ff:ff:ff:ff" {
				// 	address = transmitter
				// }
			}
		}
		if address != "" && rssi != 0 {
			newPacket := Packet{
				Mac:       address,
				RSSI:      rssi,
				Timestamp: time.Now(),
			}
			packets = append(packets, newPacket)
			log.Debugf("%s) %s->%s: %d", "Channel 1", transmitter, receiver, newPacket.RSSI)
		}
	}

	if bluetoothFlag {
		initiateBluetoothScanning()
	}
}
