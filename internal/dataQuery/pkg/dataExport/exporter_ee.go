//go:build enterprise

package data_export

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"log"
	"os"
	"sync"
)

// MySQL extractor
type Extractor interface {
	Start() error
	Columns() []string
	Raws() ([]*Raw, bool, error)
	Close() error
}

// data masking
type Transfer interface {
	transRow([]*Raw) ([]*Raw, error)
}

/*
input: header, rows
output: io.reader
*/
// Exporter
type Exporter interface {
	io.ReadCloser
	WriteHeader([]string) error
	WriteRaws([]*Raw) error
}

// csv exporter
type CsvExport struct {
	reader *io.PipeReader
	writer *io.PipeWriter
	cw     *csv.Writer
}

func NewCsvExport() Exporter {
	reader, writer := io.Pipe()
	return &CsvExport{
		reader: reader,
		writer: writer,
		cw:     csv.NewWriter(writer),
	}
}

func (ce *CsvExport) WriteHeader(headers []string) error {
	return ce.cw.Write(headers)
}

func (ce *CsvExport) WriteRaws(raws []*Raw) error {
	row := make([]string, len(raws))
	for i, raw := range raws {
		row[i] = raw.Value
	}

	err := ce.cw.Write(row)
	if err != nil {
		return err
	}
	ce.cw.Flush()
	return nil
}

func (ce *CsvExport) Read(data []byte) (n int, err error) {
	return ce.reader.Read(data)
}

func (ce *CsvExport) Close() error {
	return ce.writer.Close()
}

type ExportTask struct {
	extract        Extractor
	exportFileName string
	export         Exporter
	trans          Transfer
}

func NewExportTask() *ExportTask {
	return &ExportTask{}
}

func (et *ExportTask) WithExtract(extra Extractor) *ExportTask {
	et.extract = extra
	return et
}

func (et *ExportTask) WithExporter(fileName string, export Exporter) *ExportTask {
	et.exportFileName = fileName
	et.export = export
	return et
}

func (et *ExportTask) WithTransfer(trans Transfer) *ExportTask {
	et.trans = trans
	return et
}

func (et *ExportTask) Start() error {
	defer func() {
		et.extract.Close()
		et.export.Close()
	}()
	err := et.extract.Start()
	if err != nil {
		return err
	}

	err = et.export.WriteHeader(et.extract.Columns())
	if err != nil {
		return err
	}
	for {
		raws, exist, err := et.extract.Raws()
		if err != nil {
			return err
		}
		if !exist {
			break
		}
		if et.trans != nil {
			raws, err = et.trans.transRow(raws)
			if err != nil {
				return err
			}
		}
		err = et.export.WriteRaws(raws)
		if err != nil {
			return err
		}
	}
	return nil
}

func (et *ExportTask) Output() io.Reader {
	return et.export
}

func ExportTasksToZip(fileName string, tasks []*ExportTask) error {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	for _, task := range tasks {
		err := exportTasksToZip(w, task)
		if err != nil {
			return err
		}
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func exportTasksToZip(w *zip.Writer, task *ExportTask) error {
	wait := sync.WaitGroup{}
	wait.Add(1)
	go func() {
		defer wait.Done()
		f, err := w.Create(task.exportFileName)
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(f, task.Output())
		if err != nil {
			log.Fatal(err)
		}
	}()
	err := task.Start()
	wait.Wait()
	return err
}
