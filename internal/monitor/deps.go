package monitor

import "ddmqtt/internal/ddmrpc"

type DDMRPCClient interface {
	CountMonitors() (int, error)
	GetAssetAttributes(int) (ddmrpc.AssetAttributes, error)
	GetCapabilities(serviceTag string) (ddmrpc.Capabilities, error)
	GetFirmwareVersion(serviceTag string) (string, error)
	GetMonitorActiveHours(serviceTag string) (int, error)
	GetBrightnessLevel(serviceTag string) (int, error)
	SetBrightnessLevel(serviceTag string, value int) error
	GetContrastLevel(serviceTag string) (int, error)
	SetContrastLevel(serviceTag string, value int) error
	GetPower(serviceTag string) (string, error)
	SetPower(serviceTag string, value string) error
	GetActiveInput(serviceTag string) (byte, error)
	SetActiveInput(serviceTag string, input byte) error
	Reset(serviceTag string) error
}
