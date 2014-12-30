package main

import (
	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/av"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-ninja/model"
)

type MediaPlayer struct {
	player          *devices.MediaPlayerDevice
	transportClient *av.AVTransport1
	renderingClient *av.RenderingControl1
}

func NewMediaPlayer(driver ninja.Driver, conn *ninja.Connection, id string, upnpDevice goupnp.Device) (*MediaPlayer, error) {

	device := &MediaPlayer{}

	name := upnpDevice.FriendlyName
	if name == "" && upnpDevice.Manufacturer != "" && upnpDevice.ModelName != "" {
		name = upnpDevice.Manufacturer + " " + upnpDevice.ModelName
	}

	signatures := map[string]string{
		"upnp:deviceType":       upnpDevice.DeviceType,
		"upnp:friendlyName":     upnpDevice.FriendlyName,
		"upnp:manufacturer":     upnpDevice.Manufacturer,
		"upnp:manufacturerURL":  upnpDevice.ManufacturerURL.URL.String(),
		"upnp:modelDescription": upnpDevice.ModelDescription,
		"upnp:modelName":        upnpDevice.ModelName,
		"upnp:modelNumber":      upnpDevice.ModelNumber,
		"upnp:modelURL":         upnpDevice.ModelURL.URL.String(),
		"upnp:serialNumber":     upnpDevice.SerialNumber,
		"upnp:UDN":              upnpDevice.UDN,
		"upnp:UPC":              upnpDevice.UPC,
		"ninja:thingType":       "mediaplayer",
	}

	for name, val := range signatures {
		if val == "" {
			delete(signatures, name)
		}
	}

	player, err := devices.CreateMediaPlayerDevice(driver, &model.Device{
		NaturalID:     id,
		NaturalIDType: "upnp",
		Name:          &id,
		Signatures:    &signatures,
	}, conn)

	if err != nil {
		return nil, err
	}

	device.player = player

	return device, nil
}

func (d *MediaPlayer) SetTransportClient(transportClient *av.AVTransport1) error {
	if d.transportClient == nil {

		d.player.ApplyPlayPause = d.applyPlayPause
		d.player.ApplyStop = d.applyStop
		d.player.ApplyPlaylistJump = d.applyPlaylistJump
		if err := d.player.EnableControlChannel([]string{"playing", "paused", "stopped"}); err != nil {
			d.player.Log().Fatalf("Failed to enable control channel: %s", err)
		}
	}

	d.transportClient = transportClient

	return nil
}

func (d *MediaPlayer) SetRenderingClient(renderingClient *av.RenderingControl1) error {

	if d.renderingClient == nil {
		d.player.ApplyVolume = d.applyVolume
		if err := d.player.EnableVolumeChannel(true); err != nil {
			d.player.Log().Fatalf("Failed to enable volume channel: %s", err)
		}
	}

	d.renderingClient = renderingClient

	return nil
}

func (d *MediaPlayer) applyPlaylistJump(delta int) error {
	if delta > 0 {
		return d.transportClient.Next(0)
	}
	return d.transportClient.Previous(0)
}

func (d *MediaPlayer) applyPlayPause(play bool) error {

	if play {
		return d.transportClient.Play(0, "1")
	}
	return d.transportClient.Pause(0)
}

func (d *MediaPlayer) applyStop() error {
	return d.transportClient.Stop(0)
}

func (d *MediaPlayer) applyVolume(state *channels.VolumeState) (err error) {
	d.player.Log().Infof("applyVolume called, volume %v", state)

	if state.Muted != nil {
		err1 := d.renderingClient.SetMute(0, "Master", *state.Muted)
		if err1 != nil {
			err = err1
		}
	}
	if state.Level != nil {
		err2 := d.renderingClient.SetVolume(0, "Master", uint16(*state.Level*100.00))
		if err == nil && err2 != nil {
			err = err2
		}
	}

	return
}
