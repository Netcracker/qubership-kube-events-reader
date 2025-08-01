package filter

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/yaml"
	"log/slog"
	"os"
)

type Filters struct {
	Sinks []*Sink `json:"sinks"`
}

type Sink struct {
	Name    string       `json:"name"`
	Match   []EventMatch `json:"match,omitempty"`
	Exclude []EventMatch `json:"exclude,omitempty"`
}

type EventMatch struct {
	Type                string `json:"type"`
	Kind                string `json:"kind"`
	Reason              string `json:"reason"`
	Namespace           string `json:"namespace"`
	ReportingController string `json:"reportingController"`
	ReportingInstance   string `json:"reportingInstance"`
	Message             string `json:"message"`
}

func ParseFiltersConfiguration(configPath string) (*Filters, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Warn("File with configuration of filtering events does not exist", "path", configPath)
			return &Filters{}, nil
		}
		return nil, err
	}
	var filters Filters
	if err = yaml.Unmarshal(content, &filters); err != nil {
		return nil, err
	}
	return &filters, nil
}

func (f *Filters) GetSinkFiltersByName(sink string) *Sink {
	for _, s := range f.Sinks {
		if s.Name == sink {
			return s
		}
	}
	return nil
}

func ValidateFileSize(filePath string, maxSize int64) (err error) {
	// Open the file and get its size.

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		cle := file.Close()
		if cle != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", cle)
		}
	}()

	stats, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	if stats.Size() > maxSize {
		return fmt.Errorf("file exceeds maximum allowed size of %d bytes", maxSize)
	}
	return nil
}
