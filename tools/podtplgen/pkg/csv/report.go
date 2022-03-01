package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

type Reporter interface {
	Add(fileName, hash string, options []string) error
	Flush()
	Close() error
}

// Return a new Report instance
func NewReport(fileName string, debug bool) (Reporter, error) {
	csvfile, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	csvWriter := csv.NewWriter(csvfile)

	impl := &reporterImpl{
		file:   csvfile,
		writer: csvWriter,
		debug:  debug,
	}
	// write header
	impl.writer.Write([]string{"manifest", "md5hash", "chart-options"})

	return impl, nil
}

type reporterImpl struct {
	file   *os.File
	writer *csv.Writer

	debug bool
}

// Add use to add a new entry in the CSV report.
func (i *reporterImpl) Add(fileName, hash string, options []string) error {
	var err error
	for id := range options {
		if strings.ContainsRune(options[id], '\n') {
			options[id] = strings.Split(options[id], "=")[0] + "={}"
		}
	}
	csvLine := []string{fileName, hash, strings.Join(options, "|")}
	if err = i.writer.Write(csvLine); err != nil {
		return fmt.Errorf("unable to write in csv file, err: %w", err)
	}
	if i.debug {
		fmt.Println(csvLine)
	}

	return err
}

// Flush writes any buffered data to the underlying io.File.
func (i *reporterImpl) Flush() {
	i.writer.Flush()
}

// Close closes the File, rendering it unusable for I/O.
// On files that support SetDeadline, any pending I/O operations will
// be canceled and return immediately with an error.
// Close will return an error if it has already been called.
func (i *reporterImpl) Close() error {
	var err error
	if i.file != nil {
		err = i.file.Close()
		i.file = nil
	}
	return err
}
