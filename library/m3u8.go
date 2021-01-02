package library

import (
	"io"
	"text/template"
	"time"
)

const tmpl = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-ALLOW-CACHE:NO
#EXT-X-TARGETDURATION:5
#EXT-X-MEDIA-SEQUENCE:1
{{- range $file := .files }}
{{- $time := ($file | parseTime )}}
#EXT-X-PROGRAM-DATE-TIME:{{ $time | formatTime }}
#EXTINF:5,
{{ $.baseURL }}{{ $file -}}
{{ end }}

#EXT-X-ENDLIST
`

type audioFileName string

func (l *recordingDirectory) generateM3U8(baseURL string, files []string, w io.Writer) error {
	t := template.Must(template.New("m3u8-template").
		Funcs(template.FuncMap{
			"parseTime": func(filename string) (time.Time, error) {
				// 20201231_230000_1hVP0.aac
				timePart := filename[0:15]
				return time.Parse("20060102_150405", timePart)
			},
			"formatTime": func(t time.Time) string {
				// 2020-12-31T23:00:05+09:00
				return t.Format("2006-01-12T15:04:05+09:00")
			},
		}).
		Parse(tmpl))

	return t.Execute(w, map[string]interface{}{
		"baseURL": baseURL,
		"files":   files,
	})
}
