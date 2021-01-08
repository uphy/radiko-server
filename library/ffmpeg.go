package library

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type ffmpeg struct {
	*exec.Cmd
}

func newFfmpeg(ctx context.Context) (*ffmpeg, error) {
	cmdPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, err
	}

	return &ffmpeg{exec.CommandContext(
		ctx,
		cmdPath,
	)}, nil
}

func (f *ffmpeg) setDir(dir string) {
	f.Dir = dir
}

func (f *ffmpeg) setArgs(args ...string) {
	f.Args = append(f.Args, args...)
}

func (f *ffmpeg) setInput(input string) {
	f.setArgs("-i", input)
}

func (f *ffmpeg) run(output string) error {
	f.setArgs(output)
	log.Printf("Execute ffmpeg: args=%v", f.Args)
	return f.Run()
}

func (f *ffmpeg) start(output string) error {
	f.setArgs(output)
	return f.Start()
}

func (f *ffmpeg) wait() error {
	return f.Wait()
}

func (f *ffmpeg) stdinPipe() (io.WriteCloser, error) {
	return f.StdinPipe()
}

func (f *ffmpeg) stderrPipe() (io.ReadCloser, error) {
	return f.StderrPipe()
}

// ConvertAACtoMP3 converts an aac file to a mp3 file.
func ConvertAACtoMP3(ctx context.Context, input, output string) error {
	f, err := newFfmpeg(ctx)
	if err != nil {
		return err
	}

	f.setInput(input)
	f.setArgs(
		"-c:a", "libmp3lame",
		"-ac", "2",
		"-q:a", "2",
		"-y", // overwrite the output file without asking
	)
	// TODO: Collect log
	return f.run(output)
}

// ConcatAACFilesFromList concatenates files from the list of resources.
func ConcatAACFilesFromList(ctx context.Context, resourcesDir string) (string, error) {
	files, err := ioutil.ReadDir(resourcesDir)
	if err != nil {
		return "", err
	}

	allFilePaths := []string{}
	for _, f := range files {
		p := filepath.Join(resourcesDir, f.Name())
		allFilePaths = append(allFilePaths, p)
	}
	concatedFile := filepath.Join(resourcesDir, "concated.aac")
	if err := ConcatAACFilesAll(ctx, allFilePaths, resourcesDir, concatedFile); err != nil {
		return "", err
	}

	return concatedFile, nil
}

// ConcatAACFiles concatenate files of the same type.
func ConcatAACFilesAll(ctx context.Context, files []string, resourcesDir string, output string) error {
	// input is a path to a file which lists all the aac files.
	// it may include a lot of aac file and exceed max number of file descriptor.
	oneConcatNum := 100
	if len(files) > oneConcatNum {
		reducedFiles := files[:oneConcatNum]
		restFiles := files[oneConcatNum:]
		// reducedFiles -> reducedFiles[0]
		tmpOutputFile, err := ioutil.TempFile(resourcesDir, "tmp-concatenated-*.aac")
		if err != nil {
			fmt.Println("Failed to call ioutil.TempFile")
			return err
		}
		defer os.Remove(tmpOutputFile.Name())
		err = ConcatAACFiles(ctx, reducedFiles, resourcesDir, tmpOutputFile.Name())
		if err != nil {
			fmt.Printf("Failed to ConcatAACFiles: %v\n", err)
			return err
		}
		return ConcatAACFilesAll(ctx, append([]string{tmpOutputFile.Name()}, restFiles...), resourcesDir, output)
	} else {
		return ConcatAACFiles(ctx, files, resourcesDir, output)
	}
}

func ConcatAACFiles(ctx context.Context, input []string, resourcesDir string, output string) error {
	listFile, err := generateFileList(resourcesDir, input)
	if err != nil {
		return err
	}
	defer os.Remove(listFile)

	f, err := newFfmpeg(ctx)
	if err != nil {
		return err
	}

	f.setArgs(
		"-f", "concat",
		"-safe", "0",
		"-y",
	)
	f.setInput(listFile)
	f.setArgs("-c", "copy")
	// TODO: Collect log
	err = f.run(output)
	return err
}

func generateFileList(resourcesDir string, input []string) (string, error) {
	listFile, err := ioutil.TempFile(resourcesDir, "aac_resources")
	if err != nil {
		return "", err
	}
	defer listFile.Close()

	for _, f := range input {
		abs, err := filepath.Abs(f)
		if err != nil {
			return "", err
		}
		p := fmt.Sprintf("file '%s'\n", abs)
		if _, err := listFile.WriteString(p); err != nil {
			return "", err
		}
	}
	return listFile.Name(), nil
}
