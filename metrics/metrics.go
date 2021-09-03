package metrics

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var client Client

// Client holds the info about client
type Client struct {
	Service string
	Host    string
	Port    string
}

// Config contains the configuration for the metrics
type Config struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Service string `json:"service"`
}

type metricDump struct {
	Name      string                 `json:"name"`
	Filters   map[string]string      `json:"filters"`
	Fields    map[string]interface{} `json:"fields"`
	TimeStamp int64                  `json:"timestamp"`
}

// Metric contains the metric to be sent
type Metric struct {
	metricDump
	sync.Mutex
}

// SetName sets the metric name
func (m *Metric) SetName(name string) {
	m.Name = name
}

// GetName returns the metric name
func (m *Metric) GetName() string {
	return m.Name
}

// AddFilter add the filter to the metric
func (m *Metric) AddFilter(key, value string) error {
	if len(key) <= 0 || len(value) <= 0 {
		return errors.New("Key/value can't be empty")
	}
	m.Lock()
	m.Filters[key] = value
	m.Unlock()
	return nil
}

// AddFilters add all the filters to the metric
func (m *Metric) AddFilters(filters map[string]string) error {
	for key, value := range filters {
		if err := m.AddFilter(key, value); err != nil {
			return err
		}
	}
	return nil
}

// RemoveFilter removes the filter from the metric
func (m *Metric) RemoveFilter(key string) {
	m.Lock()
	delete(m.Filters, key)
	m.Unlock()
	return
}

// GetFilters returns all the filters for the metric
func (m *Metric) GetFilters() map[string]string {
	m.Lock()
	defer m.Unlock()
	return m.Filters
}

// AddField adds the field to the metric
func (m *Metric) AddField(key string, value interface{}) error {
	if len(key) <= 0 {
		return errors.New("Field name can't be empty")
	}
	m.Lock()
	m.Fields[key] = value
	m.Unlock()
	return nil
}

// AddFields adds all the fields to the metric
func (m *Metric) AddFields(fields map[string]interface{}) error {
	for key, value := range fields {
		if err := m.AddField(key, value); err != nil {
			return err
		}
	}
	return nil
}

// RemoveField removes the field from the metric
func (m *Metric) RemoveField(key string) {
	m.Lock()
	delete(m.Fields, key)
	m.Unlock()
	return
}

// GetFields returns all the fields of the metric
func (m *Metric) GetFields() map[string]interface{} {
	m.Lock()
	defer m.Unlock()
	return m.Fields
}

// SetTimeStamp sets the timestamp for the metric
func (m *Metric) SetTimeStamp(t time.Time) {
	m.TimeStamp = t.Unix()
}

// GetTimeStamp returns the timestamp set for the metric
func (m *Metric) GetTimeStamp() int64 {
	return m.TimeStamp
}

// Serialize marshals the metric
func (m *Metric) Serialize() ([]byte, error) {
	data, err := json.Marshal(m.metricDump)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

// String returns string representation of the metric
func (m *Metric) String() (string, error) {
	met, err := m.Serialize()
	if err != nil {
		return "", err
	}
	return string(met), nil
}

// InitClient initializes the client
func InitClient(config Config) error {
	client = Client{}
	if len(config.Host) <= 0 {
		return errors.New("Failed initializing the metric client. Hostname is empty")
	}
	if len(string(config.Port)) <= 0 {
		return errors.New("Failed initializing the metric client. Port is empty")
	}
	if len(config.Service) <= 0 {
		return errors.New("Failed initializing the metric client. Service is empty")
	}
	client.Host = config.Host
	client.Port = strconv.Itoa(config.Port)
	client.Service = config.Service
	return nil
}

// NewMetric is used to create new metric object
func NewMetric(name string, filters map[string]string, fields map[string]interface{}) (*Metric, error) {
	metric := new(Metric)
	metric.Filters = make(map[string]string)
	metric.Fields = make(map[string]interface{})
	metric.SetName(name)
	if err := metric.AddFilters(filters); err != nil {
		return metric, err
	}
	if err := metric.AddFields(fields); err != nil {
		return metric, err
	}
	metric.SetTimeStamp(time.Now())
	return metric, nil
}

// SendMetric sends the metric
func SendMetric(metric *Metric) error {
	hostName, err := os.Hostname()
	if err != nil {
		return errors.New("Failed sending the metric. Hostname not found")
	}
	metric.AddFilter("host", hostName)
	metric.AddFilter("service", client.Service)
	msg, err := metric.Serialize()
	if err != nil {
		return errors.New("Failed sending the metric. Couldn't serialize the metric")
	}
	conn, err := net.Dial("udp", client.Host+":"+client.Port)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(msg)
	if err != nil {
		return err
	}
	return nil
}
