package browscap

import (
	"fmt"
	"strings"
)

type deviceType uint16
type pointingMethod uint8

const (
	_                  deviceType = 1 << iota
	dtMobilePhone
	dtMobileDevice
	dtTablet
	dtDesktop
	dtTV
	dtConsole
	dtFonePad
	dtEReader
	dtCarEntertainment
)

const (
	_             pointingMethod = 1 << iota
	pmMouse
	pmTouchscreen
	pmJoystick
	pmStylus
	pmClickwheel
	pmTrackpad
	pmTrackball
)

type Browser struct {
	bs *Browscap

	browser  uint32
	platform uint32

	deviceType
	pointingMethod

	browserBits, platformBits int
}

func (b *Browser) setPointingMethod(pm string) {
	switch pm {
	case "mouse":
		b.pointingMethod = pmMouse
	case "touchscreen":
		b.pointingMethod = pmTouchscreen
	case "joystick":
		b.pointingMethod = pmJoystick
	case "stylus":
		b.pointingMethod = pmStylus
	case "clickwheel":
		b.pointingMethod = pmClickwheel
	case "trackpad":
		b.pointingMethod = pmTrackpad
	case "trackball":
		b.pointingMethod = pmTrackball
	}
}

func (d deviceType) String() string {
	switch d {
	case dtMobilePhone, dtMobileDevice:
		return "mobile"
	case dtTablet:
		return "tablet"
	case dtDesktop:
		return "desktop"
	case dtTV:
		return "tv"
	case dtConsole:
		return "console"
	case dtFonePad:
		return "fonepad"
	case dtEReader:
		return "e-reader"
	case dtCarEntertainment:
		return "car-system"
	default:
		return "unknown"
	}
}

func (pm pointingMethod) String() string {
	switch pm {
	case pmMouse:
		return "mouse"
	case pmTouchscreen:
		return "touchscreen"
	case pmJoystick:
		return "joystick"
	case pmStylus:
		return "stylus"
	case pmClickwheel:
		return "clickwheel"
	case pmTrackpad:
		return "trackpad"
	case pmTrackball:
		return "trackball"
	default:
		return "unknown"
	}
}

func (b *Browser) setDeviceType(d string) {
	switch d {
	case "Mobile Phone":
		b.deviceType = dtMobilePhone
	case "Mobile Device":
		b.deviceType = dtMobileDevice
	case "Tablet":
		b.deviceType = dtTablet
	case "Desktop":
		b.deviceType = dtDesktop
	case "TV Device":
		b.deviceType = dtTV
	case "Console":
		b.deviceType = dtConsole
	case "FonePad":
		b.deviceType = dtFonePad
	case "Ebook Reader":
		b.deviceType = dtEReader
	case "Car Entertainment System":
		b.deviceType = dtCarEntertainment
	}
}

func (b *Browser) mapArray(opts []string) {
	b.browserBits = fBrowserBits.GetInt(opts)
	b.platformBits = fPlatformBits.GetInt(opts)

	b.setDeviceType(fDeviceType.GetString(opts))
	b.setPointingMethod(fDevicePointingMethod.GetString(opts))

	value, hash := fPlatform.Hash(opts)
	if _, ok := b.bs.platforms[hash]; !ok {
		b.bs.platforms[hash] = strings.ToLower(value)
	}
	b.platform = hash

	value, hash = fBrowser.Hash(opts)
	if _, ok := b.bs.platforms[hash]; !ok {
		b.bs.browsers[hash] = strings.ToLower(value)
	}
	b.browser = hash
}

func (b Browser) String() string {
	return fmt.Sprintf("{%s@%s %s with %s}", b.bs.browsers[b.browser], b.bs.platforms[b.platform], b.deviceType, b.pointingMethod)
}
