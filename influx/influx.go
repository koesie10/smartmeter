package influx

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"strconv"
	"time"

	"github.com/koesie10/smartmeter/smartmeter"
)

func NewElectricityPoint(t time.Time, p *smartmeter.P1Packet, measurementName string, tags map[string]string) (*write.Point, error) {
	tags = copyTags(tags)
	tags["equipment_id"] = p.Electricity.EquipmentID
	tags["tariff"] = strconv.Itoa(p.Electricity.Tariff)
	tags["switch_position"] = strconv.Itoa(p.Electricity.SwitchPosition)

	fields := make(map[string]interface{})
	fields["threshold"] = p.Electricity.Threshold

	fields["tariff1_consumed"] = p.Electricity.Tariffs[0].Consumed
	fields["tariff1_produced"] = p.Electricity.Tariffs[0].Produced
	fields["tariff2_consumed"] = p.Electricity.Tariffs[1].Consumed
	fields["tariff2_produced"] = p.Electricity.Tariffs[1].Produced

	fields["current_consumed"] = p.Electricity.CurrentConsumed
	fields["current_produced"] = p.Electricity.CurrentProduced

	fields["number_of_power_failures"] = p.Electricity.NumberOfPowerFailures
	fields["number_of_long_power_failures"] = p.Electricity.NumberOfLongPowerFailures

	return influxdb2.NewPoint(measurementName, tags, fields, t), nil
}

func NewPhasePoint(t time.Time, p *smartmeter.P1Packet, phase int, measurementName string, tags map[string]string) (*write.Point, error) {
	tags = copyTags(tags)
	tags["equipment_id"] = p.Electricity.EquipmentID
	tags["tariff"] = strconv.Itoa(p.Electricity.Tariff)
	tags["switch_position"] = strconv.Itoa(p.Electricity.SwitchPosition)
	// Our phases are 0-indexed in the slice, while they are named in a 1-index fashion.
	// We will use the 1-indexed tags
	tags["phase"] = strconv.Itoa(phase + 1)

	pp := p.Electricity.Phases[phase]

	fields := make(map[string]interface{})
	fields["number_of_voltage_sags"] = pp.NumberOfVoltageSags
	fields["number_of_voltage_swells"] = pp.NumberOfVoltageSwells

	fields["instantaneous_voltage"] = pp.InstantaneousVoltage
	fields["instantaneous_current"] = pp.InstantaneousCurrent
	fields["instantaneous_active_positive_power"] = pp.InstantaneousActivePositivePower
	fields["instantaneous_active_negative_power"] = pp.InstantaneousActiveNegativePower

	return influxdb2.NewPoint(measurementName, tags, fields, t), nil
}

func NewGasPoint(p *smartmeter.P1Packet, measurementName string, tags map[string]string) (*write.Point, error) {
	tags = copyTags(tags)
	tags["equipment_id"] = p.Gas.EquipmentID
	tags["device_type"] = strconv.Itoa(p.Gas.DeviceType)
	tags["valve_position"] = strconv.Itoa(p.Gas.ValvePosition)

	fields := make(map[string]interface{})
	fields["consumed"] = p.Gas.Consumed

	return influxdb2.NewPoint(measurementName, tags, fields, p.Gas.MeasuredAt), nil
}

func copyTags(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return tags
	}
	result := make(map[string]string, len(tags))
	for k, v := range tags {
		result[k] = v
	}
	return result
}
