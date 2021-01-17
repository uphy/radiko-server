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
	client     *radiko.Client
	ctx        context.Context
	recordings []Recording
	keywords   *keywords
	lastUpdate *time.Time
}

const (
	DatetimeLayout = "20060102150405"
	TZ             = "Asia/Tokyo"
)

func New(baseDir string) (*Library, error) {
	client, err := radiko.New("")
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	client.AuthorizeToken(ctx)

	location, _ := time.LoadLocation(TZ)

	keywords, err := loadKeywords(filepath.Join(baseDir, "keywords.json"))
	if err != nil {
		return nil, err
	}

	return &Library{baseDir, location, client, ctx, nil, keywords, nil}, nil
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

func (l *Library) refreshClient() error {
	if l.lastUpdate != nil && time.Now().Before((*l.lastUpdate).Add(time.Minute*5)) {
		return nil
	}

	client, err := radiko.New("")
	if err != nil {
		return err
	}
	ctx := context.Background()
	client.AuthorizeToken(ctx)
	l.client = client
	t := time.Now()
	l.lastUpdate = &t
	return nil
}

// Record records radiko's program
func (l *Library) Record(stationID string, start time.Time) error {
	dir := l.recordingDirectory(stationID, start)
	dir.create()

	if dir.ready() {
		log.Infof("Already recorded: stationID=%s, start=%s", stationID, start)
		return nil
	}

	// Get program
	if err := l.refreshClient(); err != nil {
		return err
	}
	pg, err := l.client.GetProgramByStartTime(l.ctx, stationID, start)
	if err != nil {
		return fmt.Errorf("Failed to get program: stationID=%s, start=%s, cause=%w", stationID, start, err)
	}
	end, _ := l.ParseTime(pg.To)

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

	dir.saveDetail(&detail)
	dir.saveStatus(&Status{
		Status:           StatusDownloading,
		DownloadProgress: 0,
	})

	// reload library
	go l.Load()

	// Get M3U8 playlist
	uri, err := l.client.TimeshiftPlaylistM3U8(l.ctx, stationID, start)
	if err != nil {
		return fmt.Errorf("Failed to get m3u8 playlist url.  The radio program may not be ready for timeshift play: %w", err)
	}

	// Download audio files
	chunklist, err := radiko.GetChunklistFromM3U8(uri)
	if err != nil {
		return fmt.Errorf("Failed to get m3u8: %w", err)
	}

	// Download
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
		return err
	}
	dir.updateStatus(&Status{
		Status:           StatusConverting,
		DownloadProgress: 1,
	}, true)
	// Concat aac files
	concatedFile, err := ConcatAACFilesFromList(l.ctx, dir.filesDir())
	if err != nil {
		dir.updateStatus(&Status{
			Status:           StatusError,
			Error:            fmt.Sprintf("Failed to concat aac files: %v", err),
			DownloadProgress: 1,
		}, true)
		return err
	}
	os.Remove(dir.aacFile())
	if err := os.Rename(concatedFile, dir.aacFile()); err != nil {
		os.Remove(concatedFile)
		dir.updateStatus(&Status{
			Status:           StatusError,
			Error:            fmt.Sprintf("Failed to rename aac: %v", err),
			DownloadProgress: 1,
		}, true)
		return err
	}
	// Convert aac file
	os.Remove(dir.mp3File())
	if err := ConvertAACtoMP3(l.ctx, dir.aacFile(), dir.mp3File()); err != nil {
		os.Remove(dir.mp3File())
		dir.updateStatus(&Status{
			Status:           StatusError,
			Error:            fmt.Sprintf("Failed to convert aac to mp3: %v", err),
			DownloadProgress: 1,
		}, true)
		return err
	}
	// Finished
	dir.updateStatus(&Status{
		Status:           StatusReady,
		DownloadProgress: 1,
	}, true)
	return nil
}

func (l *Library) Migrate() error {
	recordings, err := l.List()
	if err != nil {
		return err
	}
	for _, recording := range recordings {
		log.Infof("Migrating recording directory: recording=%v", recording)
		if err := l.generateAACMP3(recording.StationID, recording.Start); err != nil {
			log.Errorf("Failed to generate AAC/MP3 file: %s", err)
		}
	}
	return nil
}

// for migration
func (l *Library) generateAACMP3(stationID string, start time.Time) error {
	dir := l.recordingDirectory(stationID, start)
	aacFile := dir.aacFile()
	mp3File := dir.mp3File()
	if _, err := os.Stat(aacFile); os.IsNotExist(err) {
		concatedFile, err := ConcatAACFilesFromList(l.ctx, dir.filesDir())
		if err != nil {
			return err
		}
		if err := os.Rename(concatedFile, aacFile); err != nil {
			os.Remove(concatedFile)
			return err
		}
	}
	if _, err := os.Stat(mp3File); os.IsNotExist(err) {
		// Convert aac file
		if err := ConvertAACtoMP3(l.ctx, aacFile, mp3File); err != nil {
			os.Remove(mp3File)
			return err
		}
	}
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

func (l *Library) MP3(stationID string, start time.Time) string {
	return l.recordingDirectory(stationID, start).mp3File()
}

func (l *Library) AAC(stationID string, start time.Time) string {
	return l.recordingDirectory(stationID, start).aacFile()
}

func (l *Library) RegisterKeyword(keyword string) error {
	if len(keyword) <= 2 {
		return fmt.Errorf("keyword too short: %s", keyword)
	}
	return l.keywords.add(keyword)
}

func (l *Library) UnregisterKeyword(keyword string) error {
	return l.keywords.remove(keyword)
}

func (l *Library) Keywords() ([]string, error) {
	return l.keywords.keywordsSlice(), nil
}

func (l *Library) ScanAndRecord() error {
	keywords := l.keywords.keywordsSlice()
	if len(keywords) == 0 {
		return nil
	}

	currentTime := time.Now().Add(-time.Hour)
	if err := l.refreshClient(); err != nil {
		return err
	}
	stations, err := l.client.GetStations(l.ctx, time.Now())
	if err != nil {
		return fmt.Errorf("Failed to get stations: %w", err)
	}

	for _, station := range stations {
		stationID := station.ID
		log.Infof("Getting weekly programs: stationID=%s", stationID)
		programs, err := l.client.GetWeeklyPrograms(l.ctx, stationID)
		if err != nil {
			return fmt.Errorf("Failed to get weekly programs: stationID=%s, err=%w", stationID, err)
		}
		for _, program := range programs {
			for _, prog := range program.Progs.Progs {
				// Check if the program has finished
				programEnd, err := l.ParseTime(prog.To)
				if err != nil {
					return fmt.Errorf("Failed to parse program end time: prog=%v, err=%w", prog, err)
				}
				if programEnd.After(currentTime) {
					continue
				}

				// Check if the program title match with the keywords
				found := false
				for _, keyword := range keywords {
					if strings.Contains(prog.Title, keyword) {
						found = true
						break
					}
				}
				if !found {
					continue
				}

				// Download
				log.Infof("Downloading program: stationID=%s, start=%s, title=%s", stationID, prog.Ft, prog.Title)
				t, err := l.ParseTime(prog.Ft)
				if err != nil {
					return fmt.Errorf("Failed to parse program start time: prog=%v, err=%w", prog, err)
				}
				if err := l.Record(stationID, t); err != nil {
					return fmt.Errorf("Failed to record: stationID=%s, start=%s, err=%w", stationID, prog.Ft, err)
				}
			}
		}
	}

	return nil
}
