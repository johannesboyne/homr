package main

import (
	"fmt"
	"log"
	"time"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

func initiateBluetoothScanning() {
	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	peripherals := make(chan gatt.Peripheral, 1024)
	done := make(chan bool, 1)

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered(peripherals)),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected(done)),
	)

	err = d.Init(onStateChanged)
	if err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(30 * time.Second)
		d.StopScanning()

		for {
			p := <-peripherals
			fmt.Printf("try connecting to %s (%v)\n", p.ID(), len(peripherals))
			p.Device().Connect(p)
			select {
			case <-done:
				fmt.Println("done")
			case <-time.After(3 * time.Second):
				fmt.Println("timeout for:", p.ID())
				if len(peripherals) == 0 {
					break
				}
			}
			if len(peripherals) == 0 {
				break
			}
		}

		fmt.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
	}
}

func onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(peripherals chan gatt.Peripheral) func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	return func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
		fmt.Printf("found p: %s\n", p.ID())
		peripherals <- p
	}
}
func onPeriphConnected(p gatt.Peripheral, err error) {
	defer p.Device().CancelConnection(p)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("--- %s ---\n", p.ID())
	fmt.Printf("Connected:  %s\n", p.ID())

	fmt.Printf("set mtu\n")
	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}

	// Discovery services
	fmt.Println("discover services")
	ds, err := p.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return
	}
	fmt.Println("discovered services #", len(ds))

	for _, s := range ds {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		fmt.Println("discover characteristics")
		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}
		fmt.Println("discovered characteristics #", len(cs))

		for _, c := range cs {
			msg := "  Characteristic  " + c.UUID().String()
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
			fmt.Println(msg)
		}
	}
	fmt.Printf("----------------\n")
	time.Sleep(5 * time.Second)
}

func onPeriphDisconnected(done chan bool) func(p gatt.Peripheral, err error) {
	return func(p gatt.Peripheral, err error) {
		fmt.Printf("Disconnected %s\n", p.ID())
		done <- true
	}
}
