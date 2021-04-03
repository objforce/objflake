// Package objflake implements Objflake, a distributed unique ID generator inspired by Twitter's Snowflake.
//
// A objforce ID is composed of
//      5 bits unused
//     39 bits for time in units of 10 msec
//      8 bits for a sequence number
//     12 bits for a machine id
package objflake

import (
	"errors"
	"net"
	"sync"
	"time"
)

// These constants are the bit lengths of Objflake ID parts.
const (
	BitLenTime      = 39                                   // bit length of time
	BitLenSequence  = 8                                    // bit length of sequence number
	BitLenMachineID = 63 - 4 - BitLenTime - BitLenSequence // bit length of machine id
)

// Settings configures Objflake:
//
// StartTime is the time since which the Objflake time is defined as the elapsed time.
// If StartTime is 0, the start time of the Objflake is set to "2014-09-01 00:00:00 +0000 UTC".
// If StartTime is ahead of the current time, Objflake is not created.
//
// MachineID returns the unique ID of the Objflake instance.
// If MachineID returns an error, Objflake is not created.
// If MachineID is nil, default MachineID is used.
// Default MachineID returns the lower 16 bits of the private IP address.
//
// CheckMachineID validates the uniqueness of the machine ID.
// If CheckMachineID returns false, Objflake is not created.
// If CheckMachineID is nil, no validation is done.
type Settings struct {
	StartTime      time.Time
	MachineID      func() (uint16, error)
	CheckMachineID func(uint16) bool
}

// Objflake is a distributed unique ID generator.
type Objflake struct {
	mutex       *sync.Mutex
	startTime   int64
	elapsedTime int64
	sequence    uint16
	machineID   uint16
}

// NewObjflake returns a new Objflake configured with the given Settings.
// NewObjflake returns nil in the following cases:
// - Settings.StartTime is ahead of the current time.
// - Settings.MachineID returns an error.
// - Settings.CheckMachineID returns false.
func NewObjflake(st Settings) *Objflake {
	sf := new(Objflake)
	sf.mutex = new(sync.Mutex)
	sf.sequence = uint16(1<<BitLenSequence - 1)

	if st.StartTime.After(time.Now()) {
		return nil
	}
	if st.StartTime.IsZero() {
		sf.startTime = toObjflakeTime(time.Date(2014, 9, 1, 0, 0, 0, 0, time.UTC))
	} else {
		sf.startTime = toObjflakeTime(st.StartTime)
	}

	var err error
	if st.MachineID == nil {
		sf.machineID, err = lower16BitPrivateIP()
	} else {
		sf.machineID, err = st.MachineID()
	}
	if err != nil || (st.CheckMachineID != nil && !st.CheckMachineID(sf.machineID)) {
		return nil
	}

	return sf
}

// NextID generates a next unique ID.
// After the Objflake time overflows, NextID returns an error.
func (sf *Objflake) NextID() (uint64, error) {
	const maskSequence = uint16(1<<BitLenSequence - 1)

	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	current := currentElapsedTime(sf.startTime)
	if sf.elapsedTime < current {
		sf.elapsedTime = current
		sf.sequence = 0
	} else { // sf.elapsedTime >= current
		sf.sequence = (sf.sequence + 1) & maskSequence
		if sf.sequence == 0 {
			sf.elapsedTime++
			overtime := sf.elapsedTime - current
			time.Sleep(sleepTime((overtime)))
		}
	}

	return sf.toID()
}

const objflakeTimeUnit = 1e7 // nsec, i.e. 10 msec

func toObjflakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / objflakeTimeUnit
}

func currentElapsedTime(startTime int64) int64 {
	return toObjflakeTime(time.Now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UTC().UnixNano()%objflakeTimeUnit)*time.Nanosecond
}

func (sf *Objflake) toID() (uint64, error) {
	if sf.elapsedTime >= 1<<BitLenTime {
		return 0, errors.New("over the time limit")
	}

	return uint64(sf.elapsedTime)<<(BitLenSequence+BitLenMachineID) |
		uint64(sf.sequence)<<BitLenMachineID |
		uint64(sf.machineID), nil
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

// Decompose returns a set of Objflake ID parts.
func Decompose(id uint64) map[string]uint64 {
	const maskSequence = uint64((1<<BitLenSequence - 1) << BitLenMachineID)
	const maskMachineID = uint64(1<<BitLenMachineID - 1)

	msb := id >> 63
	time := id >> (BitLenSequence + BitLenMachineID)
	sequence := id & maskSequence >> BitLenMachineID
	machineID := id & maskMachineID
	return map[string]uint64{
		"id":         id,
		"msb":        msb,
		"time":       time,
		"sequence":   sequence,
		"machine-id": machineID,
	}
}
