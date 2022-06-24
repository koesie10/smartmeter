package mqtt

import "fmt"

// https://developers.home-assistant.io/docs/core/entity/sensor/#long-term-statistics
func (d *homeAssistantDiscovery) configureEntities() []*homeAssistantEntity {
	var result []*homeAssistantEntity

	result = append(result,
		d.configureEntity("electricity_equipment_id", &homeAssistantEntity{
			Name:          "Electricity Equipment ID",
			ValueTemplate: "{{ value_json.Electricity.EquipmentID }}",
		}),
	)

	for tariff := 0; tariff < 2; tariff++ {
		result = append(
			result,
			d.configureEntity(fmt.Sprintf("tariff%d_consumed", tariff+1), &homeAssistantEntity{
				DeviceClass:       "energy",
				Name:              fmt.Sprintf("Energy Consumption (tariff %d)", tariff+1),
				StateClass:        "total",
				UnitOfMeasurement: "kWh",
				ValueTemplate:     fmt.Sprintf("{{ value_json.Electricity.Tariffs[%d].Consumed }}", tariff),
			}),
			d.configureEntity(fmt.Sprintf("tariff%d_produced", tariff+1), &homeAssistantEntity{
				DeviceClass:       "energy",
				Name:              fmt.Sprintf("Energy Production (tariff %d)", tariff+1),
				StateClass:        "total",
				UnitOfMeasurement: "kWh",
				ValueTemplate:     fmt.Sprintf("{{ value_json.Electricity.Tariffs[%d].Produced }}", tariff),
			}),
		)
	}

	result = append(result,
		d.configureEntity("tariff", &homeAssistantEntity{
			Name:          "Energy Tariff",
			ValueTemplate: "{{ value_json.Electricity.Tariff }}",
		}),

		d.configureEntity("current_consumption", &homeAssistantEntity{
			DeviceClass:       "power",
			Name:              "Energy Consumption",
			StateClass:        "measurement",
			UnitOfMeasurement: "kW",
			ValueTemplate:     "{{ value_json.Electricity.CurrentConsumed }}",
		}),
		d.configureEntity("current_production", &homeAssistantEntity{
			DeviceClass:       "power",
			Name:              "Energy Production",
			StateClass:        "measurement",
			UnitOfMeasurement: "kW",
			ValueTemplate:     "{{ value_json.Electricity.CurrentProduced }}",
		}),
	)

	for phase := 0; phase < 3; phase++ {
		result = append(
			result,
			d.configureEntity(fmt.Sprintf("phase%d_instantaneous_voltage", phase+1), &homeAssistantEntity{
				DeviceClass:       "voltage",
				Name:              fmt.Sprintf("Instantaneous voltage (phase %d)", phase+1),
				StateClass:        "measurement",
				UnitOfMeasurement: "V",
				ValueTemplate:     fmt.Sprintf("{{ value_json.Electricity.Phases[%d].InstantaneousVoltage }}", phase),
			}),
			d.configureEntity(fmt.Sprintf("phase%d_instantaneous_current", phase+1), &homeAssistantEntity{
				DeviceClass:       "current",
				Name:              fmt.Sprintf("Instantaneous current (phase %d)", phase+1),
				StateClass:        "measurement",
				UnitOfMeasurement: "A",
				ValueTemplate:     fmt.Sprintf("{{ value_json.Electricity.Phases[%d].InstantaneousCurrent }}", phase),
			}),
		)
	}

	result = append(result,
		d.configureEntity("gas_equipment_id", &homeAssistantEntity{
			Name:          "Gas Equipment ID",
			ValueTemplate: "{{ value_json.Gas.EquipmentID }}",
		}),

		d.configureEntity("gas_consumed", &homeAssistantEntity{
			DeviceClass:       "gas",
			Name:              "Gas Consumed",
			StateClass:        "total",
			UnitOfMeasurement: "m³",
			ValueTemplate:     "{{ value_json.Gas.Consumed }}",
		}),
	)

	return result
}
