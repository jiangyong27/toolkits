package time

import (
	"fmt"
	"os"
	"sort"
	"time"
)

type TimerCounterParam struct {
	key       string
	startTime *time.Time
	endTime   *time.Time
}

type TimerCounterItem struct {
	total_num       uint64
	total_cost_usec uint64
	max_cost_usec   uint64
	min_cost_usec   uint64
	avg_cost_usec   uint64

	threshold1_num uint64
	threshold2_num uint64
	threshold3_num uint64
}

type TimerCounter struct {
	mertics     map[string]*TimerCounterItem
	thresold1   uint64
	thresold2   uint64
	thresold3   uint64
	outPath     string
	outInterval uint32
	outfile     *os.File
	nowDay      string
	paramChan   chan *TimerCounterParam
}

func NewTimerCounter(path string, interval uint32) *TimerCounter {
	tc := &TimerCounter{
		mertics:     make(map[string]*TimerCounterItem),
		paramChan:   make(chan *TimerCounterParam, 1000),
		thresold1:   100,
		thresold2:   500,
		thresold3:   1000,
		outPath:     path,
		outInterval: interval,
	}

	tc.change()
	go tc.run()
	return tc
}

func (tc *TimerCounter) AddStat(key string, start *time.Time, end *time.Time) {
	tc.paramChan <- &TimerCounterParam{
		key:       key,
		startTime: start,
		endTime:   end,
	}
}

func (tc *TimerCounter) change() {
	nowDay := time.Now().Format("2006-01-02")
	if nowDay == tc.nowDay {
		return
	}
	if tc.outfile != nil {
		tc.outfile.Close()
	}

	tc.nowDay = nowDay
	outfile, err := os.OpenFile(tc.outPath+"."+tc.nowDay, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		panic(err.Error())
	}
	tc.outfile = outfile
	tc.nowDay = nowDay

	delDay := time.Unix(time.Now().Unix()-3600*24*7, 0).Format("2006-01-02")
	os.Remove(tc.outPath + "." + delDay)

}

func (tc *TimerCounter) add(param *TimerCounterParam) {
	mertic, ok := tc.mertics[param.key]
	if !ok {
		mertic = new(TimerCounterItem)
		tc.mertics[param.key] = mertic

	}

	mertic.total_num += 1
	if param.startTime == nil {
		return
	}

	if param.endTime == nil {
		now := time.Now()
		param.endTime = &now
	}

	cost_us := uint64(param.endTime.Sub(*param.startTime).Nanoseconds() / 1000)
	cost_ms := uint64(cost_us / 1000)
	if cost_ms >= tc.thresold3 {
		mertic.threshold3_num += 1
	} else if cost_ms >= tc.thresold2 {
		mertic.threshold2_num += 1
	} else if cost_ms >= tc.thresold1 {
		mertic.threshold1_num += 1
	}

	mertic.total_cost_usec += cost_us
	if cost_us > mertic.max_cost_usec {
		mertic.max_cost_usec = cost_us
	}

	if cost_us < mertic.min_cost_usec {
		mertic.min_cost_usec = cost_us
	}
}

func (tc *TimerCounter) output() {
	keys := make([]string, 0)
	for key, _ := range tc.mertics {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	szTmpStr1 := fmt.Sprintf(">%.3fms", float32(tc.thresold1))
	szTmpStr2 := fmt.Sprintf(">%.3fms", float32(tc.thresold2))
	szTmpStr3 := fmt.Sprintf(">%.3fms", float32(tc.thresold3))

	buf := fmt.Sprintf("[%s] ===========================================================================\n",
		time.Now().Format("2006-01-02 15:04:05"))

	buf += fmt.Sprintf("%-20s|%15s|%15s|%15s|%15s|%15s|%11s|%11s|%11s|\n",
		"STATKEY", "TOTAL(N)", "TOTAL(S)", "AVG(ms)", "MAX(ms)", "MIN(ms)", szTmpStr1, szTmpStr2, szTmpStr3)
	for _, key := range keys {
		mertic, ok := tc.mertics[key]
		if !ok {
			continue
		}
		buf += fmt.Sprintf("%-20s|%15d|%15.2f|%15.2f|%15.2f|%15.2f|%11d|%11d|%11d|\n",
			key, mertic.total_num,
			float32(mertic.total_cost_usec)/float32(1000)/float32(1000),
			float32(mertic.total_cost_usec)/float32(1000)/float32(mertic.total_num),
			float32(mertic.max_cost_usec)/float32(1000),
			float32(mertic.min_cost_usec)/float32(1000),
			mertic.threshold1_num,
			mertic.threshold2_num,
			mertic.threshold3_num)
	}
	tc.mertics = make(map[string]*TimerCounterItem)
	tc.outfile.WriteString(buf)
}

func (tc *TimerCounter) run() {
	ticker := time.NewTicker(time.Duration(tc.outInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tc.output()
		case param := <-tc.paramChan:
			tc.add(param)
		}
	}
}
