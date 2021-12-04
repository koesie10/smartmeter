package debugjson

import (
	"encoding/json"
	"fmt"
	"github.com/koesie10/smartmeter/smartmeter"
)

var _ smartmeter.Publisher = (*publisher)(nil)

type publisher struct{}

func NewPublisher() (smartmeter.Publisher, error) {
	return &publisher{}, nil
}

func (p *publisher) Publish(packet *smartmeter.P1Packet) error {
	data, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to marshal packet to JSON: %w", err)
	}

	fmt.Printf("JSON DEBUG: %s\n", string(data))

	return nil
}

func (p *publisher) Close() error {
	return nil
}
