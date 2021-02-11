package histogram

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gosuri/uitable"
	"github.com/i582/cfmt"
)

// Sample histogram
type Sample struct {
	sync.Mutex
	Ticker         *time.Ticker
	StartTime      time.Time
	TimeSleep      int
	Count          int64
	Times          timeSlice
	MetricBodySize struct {
		P1  int64
		P10 int64
		P50 int64
		P97 int64
		P99 int64
	}
	MetricReqSec struct {
		P1  int64
		P10 int64
		P50 int64
		P97 int64
		P99 int64
	}
}

// Metrics histogram structure
type Metrics struct {
	sync.Mutex
	Time struct {
		Avg    time.Duration
		P1     int64
		P10    int64
		P50    int64
		P97    int64
		P99    int64
		StdDev time.Duration
		Max    time.Duration
		Min    time.Duration
	}
}

type timeSlice []time.Duration

func (ts timeSlice) Len() int           { return len(ts) }
func (ts timeSlice) Less(i, j int) bool { return int64(ts[i]) < int64(ts[j]) }
func (ts timeSlice) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }

// New to start histogram
func New() *Sample {
	return &Sample{
		StartTime: time.Now(),
	}
}

// SetTimeSleep to set time to sleep
func (s *Sample) SetTimeSleep(d int) *Sample {
	s.Ticker = time.NewTicker(time.Duration(d) * time.Second)
	s.TimeSleep = d

	return s
}

// AddTime to add sample time duration
func (s *Sample) AddTime(t time.Duration) *Sample {
	s.Times = append(s.Times, t)
	atomic.AddInt64(&s.Count, 1)

	return s
}

// AddSize to add size from sample data
func (s *Sample) AddSize(t time.Duration, size int64) *Sample {
	atomic.AddInt64(&s.Count, 1)

	p1 := float64(s.TimeSleep) * 0.01
	p10 := float64(s.TimeSleep) * 0.10
	p50 := float64(s.TimeSleep) / 2
	p97 := float64(s.TimeSleep) * 0.97
	p99 := float64(s.TimeSleep) * 0.99

	s.Lock()
	if t.Seconds() <= p1 {
		s.MetricReqSec.P1++
		s.MetricBodySize.P1 = size
	} else if t.Seconds() > p1 && t.Seconds() <= p10 {
		if s.MetricReqSec.P10 == 0 {
			s.MetricReqSec.P10 = s.MetricReqSec.P1
		}
		s.MetricReqSec.P10++
		s.MetricBodySize.P10 = size
	} else if t.Seconds() > p10 && t.Seconds() <= p50 {
		if s.MetricReqSec.P50 == 0 {
			s.MetricReqSec.P50 = s.MetricReqSec.P10
		}
		s.MetricReqSec.P50++
		s.MetricBodySize.P50 = size
	} else if t.Seconds() > p50 && t.Seconds() <= p97 {
		if s.MetricReqSec.P50 == 0 {
			s.MetricReqSec.P97 = s.MetricReqSec.P50
		}
		s.MetricReqSec.P97++
		s.MetricBodySize.P97 = size
	} else if t.Seconds() > p97 && t.Seconds() <= p99 {
		if s.MetricReqSec.P99 == 0 {
			s.MetricReqSec.P99 = s.MetricReqSec.P97
		}
		s.MetricReqSec.P99++
		s.MetricBodySize.P99 = size
	}
	s.Unlock()

	return s
}

// CalcLatency to calculate sample data
func (s *Sample) CalcLatency() {
	m := &Metrics{}
	if atomic.LoadInt64(&s.Count) == 0 {
		return
	}

	s.Lock()

	times := make(timeSlice, s.Count)
	copy(times, s.Times[:s.Count])
	sort.Sort(times)

	s.Unlock()

	m.Time.Avg = times.avg()
	m.Time.Min = times.min()
	m.Time.Max = times.max()
	m.Time.StdDev = times.stdDev()
	m.Time.P1 = times.percentile(0.01)
	m.Time.P10 = times.percentile(0.10)
	m.Time.P50 = times[len(times)/2].Milliseconds()
	m.Time.P97 = times.percentile(0.97)
	m.Time.P99 = times.percentile(0.99)

	cliPrintLatency(m)
}

// CalcReqBytes to calculate sample data
func (s *Sample) CalcReqBytes() {
	cliPrintReqBytes(s)
}

func (ts timeSlice) avg() (total time.Duration) {
	for _, t := range ts {
		total += t
	}
	return time.Duration(int(total) / ts.Len())
}

func avgReq(d []int64) (total int64) {
	for _, t := range d {
		total += t
	}

	return total / int64(len(d))
}

func (ts timeSlice) stdDev() time.Duration {
	m := ts.avg()
	s := 0.00

	for _, t := range ts {
		s += math.Pow(float64(m-t), 2)
	}

	msq := s / float64(ts.Len())

	return time.Duration(math.Sqrt(msq))
}

func stdDevReq(d []int64) float64 {
	m := avgReq(d)
	s := 0.00

	for _, t := range d {
		s += math.Pow(float64(m-t), 2)
	}

	msq := s / float64(len(d))

	return math.Sqrt(msq)
}

func (ts timeSlice) min() time.Duration {
	return ts[0]
}

func (ts timeSlice) max() time.Duration {
	return ts[ts.Len()-1]
}

func (ts timeSlice) percentile(p float64) int64 {
	tp := int(float64(ts.Len())*p+0.5) - 1
	if tp < 0 {
		tmp := ts[0]
		return tmp.Milliseconds()
	}

	tmp := ts[tp]
	return tmp.Milliseconds()
}

