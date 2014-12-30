package main

import (
	"time"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/av"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")
var log = logger.GetLogger(info.Name)

type Driver struct {
	support.DriverSupport
	devices map[string]*MediaPlayer
}

func NewDriver() (*Driver, error) {

	driver := &Driver{
		devices: make(map[string]*MediaPlayer),
	}

	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize driver: %s", err)
	}

	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export driver: %s", err)
	}

	return driver, nil
}

func (d *Driver) getDevice(upnpDevice goupnp.Device) (*MediaPlayer, error) {
	id := upnpDevice.UDN

	player, ok := d.devices[id]

	var err error

	if !ok {
		player, err = NewMediaPlayer(d, d.Conn, id, upnpDevice)
		d.devices[id] = player
	}

	return player, err
}

func (d *Driver) Start(_ interface{}) error {
	log.Infof("Driver Starting")

	go func() {
		for {
			d.Search()
			time.Sleep(time.Minute)
		}
	}()

	return nil
}

func (d *Driver) Search() error {

	transportClients, errors, err := av.NewAVTransport1Clients()

	if err != nil {
		log.Fatalf("Failed to find transport clients: %s", err)
	}

	for _, e := range errors {
		log.Warningf("Error finding transport clients: %s", e)
	}

	for _, client := range transportClients {
		device, err := d.getDevice(client.ServiceClient.RootDevice.Device)
		if err != nil {
			log.Warningf("Found a transport client, but couldn't create the device. %s", err)
			continue
		}
		device.SetTransportClient(client)
	}

	renderingClients, errors, err := av.NewRenderingControl1Clients()

	if err != nil {
		log.Fatalf("Failed to find transport clients: %s", err)
	}

	for _, e := range errors {
		log.Warningf("Error finding transport clients: %s", e)
	}

	for _, client := range renderingClients {
		device, err := d.getDevice(client.ServiceClient.RootDevice.Device)
		if err != nil {
			log.Warningf("Found a transport client, but couldn't create the device. %s", err)
			continue
		}
		device.SetRenderingClient(client)
	}

	return nil
}
