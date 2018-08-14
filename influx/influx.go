package influx

import (
	"strconv"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/koesie10/smartmeter/smartmeter"
)

func NewElectricityPoint(t time.Time, p *smartmeter.P1Packet, measurement string, tags map[string]string) (*client.Point, error) {
	tags = copyTags(tags)
	tags["equipment_id"] = p.Electricity.EquipmentID
	tags["tariff"] = strconv.Itoa(p.Electricity.Tariff)
	tags["switch_position"] = strconv.Itoa(p.Electricity.SwitchPosition)

	fields := make(map[string]interface{})
	fields["threshold"] = p.Electricity.Threshold

	fields["tariff1_consumed"] = p.Electricity.Tariff1.Consumed
	fields["tariff1_produced"] = p.Electricity.Tariff1.Produced
	fields["tariff2_consumed"] = p.Electricity.Tariff2.Consumed
	fields["tariff2_produced"] = p.Electricity.Tariff2.Produced

	fields["current_consumed"] = p.Electricity.CurrentConsumed
	fields["current_produced"] = p.Electricity.CurrentProduced

	return client.NewPoint(measurement, tags, fields, t)
}

func NewGasPoint(p *smartmeter.P1Packet, measurement string, tags map[string]string) (*client.Point, error) {
	tags = copyTags(tags)
	tags["equipment_id"] = p.Gas.EquipmentID
	tags["device_type"] = strconv.Itoa(p.Gas.DeviceType)
	tags["valve_position"] = strconv.Itoa(p.Gas.ValvePosition)

	fields := make(map[string]interface{})
	fields["consumed"] = p.Gas.Consumed

	return client.NewPoint(measurement, tags, fields, p.Gas.MeasuredAt)
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
