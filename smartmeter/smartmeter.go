package smartmeter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "060102150405"

var newGasFormat = regexp.MustCompile(`^0-1:24\.2\.1\((\d+)([SW])\)\((\d{5}\.\d{3})\*m3\)$`)
var oldGasFormat = regexp.MustCompile(`^0-1:24\.3\.0\((\d+)\)`)
var oldGasFormatNextLine = regexp.MustCompile(`^\((\d{5}.\d{3})\)$`)

type SmartMeter struct {
	rw io.ReadWriter
	l  Logger
}

func New(rw io.ReadWriter) (*SmartMeter, error) {
	return &SmartMeter{
		rw: rw,
		l:  NewStderrLog(),
	}, nil
}

func (sm *SmartMeter) Read() (*P1Packet, error) {
	var datagram [][]byte
	var linesRead int
	var startFound bool
	var endFound bool

	scanner := bufio.NewScanner(sm.rw)

	sm.l.Debug("Start reading lines")

	for !startFound || !endFound {
		if !scanner.Scan() {
			return nil, fmt.Errorf("failed to find enough data: %v", scanner.Err())
		}

		line := scanner.Bytes()

		linesRead++

		if bytes.ContainsRune(line, '/') {
			startFound = true
			endFound = false
			datagram = append(datagram, bytes.TrimSpace(line))
		} else if bytes.ContainsRune(line, '!') {
			endFound = true
			datagram = append(datagram, bytes.TrimSpace(line))
		} else {
			datagram = append(datagram, bytes.TrimSpace(line))
		}
	}

	sm.l.Info("Done reading one packet (containing %d lines)", len(datagram))
	sm.l.Debug("Total lines read from serial port: %d", linesRead)
	sm.l.Debug("Constructing packet from raw data")

	return sm.parsePacket(datagram)
}

func (sm *SmartMeter) parsePacket(datagram [][]byte) (*P1Packet, error) {
	p := &P1Packet{
		Raw: datagram,
	}
	var err error

	for i, line := range datagram {
		dataStart := bytes.IndexRune(line, '(')
		dataEnd := bytes.IndexRune(line, ')')
		if dataStart < 0 || dataEnd < 0 {
			continue
		}
		identifier := string(line[:dataStart])
		data := string(line[dataStart+1 : dataEnd])

		switch identifier {
		case "0-0:96.1.1":
			p.Electricity.EquipmentID = data
		case "0-0:96.14.0":
			p.Electricity.Tariff, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as tarrif: %v", data, err)
			}
		case "0-0:96.3.10":
			p.Electricity.SwitchPosition, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as switch position: %v", data, err)
			}
		case "0-0:17.0.0":
			data, p.Electricity.ThresholdUnit = sm.getValueAndUnit(data)
			p.Electricity.Threshold, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as threshold: %v", data, err)
			}
		case "1-0:1.8.1":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kWh" {
				return nil, fmt.Errorf("invalid unit for electricity delivery: %v", unit)
			}
			p.Electricity.Tariff1.Consumed, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity delivery: %v", data, err)
			}
		case "1-0:1.8.2":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kWh" {
				return nil, fmt.Errorf("invalid unit for electricity delivery: %v", unit)
			}
			p.Electricity.Tariff2.Consumed, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity delivery: %v", data, err)
			}
		case "1-0:2.8.1":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kWh" {
				return nil, fmt.Errorf("invalid unit for electricity delivery: %v", unit)
			}
			p.Electricity.Tariff1.Produced, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity delivery: %v", data, err)
			}
		case "1-0:2.8.2":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kWh" {
				return nil, fmt.Errorf("invalid unit for electricity delivery: %v", unit)
			}
			p.Electricity.Tariff2.Produced, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity delivery: %v", data, err)
			}
		case "1-0:1.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kW" {
				return nil, fmt.Errorf("invalid unit for electricity usage: %v", unit)
			}
			p.Electricity.CurrentConsumed, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity usage: %v", data, err)
			}
		case "1-0:2.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kW" {
				return nil, fmt.Errorf("invalid unit for electricity usage: %v", unit)
			}
			p.Electricity.CurrentProduced, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity usage: %v", data, err)
			}
		case "0-1:96.1.0":
			p.Gas.EquipmentID = data
		case "0-1:24.1.0":
			p.Gas.DeviceType, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as device type: %v", data, err)
			}
		case "0-1:24.4.0":
			p.Gas.ValvePosition, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as valve position: %v", data, err)
			}
		case "0-1:24.2.1":
			result := newGasFormat.FindStringSubmatch(string(line))
			if result == nil {
				return nil, fmt.Errorf("failed to parse %v as gas format", string(line))
			}
			p.Gas.Consumed, err = strconv.ParseFloat(result[3], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as gas consumption: %v", result[3], err)
			}
			p.Gas.MeasuredAt, err = time.Parse(dateFormat, result[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as gas measurement time: %v", result[1], err)
			}
		case "0-1:24.3.0":
			result := oldGasFormatNextLine.FindStringSubmatch(string(datagram[i+1]))
			if result == nil {
				return nil, fmt.Errorf("failed to parse %v as gas format", string(line))
			}
			p.Gas.Consumed, err = strconv.ParseFloat(result[1], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as gas consumption: %v", result[1], err)
			}
			result = oldGasFormat.FindStringSubmatch(string(line))
			if result == nil {
				return nil, fmt.Errorf("failed to parse %v as gas format", string(line))
			}
			p.Gas.MeasuredAt, err = time.Parse(dateFormat, result[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as gas measurement time: %v", result[1], err)
			}
		case "0-0:96.13.1":
			p.Message.Code = data
		case "0-0:96.13.0":
			p.Message.Text = data
		}
	}

	return p, nil
}

func (sm *SmartMeter) getValueAndUnit(data string) (string, string) {
	index := strings.LastIndex(data, "*")
	if index < 0 {
		return data, ""
	}

	return data[:index], data [index+1:]
}

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
