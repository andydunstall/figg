package bench

import (
	"fmt"
	"math"
	"time"

	humanize "github.com/dustin/go-humanize"
)

type sample struct {
	Messages     int
	MessageBytes int
	Start        time.Time
	End          time.Time
}

func newSample(messages int, messageSize int, start time.Time, end time.Time) *sample {
	return &sample{
		Messages:     messages,
		MessageBytes: messages * messageSize,
		Start:        start,
		End:          end,
	}
}

// Throughput of bytes per second
func (s *sample) Throughput() float64 {
	return float64(s.MessageBytes) / s.Duration().Seconds()
}

// Rate of meessages in the job per second
func (s *sample) Rate() int64 {
	return int64(float64(s.Messages) / s.Duration().Seconds())
}

func (s *sample) String() string {
	rate := humanize.Comma(s.Rate())
	throughput := HumanBytes(s.Throughput(), false)
	return fmt.Sprintf("%s msgs/sec ~ %s/sec", rate, throughput)
}

// Duration that the sample was active
func (s *sample) Duration() time.Duration {
	return s.End.Sub(s.Start)
}

type sampleGroup struct {
	sample
	Samples []*sample
}

func newSampleGroup() *sampleGroup {
	return &sampleGroup{
		Samples: []*sample{},
	}
}

func (sg *sampleGroup) AddSample(s *sample) {
	sg.Samples = append(sg.Samples, s)

	if len(sg.Samples) == 1 {
		sg.Start = s.Start
		sg.End = s.End
	}

	sg.Messages += s.Messages
	sg.MessageBytes += s.MessageBytes

	if s.Start.Before(sg.Start) {
		sg.Start = s.Start
	}

	if s.End.After(sg.End) {
		sg.End = s.End
	}
}

// HumanBytes formats bytes as a human readable string
func HumanBytes(bytes float64, si bool) string {
	var base = 1024
	pre := []string{"K", "M", "G", "T", "P", "E"}
	var post = "B"
	if si {
		base = 1000
		pre = []string{"k", "M", "G", "T", "P", "E"}
		post = "iB"
	}
	if bytes < float64(base) {
		return fmt.Sprintf("%.2f B", bytes)
	}
	exp := int(math.Log(bytes) / math.Log(float64(base)))
	index := exp - 1
	units := pre[index] + post
	return fmt.Sprintf("%.2f %s", bytes/math.Pow(float64(base), float64(exp)), units)
}
