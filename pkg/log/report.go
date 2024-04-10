package log

import (
	"encoding/json"
	glog "log"
	"sync"
	"time"
)

var reportLogger *glog.Logger

func init() {
	w := Output("./report/data.json", &OutputOptions{Rotate: "day", MaxBackups: 30})
	reportLogger = glog.New(w, "", 0)
}

// Report is used for data reporting.
type Report struct {
	mu     *sync.Mutex
	data   map[string]any
	start  time.Time
	action string
}

// NewReport returns a new instance.
func NewReport(action string) *Report {
	r := &Report{
		mu:     new(sync.Mutex),
		data:   make(map[string]any, 16),
		start:  time.Now(),
		action: action,
	}
	return r
}

// Set key and value.
func (r *Report) Set(k string, v any) *Report {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[k] = v
	return r
}

// Sets key value pairs from map.
func (r *Report) Sets(kvs map[string]any) *Report {
	r.mu.Lock()
	defer r.mu.Unlock()

	for k, v := range kvs {
		r.data[k] = v
	}
	return r
}

// Done writes the report data to file.
func (r *Report) Done() {
	if reportLogger == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.data["action"] = r.action
	r.data["timestamp"] = r.start.UnixMilli()
	r.data["latency_ms"] = time.Since(r.start).Milliseconds()

	bs, err := json.Marshal(r.data)
	if err != nil {
		log.Errorf("marshal report data: %+v, err: %v", r.data, err)
		return
	}

	reportLogger.Printf("%s", bs)
}
