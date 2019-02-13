package smartmeter

import (
	"time"
)

type P1Packet struct {
	// DSMRVersion is the version information for P1 output (1-3:0.2.8)
	DSMRVersion string
	// Timestamp is the date-time stamp of the P1 message (0-0:1.0.0)
	Timestamp time.Time

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

	// Tariffs contains the client electricity delivery during the tariff, currently exactly two tariff2s.
	Tariffs []Tariff

	// CurrentConsumed contains the actual electricity power delivered in kW (1-0:1.7.0)
	CurrentConsumed float64
	// CurrentProduced contains the actual electricity power produced in kW (1-0:2.7.0)
	CurrentProduced float64

	// NumberOfPowerFailures contains the number of power failures in any phase (0-0:96.7.21)
	NumberOfPowerFailures int
	// NumberOfLongPowerFailures contains the number of long power failures in any phase (0-0:96.7.9)
	NumberOfLongPowerFailures int

	// Phases contains the data of phases, currently exactly three phases.
	Phases []Phase

	// PowerFailureEventLog contains long power failures (1-0:99.97.0)
	PowerFailureEventLog []PowerFailure
}

type Tariff struct {
	// Consumed is the electricity delivered to client during this tariff in kWh (1-0:1.8.1/1-0:1.8.2)
	Consumed float64
	// Produced is the electricity delivered by client during this tariff in kWh (1-0:2.8.1/1-0:2.8.2)
	Produced float64
}

type Phase struct {
	// NumberOfVoltageSage contains the number of voltage sags in this phase (1-0:32.32.0/1-0:52.32.0/1-0:72.32.0)
	NumberOfVoltageSags int
	// NumberOfVoltageSwells contains the number of voltage swells in this phase (1-0:32.36.0/1-0:52.36.0/1-0:72.36.0)
	NumberOfVoltageSwells int

	// InstantaneousVoltage contains the instantaneous voltage in this phase in V (1-0:32.7.0/1-0:52.7.0/1-0:72.7.0)
	InstantaneousVoltage float64
	// InstantaneousCurrent contains the instantaneous current in this phase in A (1-0:31.7.0/1-0:51.7.0/1-0:71.7.0)
	InstantaneousCurrent float64
	// InstantaneousActivePositivePower contains the instantaneous active power (+P) in this phase in kW (1-0:21.7.0/1-0:41.7.0/1-0:61.7.0)
	InstantaneousActivePositivePower float64
	// InstantaneousActivePositivePower contains the instantaneous active power (-P) in this phase in kW (1-0:22.7.0/1-0:42.7.0/1-0:62.7.0)
	InstantaneousActiveNegativePower float64
}

type PowerFailure struct {
	// Timestamp contains the timestamp of the end of the power failure
	Timestamp time.Time
	// Duration contains the length of the power failure
	Duration time.Duration
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
