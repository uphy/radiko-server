package library

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	StatusDownloading = "DOWNLOADING"
	StatusReady       = "READY"
	StatusError       = "FAILED"
)

type (
	Recording struct {
		Title     string    `json:"title"`
		StationID string    `json:"stationId"`
		Start     time.Time `json:"start"`
		End       time.Time `json:"end"`
	}
	RecordingDetail struct {
		Recording
		Description string `json:"description"`
		Subtitle    string `json:"subtitle"`
		URL         string `json:"url"`
		Info        string `json:"info"`
	}

	Status struct {
		Status           string  `json:"status"`
		DownloadProgress float32 `json:"downloadProgress"`
		Error            string  `json:"error,omitempty"`
	}
	recordingDirectory struct {
		dir              string
		nextStatusUpdate *time.Time
	}
)

func (l *recordingDirectory) exists() bool {
	if _, err := os.Stat(l.dir); os.IsNotExist(err) {
		return false
	}
	return true
}

func (l *recordingDirectory) create() error {
	if err := os.MkdirAll(l.dir, 0777); err != nil {
		return err
	}
	return os.Mkdir(l.filesDir(), 0777)
}

func (l *recordingDirectory) loadDetail() (*RecordingDetail, error) {
	var recordingDetail RecordingDetail
	if err := l.loadJSON(filepath.Join(l.dir, "info.json"), &recordingDetail); err != nil {
		return nil, err
	}
	return &recordingDetail, nil
}

func (l *recordingDirectory) saveDetail(detail *RecordingDetail) error {
	return l.saveJSON(filepath.Join(l.dir, "info.json"), detail)
}

func (l *recordingDirectory) loadStatus() (*Status, error) {
	var status Status
	if err := l.loadJSON(filepath.Join(l.dir, "status.json"), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (l *recordingDirectory) updateStatus(status *Status, force bool) error {
	t := time.Now()
	if !force && l.nextStatusUpdate != nil && l.nextStatusUpdate.After(t) {
		return nil
	}
	err := l.saveStatus(status)
	next := t.Add(time.Second)
	l.nextStatusUpdate = &next
	return err
}

func (l *recordingDirectory) saveStatus(status *Status) error {
	return l.saveJSON(filepath.Join(l.dir, "status.json"), status)
}

func (l *recordingDirectory) saveJSON(file string, v interface{}) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(v)
}

func (l *recordingDirectory) loadJSON(file string, v interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	return json.NewDecoder(f).Decode(&v)
}

func (l *recordingDirectory) filesDir() string {
	return filepath.Join(l.dir, "files")
}
