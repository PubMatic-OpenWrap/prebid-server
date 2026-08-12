package openrtb_ext

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"golang.org/x/exp/rand"
)

var (
	pid            = os.Getpid()
	wsRemoverRegex = regexp.MustCompile(`>\s+<`)
	XMLLogFormat   = "\n[XML_PARSER] parser:[%s] method:[%s] response:[%s]"
)

const (
	maxFileSize   = 1 * 1024 * 1024 * 1024
	maxBufferSize = 256 * 1024
	maxFiles      = 10
	flushInterval = time.Second * time.Duration(300)
)

const (
	XMLParserETree   = "etree"
	XMLParserFastXML = "fastxml"
)

// Writer interface can be used to define variable returned by GetWriter() method
type Writer interface {
	Write(data []byte) (int, error)
	Sync() error
}

// FileWriter ...
type FileWriter struct {
	mu       sync.Mutex
	file     Writer
	fileName string
}

// NewFileWriter ...
func NewFileWriter(dirPath, fileName, ext string) Writer {
	//create directory if not exists
	_ = os.MkdirAll(dirPath, 0755)

	writer := &FileWriter{
		mu:       sync.Mutex{},
		file:     NewBufferFileWriter(dirPath, fileName, ext),
		fileName: fileName,
	}

	go purge(dirPath, fileName, ext)
	go writer.flush(flushInterval)

	return writer
}

// Flushd ...
func (f *FileWriter) flush(t time.Duration) {
	defer func() {
		if errInterface := recover(); errInterface != nil {
			glog.Infof("Recovered panic \n Error: %v \n StackTrace: %v", errInterface, string(debug.Stack()))
		}
	}()

	for {
		f.Sync()
		time.Sleep(t)
	}
}

// Sync ...
func (f *FileWriter) Sync() (err error) {
	f.mu.Lock()
	err = f.file.Sync()
	f.mu.Unlock()
	return err
}

// Write ...
func (f *FileWriter) Write(data []byte) (n int, err error) {
	f.mu.Lock()
	n, err = f.file.Write(data)
	f.mu.Unlock()
	return n, err
}

// purge files
func purge(dirPath, fileName, ext string) {
	fileFormat := dirPath + fileName + "*" + ext
	for {
		_purge(fileFormat, maxFiles)
		time.Sleep(flushInterval)
	}
}

func _purge(fileFormat string, maxFiles int) {
	defer func() {
		if errInterface := recover(); errInterface != nil {
			glog.Infof("Recovered panic \n Error: %v \n StackTrace: %v", errInterface, string(debug.Stack()))
		}
	}()

	files, _ := filepath.Glob(fileFormat)
	sort.Strings(files)

	//remove last files
	if len(files) <= maxFiles {
		//no files to purge
		return
	}

	//limit files to max files
	files = files[:len(files)-maxFiles]
	for _, file := range files {
		glog.Infof("[purger] filename:[%s]\n", file)
		if err := os.Remove(file); err != nil {
			glog.Infof("[purger] error:[purge_failed] file:[%s] message:[%s]", file, err.Error())
			//do not delete status file if original file not deleted
			continue
		}
	}
}

// bufferFileWriter ...
type bufferFileWriter struct {
	dirPath, fileName, ext string

	buf    *bufio.Writer
	file   *os.File
	nbytes uint64
}

func NewBufferFileWriter(dirPath, fileName, ext string) *bufferFileWriter {
	writer := &bufferFileWriter{
		dirPath:  dirPath,
		fileName: fileName,
		ext:      ext,
	}
	return writer
}

// Sync ...
func (b *bufferFileWriter) Sync() (err error) {
	if b.buf != nil {
		if err = b.buf.Flush(); err != nil {
			return err
		}
	}
	if b.file != nil {
		if err = b.file.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// Write ...
func (b *bufferFileWriter) Write(data []byte) (int, error) {
	if b.file == nil {
		//create new file
		if err := b.create(time.Now()); err != nil {
			return 0, err
		}
	}

	if b.nbytes+uint64(len(data)) >= maxFileSize {
		//rotate file
		if err := b.create(time.Now()); err != nil {
			return 0, err
		}
	}

	n, err := b.buf.Write(data)
	b.nbytes += uint64(n)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (b *bufferFileWriter) create(t time.Time) (err error) {
	if b.file != nil {
		if err = b.buf.Flush(); err != nil {
			return err
		}

		if err = b.file.Close(); err != nil {
			return err
		}
	}

	fname := filepath.Join(b.dirPath, fileNameFormat(b.fileName, b.ext, t))
	b.file, err = os.Create(fname)
	b.nbytes = 0
	if err != nil {
		return err
	}

	glog.Infof("[file_writer] type:[new_file] filename:[%s]\n", fname)
	b.buf = bufio.NewWriterSize(b.file, int(maxBufferSize))
	return err
}

func fileNameFormat(name, ext string, t time.Time) string {
	return fmt.Sprintf("%s.%04d%02d%02d-%02d%02d%02d.%d%s",
		name,
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		pid,
		ext)
}

type RandomGenerator interface {
	GenerateIntn(int) int
}

type RandomNumberGenerator struct{}

func (RandomNumberGenerator) GenerateIntn(n int) int {
	return rand.Intn(n)
}

func XMLLogf(format string, args ...any) {
	if bfw != nil {
		fmt.Fprintf(bfw, format, args...)
	}
}

func NormalizeXML(xml string) string {
	//replace only if trackers are injected
	xml = strings.TrimSpace(xml)                     //step1: remove heading and trailing whitespaces
	xml = wsRemoverRegex.ReplaceAllString(xml, "><") //step2: remove inbetween whitespaces
	xml = strings.ReplaceAll(xml, " = ", "=")        //step3: remove whitespaces near attribute
	xml = strings.ReplaceAll(xml, " ><", "><")       //step4: remove attribute endtag whitespace (this should be always before step2)
	xml = strings.ReplaceAll(xml, "'", "\"")         //step5: convert single quote to double quote
	return xml
}

func SetFastXMLEnablingPercentage(percentage int) {
	gFastXMLEnablingPercentage = percentage
}

func IsFastXMLEnabled() bool {
	return gFastXMLEnablingPercentage > 0 && gFastXMLEnablingPercentage >= rg.GenerateIntn(gFastXMLEnablingPercentage)
}

func IsXMLComparingModeEnabled() bool {
	return false
}

type XMLMetricsStats struct {
	ParserName  string
	ParsingTime time.Duration
	HasError    bool
}

type XMLMetrics struct {
	ParserName  string
	ParsingTime time.Duration
	HasError    bool
}

var (
	rg                         RandomGenerator
	bfw                        Writer
	gFastXMLEnablingPercentage int
)

func init() {
	rg = &RandomNumberGenerator{}
	bfw = NewFileWriter(`/var/log/ssheaderbidding/`, `fastxml`, `.txt`)
	gFastXMLEnablingPercentage = 0
}
