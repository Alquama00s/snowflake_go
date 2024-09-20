package snowflakego

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	uuidLen = 63
)

// snowflakeSettings represents the configuration of unique ids
type snowflakeSettings struct {
	timeBitlen       int8
	machineIdBitlen  int8
	sequenceNoBitlen int8
	timeUnit         uint64
}

func newSnowflakeSettings(timeLen int8, machineid int8, timeUnit uint64) (*snowflakeSettings, error) {
	seqNo := uuidLen - (timeLen + machineid)
	if seqNo <= 0 {
		return nil, nil
	}
	set := new(snowflakeSettings)
	set.timeBitlen = timeLen
	set.machineIdBitlen = machineid
	set.sequenceNoBitlen = seqNo
	set.timeUnit = timeUnit
	return set, nil
}

type snowFlake struct {
	sfSettings snowflakeSettings
	mutex      *sync.Mutex
	startTs    uint64
	currTs     uint64
	currSeq    uint64
	machineId  uint64
}

func newSnowflake(bitSettings snowflakeSettings, startTime time.Time, machineId uint64) (*snowFlake, error) {
	if startTime.After(time.Now()) {
		return nil, errors.New("start time is ahead of current time")
	}
	if bitSettings.machineIdBitlen+bitSettings.sequenceNoBitlen+bitSettings.timeBitlen != uuidLen ||
		bitSettings.machineIdBitlen == 0 || bitSettings.sequenceNoBitlen == 0 || bitSettings.timeBitlen == 0 {
		return nil, errors.New("malformed bitsettings")
	}
	if maxValue(bitSettings.machineIdBitlen) < machineId {
		return nil, errors.New("invalid machine id for given bitsettings")
	}
	sf := new(snowFlake)
	sf.sfSettings = bitSettings
	sf.startTs = uint64(startTime.UnixNano()) / bitSettings.timeUnit
	sf.machineId = machineId
	sf.mutex = new(sync.Mutex)
	sf.currSeq = 0
	return sf, nil
}

func (sf *snowFlake) nextId() (uint64, error) {
	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	currentTs := time.Now().UnixNano()/int64(sf.sfSettings.timeUnit) - int64(sf.startTs)

	if sf.currTs < uint64(currentTs) {
		sf.currTs = uint64(currentTs)
		sf.currSeq = 0
	} else { //sf.currTs  equal to currentTs
		if sf.currSeq < maxValue(sf.sfSettings.machineIdBitlen) {
			sf.currSeq++
		} else {
			fmt.Println("sequence space exhausted")
			sf.currTs++
			sleepDur := (sf.currTs - uint64(currentTs)) * sf.sfSettings.timeUnit
			time.Sleep(time.Duration(sleepDur))
		}
	}
	if maxValue(sf.sfSettings.timeBitlen) < sf.currTs {
		return 0, errors.New("time over limit")
	}

	return sf.currId(), nil
}

func (sf *snowFlake) currId() uint64 {
	return sf.currTs<<(uint64(sf.sfSettings.sequenceNoBitlen)+uint64(sf.sfSettings.machineIdBitlen)) +
		sf.machineId<<uint64(sf.sfSettings.sequenceNoBitlen) +
		sf.currSeq
}

func maxValue(bitlen int8) uint64 {
	return uint64(math.Pow(2, float64(bitlen)) - 1)
}
