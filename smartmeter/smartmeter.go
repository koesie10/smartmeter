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
	r io.Reader
	l Logger
}

func New(r io.Reader) (*SmartMeter, error) {
	return &SmartMeter{
		r: r,
		l: NewStderrLog(),
	}, nil
}

func (sm *SmartMeter) Read() (*P1Packet, error) {
	var datagram [][]byte
	var linesRead int
	var startFound bool
	var endFound bool

	scanner := bufio.NewScanner(sm.r)

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

	return sm.parsePacket(datagram)
}

func (sm *SmartMeter) parsePacket(datagram [][]byte) (*P1Packet, error) {
	p := &P1Packet{
		Timestamp: time.Now(),
		Electricity: Electricity{
			Tariffs: make([]Tariff, 2),
			Phases: make([]Phase, 3),
		},
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
		case "1-3:0.2.8":
			p.DSMRVersion = data
		case "0-0:1.0.0":
			p.Timestamp, err = time.ParseInLocation(dateFormat, data[:len(data)-1], time.Local)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as timestamp: %v", data, err)
			}
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
			p.Electricity.Tariffs[0].Consumed, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity delivery: %v", data, err)
			}
		case "1-0:1.8.2":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kWh" {
				return nil, fmt.Errorf("invalid unit for electricity delivery: %v", unit)
			}
			p.Electricity.Tariffs[1].Consumed, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity delivery: %v", data, err)
			}
		case "1-0:2.8.1":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kWh" {
				return nil, fmt.Errorf("invalid unit for electricity delivery: %v", unit)
			}
			p.Electricity.Tariffs[0].Produced, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as electricity delivery: %v", data, err)
			}
		case "1-0:2.8.2":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kWh" {
				return nil, fmt.Errorf("invalid unit for electricity delivery: %v", unit)
			}
			p.Electricity.Tariffs[1].Produced, err = strconv.ParseFloat(data, 64)
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
		case "0-0:96.7.21":
			p.Electricity.NumberOfPowerFailures, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of power failures: %v", data, err)
			}
		case "0-0:96.7.9":
			p.Electricity.NumberOfLongPowerFailures, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of long power failures: %v", data, err)
			}
		case "1-0:32.32.0":
			p.Electricity.Phases[0].NumberOfVoltageSags, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of voltage sags in phase L1: %v", data, err)
			}
		case "1-0:52.32.0":
			p.Electricity.Phases[1].NumberOfVoltageSags, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of voltage sags in phase L2: %v", data, err)
			}
		case "1-0:72.32.0":
			p.Electricity.Phases[2].NumberOfVoltageSags, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of voltage sags in phase L3: %v", data, err)
			}
		case "1-0:32.36.0":
			p.Electricity.Phases[0].NumberOfVoltageSwells, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of voltage swells in phase L1: %v", data, err)
			}
		case "1-0:52.36.0":
			p.Electricity.Phases[1].NumberOfVoltageSwells, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of voltage swells in phase L2: %v", data, err)
			}
		case "1-0:72.36.0":
			p.Electricity.Phases[2].NumberOfVoltageSwells, err = strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of voltage swells in phase L3: %v", data, err)
			}
		case "1-0:32.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "V" {
				return nil, fmt.Errorf("invalid unit for instantaneous voltage in phase L1: %v", unit)
			}
			p.Electricity.Phases[0].InstantaneousVoltage, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous voltage in phase L1: %v", data, err)
			}
		case "1-0:52.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "V" {
				return nil, fmt.Errorf("invalid unit for instantaneous voltage in phase L2: %v", unit)
			}
			p.Electricity.Phases[1].InstantaneousVoltage, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous voltage in phase L2: %v", data, err)
			}
		case "1-0:72.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "V" {
				return nil, fmt.Errorf("invalid unit for instantaneous voltage in phase L3: %v", unit)
			}
			p.Electricity.Phases[2].InstantaneousVoltage, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous voltage in phase L3: %v", data, err)
			}
		case "1-0:31.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "A" {
				return nil, fmt.Errorf("invalid unit for instantaneous current in phase L1: %v", unit)
			}
			p.Electricity.Phases[0].InstantaneousCurrent, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous current in phase L1: %v", data, err)
			}
		case "1-0:51.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "A" {
				return nil, fmt.Errorf("invalid unit for instantaneous current in phase L2: %v", unit)
			}
			p.Electricity.Phases[1].InstantaneousCurrent, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous current in phase L2: %v", data, err)
			}
		case "1-0:71.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "A" {
				return nil, fmt.Errorf("invalid unit for instantaneous current in phase L3: %v", unit)
			}
			p.Electricity.Phases[2].InstantaneousCurrent, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous current in phase L3: %v", data, err)
			}
		case "1-0:21.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kW" {
				return nil, fmt.Errorf("invalid unit for instantaneous active power P+ in phase L1: %v", unit)
			}
			p.Electricity.Phases[0].InstantaneousActivePositivePower, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous active power P+ in phase L1: %v", data, err)
			}
		case "1-0:41.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kW" {
				return nil, fmt.Errorf("invalid unit for instantaneous active power P+ in phase L2: %v", unit)
			}
			p.Electricity.Phases[1].InstantaneousActivePositivePower, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous active power P+ in phase L2: %v", data, err)
			}
		case "1-0:61.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kW" {
				return nil, fmt.Errorf("invalid unit for instantaneous active power P+ in phase L3: %v", unit)
			}
			p.Electricity.Phases[2].InstantaneousActivePositivePower, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous active power P+ in phase L3: %v", data, err)
			}
		case "1-0:22.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kW" {
				return nil, fmt.Errorf("invalid unit for instantaneous active power P- in phase L1: %v", unit)
			}
			p.Electricity.Phases[0].InstantaneousActiveNegativePower, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous active power P- in phase L1: %v", data, err)
			}
		case "1-0:42.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kW" {
				return nil, fmt.Errorf("invalid unit for instantaneous active power P- in phase L2: %v", unit)
			}
			p.Electricity.Phases[1].InstantaneousActiveNegativePower, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous active power P- in phase L2: %v", data, err)
			}
		case "1-0:99.97.0":
			numberOfPowerFailures, err := strconv.Atoi(data)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as number of power failures: %v", data, err)
			}

			index := dataEnd + 1
			lineData := line[index:]

			nextDataStart := bytes.IndexRune(lineData, '(')
			nextDataEnd := bytes.IndexRune(lineData, ')')
			if nextDataStart < 0 || nextDataEnd < 0 {
				continue
			}
			data := string(lineData[nextDataStart+1 : nextDataEnd])

			if data != "0-0:96.7.19" {
				return nil, fmt.Errorf("invalid power failure event log data format %v", data)
			}

			for i := 0; i < numberOfPowerFailures; i++ {
				item := PowerFailure{}

				index = nextDataEnd + 1
				lineData = lineData[index:]

				nextDataStart = bytes.IndexRune(lineData, '(')
				nextDataEnd = bytes.IndexRune(lineData, ')')
				if nextDataStart < 0 || nextDataEnd < 0 {
					continue
				}

				data := string(lineData[nextDataStart+1 : nextDataEnd])

				item.Timestamp, err = time.ParseInLocation(dateFormat, data[:len(data)-1], time.Local)
				if err != nil {
					return nil, fmt.Errorf("failed to parse %v as power failure timestamp: %v", data, err)
				}

				index = nextDataEnd + 1
				lineData = lineData[index:]

				nextDataStart = bytes.IndexRune(lineData, '(')
				nextDataEnd = bytes.IndexRune(lineData, ')')
				if nextDataStart < 0 || nextDataEnd < 0 {
					continue
				}

				data = string(lineData[nextDataStart+1 : nextDataEnd])

				data, unit := sm.getValueAndUnit(data)
				if unit != "s" {
					return nil, fmt.Errorf("invalid unit for power failure event log duration: %v", unit)
				}

				duration, err := strconv.Atoi(data)
				if err != nil {
					return nil, fmt.Errorf("failed to parse %v as power failure duration: %v", data, err)
				}

				item.Duration = time.Duration(duration) * time.Second

				p.Electricity.PowerFailureEventLog = append(p.Electricity.PowerFailureEventLog, item)
			}
		case "1-0:62.7.0":
			data, unit := sm.getValueAndUnit(data)
			if unit != "kW" {
				return nil, fmt.Errorf("invalid unit for instantaneous active power P- in phase L3: %v", unit)
			}
			p.Electricity.Phases[2].InstantaneousActiveNegativePower, err = strconv.ParseFloat(data, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %v as instantaneous active power P- in phase L3: %v", data, err)
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
