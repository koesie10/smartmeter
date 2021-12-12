package mqtt

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type homeAssistantDiscovery struct {
	p      *publisher
	Device *homeAssistantDevice
}

type homeAssistantDevice struct {
	Identifiers  []string `json:"identifiers,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"`
	Model        string   `json:"model,omitempty"`
	Name         string   `json:"name,omitempty"`
}

type homeAssistantEntity struct {
	InternalID string `json:"-"`

	DeviceClass       string `json:"device_class,omitempty"`
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	StateClass        string `json:"state_class,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	ValueTemplate     string `json:"value_template"`

	UniqueID string               `json:"unique_id,omitempty"`
	Device   *homeAssistantDevice `json:"device"`
}

func (p *publisher) publishDiscovery() error {
	if !p.options.HomeAssistant.DiscoveryEnabled {
		return nil
	}

	discovery := homeAssistantDiscovery{
		p: p,
		Device: &homeAssistantDevice{
			Identifiers:  p.options.HomeAssistant.DeviceIdentifiers,
			Manufacturer: p.options.HomeAssistant.DeviceManufacturer,
			Model:        p.options.HomeAssistant.DeviceModel,
			Name:         p.options.HomeAssistant.DeviceName,
		},
	}

	entities := discovery.configureEntities()
	for _, entity := range entities {
		if err := discovery.publishEntity(entity); err != nil {
			p.logger.With(zap.Error(err)).Warnf("Failed to publish entity %s", entity.InternalID)
		}
	}

	return nil
}

func (d *homeAssistantDiscovery) publishEntity(entity *homeAssistantEntity) error {
	topic := fmt.Sprintf(
		"%s/sensor/%s%s/config",
		d.p.options.HomeAssistant.DiscoveryPrefix,
		d.p.options.HomeAssistant.DevicePrefix,
		entity.InternalID,
	)

	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	token := d.p.client.Publish(topic, byte(d.p.options.HomeAssistant.DiscoveryQoS), true, string(data))
	go func(topic string) {
		token.Wait()
		if err := token.Error(); err != nil {
			d.p.logger.With(zap.Error(err)).Warnf("Failed to publish config %s to MQTT", topic)
		}
	}(topic)

	return nil
}

func (d *homeAssistantDiscovery) configureEntity(id string, config *homeAssistantEntity) *homeAssistantEntity {
	config.InternalID = id

	config.StateTopic = d.p.options.Topic
	config.UniqueID = fmt.Sprintf("%s%s", d.p.options.HomeAssistant.UniqueIDPrefix, id)
	config.Device = d.Device

	return config
}