func countDuration(c int64) (t string) {
	if c < 1000 {
		t = fmt.Sprintf("%dms", c)
	} else if c >= 1000 {
		t = fmt.Sprintf("%.2fs", float32(c/1000))
	}

	return t
}

func countRequest(c int64) string {
	var total string
	if c < 1000 {
		total = fmt.Sprintf("%d", c)
	} else if c >= 1000 {
		total = fmt.Sprintf("%dk", c/1000)
	}

	return total
}

func countReadBytes(s int64) (size string) {
	if s >= 1000000 {
		size = fmt.Sprintf("%.1f MB", float32(s/1000000))
	} else if s >= 1000 && s < 1000000 {
		size = fmt.Sprintf("%d KB", s/1000)
	} else {
		size = fmt.Sprintf("%d B", s)
	}

	return size
}

func cliPrintLatency(m *Metrics) {
	table := uitable.New()
	table.MaxColWidth = 50

	table.AddRow(
		cfmt.Sprintf("{{%s}}::bold", "STAT"),
		cfmt.Sprintf("{{%s}}::bold", "1%"),
		cfmt.Sprintf("{{%s}}::bold", "10%"),
		cfmt.Sprintf("{{%s}}::bold", "50%"),
		cfmt.Sprintf("{{%s}}::bold", "97%"),
		cfmt.Sprintf("{{%s}}::bold", "99%"),
		cfmt.Sprintf("{{%s}}::bold", "AVG"),
		cfmt.Sprintf("{{%s}}::bold", "MIN"),
		cfmt.Sprintf("{{%s}}::bold", "MAX"),
		cfmt.Sprintf("{{%s}}::bold", "StdDev"))

	table.AddRow(
		cfmt.Sprintf("{{%s}}::green|bold", "Latency"),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.P1)),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.P10)),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.P50)),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.P97)),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.P99)),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.Avg.Milliseconds())),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.Min.Milliseconds())),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.Max.Milliseconds())),
		cfmt.Sprintf("{{%s}}::green|bold", countDuration(m.Time.StdDev.Milliseconds())))

	cfmt.Println(table)
	cfmt.Println("\n")
}

func cliPrintReqBytes(s *Sample) {
	table := uitable.New()
	table.MaxColWidth = 50

	table.AddRow(
		cfmt.Sprintf("{{%s}}::bold", "STAT"),
		cfmt.Sprintf("{{%s}}::bold", "1%"),
		cfmt.Sprintf("{{%s}}::bold", "10%"),
		cfmt.Sprintf("{{%s}}::bold", "50%"),
		cfmt.Sprintf("{{%s}}::bold", "97%"),
		cfmt.Sprintf("{{%s}}::bold", "99%"),
		cfmt.Sprintf("{{%s}}::bold", "AVG"),
		cfmt.Sprintf("{{%s}}::bold", "MIN"),
		cfmt.Sprintf("{{%s}}::bold", "StdDev"))

	d := []int64{s.MetricReqSec.P1, s.MetricReqSec.P10,
		s.MetricReqSec.P50, s.MetricReqSec.P97, s.MetricReqSec.P99}
	ds := []int64{s.MetricBodySize.P1, s.MetricBodySize.P10,
		s.MetricBodySize.P50, s.MetricBodySize.P97, s.MetricBodySize.P99}

	table.AddRow(cfmt.Sprintf("{{%s}}::green|bold", "Req/Sec"),
		cfmt.Sprintf("{{%s}}::green|bold", countRequest(s.MetricReqSec.P1)),
		cfmt.Sprintf("{{%s}}::green|bold", countRequest(s.MetricReqSec.P10)),
		cfmt.Sprintf("{{%s}}::green|bold", countRequest(s.MetricReqSec.P50)),
		cfmt.Sprintf("{{%s}}::green|bold", countRequest(s.MetricReqSec.P97)),
		cfmt.Sprintf("{{%s}}::green|bold", countRequest(s.MetricReqSec.P99)),
		cfmt.Sprintf("{{%s}}::green|bold", countRequest(avgReq(d))),
		cfmt.Sprintf("{{%s}}::green|bold", countRequest(s.MetricReqSec.P1)),
		cfmt.Sprintf("{{%s}}::green|bold", countRequest(int64(stdDevReq(d)))))

	table.AddRow(cfmt.Sprintf("{{%s}}::green|bold", "Bytes/Sec"),
		cfmt.Sprintf("{{%s}}::green|bold", countReadBytes(s.MetricBodySize.P1)),
		cfmt.Sprintf("{{%s}}::green|bold", countReadBytes(s.MetricBodySize.P10)),
		cfmt.Sprintf("{{%s}}::green|bold", countReadBytes(s.MetricBodySize.P50)),
		cfmt.Sprintf("{{%s}}::green|bold", countReadBytes(s.MetricBodySize.P97)),
		cfmt.Sprintf("{{%s}}::green|bold", countReadBytes(s.MetricBodySize.P99)),
		cfmt.Sprintf("{{%s}}::green|bold", countReadBytes(avgReq(ds))),
		cfmt.Sprintf("{{%s}}::green|bold", countReadBytes(s.MetricBodySize.P1)),
		cfmt.Sprintf("{{%s}}::green|bold", countReadBytes(int64(stdDevReq(ds)))))

	cfmt.Println(table)
	cfmt.Println("\n")
}
