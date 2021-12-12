package mqtt

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type homeAssistantDevice struct {
	Identifiers  []string `json:"identifiers,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"`
	Model        string   `json:"model,omitempty"`
	Name         string   `json:"name,omitempty"`
}

type homeAssistantConfig struct {
	DeviceClass       string `json:"device_class,omitempty"`
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	StateClass        string `json:"state_class,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	ValueTemplate     string `json:"value_template"`

	UniqueID string              `json:"unique_id,omitempty"`
	Device   homeAssistantDevice `json:"device"`
}

func (p *publisher) publishDiscovery() error {
	if !p.options.HomeAssistant.DiscoveryEnabled {
		return nil
	}

	device := homeAssistantDevice{
		Identifiers:  p.options.HomeAssistant.DeviceIdentifiers,
		Manufacturer: p.options.HomeAssistant.DeviceManufacturer,
		Model:        p.options.HomeAssistant.DeviceModel,
		Name:         p.options.HomeAssistant.DeviceName,
	}

	config := homeAssistantConfig{
		DeviceClass:       "power",
		Name:              "Energy Consumption (tariff 1)",
		StateTopic:        p.options.Topic,
		StateClass:        "total_increasing",
		UnitOfMeasurement: "kWh",
		ValueTemplate:     "{{ value_json.Electricity.Tariffs[0].Consumed }}",

		UniqueID: fmt.Sprintf("%s%s", p.options.HomeAssistant.UniqueIDPrefix, "tarrif1_consumed"),
		Device:   device,
	}

	topic := fmt.Sprintf("%s/sensor/%s%s/config", p.options.HomeAssistant.DiscoveryPrefix, p.options.HomeAssistant.DevicePrefix, "tarrif1_consumed")

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	token := p.client.Publish(topic, byte(p.options.HomeAssistant.DiscoveryQoS), true, string(data))
	go func(topic string) {
		token.Wait()
		if err := token.Error(); err != nil {
			p.logger.With(zap.Error(err)).Warnf("Failed to publish config %s to MQTT", topic)
		}
	}(topic)

	return nil
}
