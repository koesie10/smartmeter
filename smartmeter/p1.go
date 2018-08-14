package smartmeter

import "time"

type P1Packet struct {
	Electricity Electricity
	Gas         Gas
	Message     Message
	Raw         [][]byte `json:"-"`
}

type Electricity struct {
	// EquipmentID is the equipment identifier (0-0:96.1.1)
	EquipmentID string
	// Tariff indicator for the electricity. (0-0:96.14.0)
	Tariff int
	// SwitchPosition (in/out/enabled) (0-0:96.3.10)
	SwitchPosition int
	// Threshold is the actual electricity threshold in the unit of ThresholdUnit (0-0:17.0.0)
	Threshold float64
	// ThresholdUnit is the unit of the Threshold, usually A or kW
	ThresholdUnit string

	// Tariff1 contains the client electricity delivery during the tariff 1
	Tariff1 Tariff
	// Tariff2 contains the client electricity delivery during the tariff 2
	Tariff2 Tariff

	// CurrentConsumed contains the actual electricity power delivered in kW (1-0:1.7.0)
	CurrentConsumed float64
	// CurrentProduced contains the actual electricity power produced in kW (1-0:2.7.0)
	CurrentProduced float64
}

type Tariff struct {
	// Consumed is the electricity delivered to client during this tariff in kWh (1-0:1.8.1/1-0:1.8.2)
	Consumed float64
	// Produced is the electricity delivered by client during this tariff in kWh (1-0:2.8.1/1-0:2.8.2)
	Produced float64
}

type Message struct {
	// Code is a text message code (0-0:96.13.1)
	Code string
	// Text is a text message (0-0:96.13.0)
	Text string
}

type Gas struct {
	// EquipmentID is the equipment identifier (0-1:96.1.0)
	EquipmentID string
	// DeviceType is the device type (0-1:24.1.0)
	DeviceType int
	// Consumed is the gas delivered to client in m^3 (0-1:24.2.1)
	Consumed float64
	// MeasuredAt is the time at which the gas was measured (0-1:24.2.1)
	MeasuredAt time.Time
	// ValvePosition is the position of the gas valve (on/off/released) (0-1:24.4.0)
	ValvePosition int
}
