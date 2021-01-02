package library

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/yyoshiki41/go-radiko"
)

type Library struct {
	baseDir    string
	location   *time.Location
	recordings []Recording
}

const (
	DatetimeLayout = "20060102150405"
	TZ             = "Asia/Tokyo"
)

func New(baseDir string) *Library {
	location, _ := time.LoadLocation(TZ)
	return &Library{baseDir, location, nil}
}

// Load loads local library file system.
func (l *Library) Load() error {
	l.recordings = make([]Recording, 0)
	filepath.Walk(l.baseDir, func(path string, info os.FileInfo, err error) error {
		dir, name := filepath.Split(path)
		if name == "status.json" {
			d := l.recordingDirectoryFromDir(dir)
			detail, err := d.loadDetail()
			if err != nil {
				return err
			}
			l.recordings = append(l.recordings, detail.Recording)
		}
		return nil
	})
	return nil
}

func (l *Library) Get(stationID string, start time.Time) (*RecordingDetail, error) {
	dir := l.recordingDirectory(stationID, start)
	return dir.loadDetail()
}

func (l *Library) GetStatus(stationID string, start time.Time) (*Status, error) {
	dir := l.recordingDirectory(stationID, start)
	return dir.loadStatus()
}

// List lists all recordings
func (l *Library) List() ([]Recording, error) {
	return l.recordings, nil
}

// Record records radiko's program
func (l *Library) Record(stationID string, start time.Time) error {
	client, err := radiko.New("")
	if err != nil {
		return err
	}
	ctx := context.Background()
	client.AuthorizeToken(ctx)

	// Get program
	pg, err := client.GetProgramByStartTime(ctx, stationID, start)
	if err != nil {
		return fmt.Errorf("Failed to get program: stationID=%s, start=%s, cause=%w", stationID, start, err)
	}
	end, _ := l.ParseTime(pg.To)

	dir := l.recordingDirectory(stationID, start)
	dir.create()

	detail := RecordingDetail{
		Recording: Recording{
			Title:     pg.Title,
			StationID: stationID,
			Start:     start,
			End:       end,
		},
		Description: pg.Desc,
		Subtitle:    pg.SubTitle,
		URL:         pg.URL,
		Info:        pg.Info,
	}
	// reload library
	go l.Load()

	dir.saveDetail(&detail)
	dir.saveStatus(&Status{
		Status:           StatusDownloading,
		DownloadProgress: 0,
	})

	// Get M3U8 playlist
	uri, err := client.TimeshiftPlaylistM3U8(ctx, stationID, start)
	if err != nil {
		return fmt.Errorf("Failed to get m3u8 playlist url: %w", err)
	}

	// Download audio files
	chunklist, err := radiko.GetChunklistFromM3U8(uri)
	if err != nil {
		return fmt.Errorf("Failed to get m3u8: %w", err)
	}

	// Download
	go func() {
		if err := bulkDownload(chunklist, dir.filesDir(), func(progress float32) {
			dir.updateStatus(&Status{
				Status:           StatusDownloading,
				DownloadProgress: progress,
			}, false)
		}); err != nil {
			// Failed to download
			os.RemoveAll(dir.filesDir())
			dir.updateStatus(&Status{
				Status:           StatusError,
				Error:            fmt.Sprintf("Failed to download audio files: %v", err),
				DownloadProgress: 0,
			}, true)
			log.Errorf("Failed to download audio files: %w", err)
		} else {
			// Successfully downloaded
			dir.updateStatus(&Status{
				Status:           StatusReady,
				DownloadProgress: 1,
			}, true)
		}
	}()
	return nil
}

func (l *Library) recordingDirectory(stationID string, start time.Time) *recordingDirectory {
	return l.recordingDirectoryFromDir(filepath.Join(l.baseDir, stationID, start.Format(DatetimeLayout)))
}

func (l *Library) recordingDirectoryFromDir(dir string) *recordingDirectory {
	return &recordingDirectory{dir, nil}
}

func (l *Library) ParseTime(start string) (time.Time, error) {
	return time.ParseInLocation(DatetimeLayout, start, l.location)
}

func (l *Library) FormatTime(start time.Time) string {
	return start.Format(DatetimeLayout)
}

func (l *Library) File(stationID string, start time.Time, filename string) string {
	return filepath.Join(l.recordingDirectory(stationID, start).filesDir(), filename)
}

func (l *Library) GenerateM3U8(baseURL string, stationID string, start time.Time, w io.Writer) error {
	dir := l.recordingDirectory(stationID, start)

	filesDir := dir.filesDir()
	rel, err := filepath.Rel(l.baseDir, filesDir)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}
	if !strings.HasSuffix(rel, "/") {
		rel = rel + "/"
	}

	f, _ := os.Open(filesDir)
	defer f.Close()
	filenames, err := f.Readdirnames(-1)
	if err != nil {
		return err
	}
	sort.Slice(filenames, func(i, j int) bool {
		f1 := filenames[i]
		f2 := filenames[j]
		return strings.Compare(f1, f2) < 0
	})
	return dir.generateM3U8(fmt.Sprintf("%srecordings/recording/%s/%s/", baseURL, stationID, l.FormatTime(start)), filenames, w)
}
