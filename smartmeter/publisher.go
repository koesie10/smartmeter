package smartmeter

type Publisher interface {
	Publish(packet *P1Packet) error

	Close() error
}
