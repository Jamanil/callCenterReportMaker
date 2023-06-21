package csvReader

import (
	"callCenterReportMaker/entity"
	"encoding/csv"
	"os"
	"time"
)

const (
	fieldsPerRecord = 4
	comma           = ';'
	comment         = '#'
	dateLayout      = "02.01.2006"
)

func New() CsvReader {
	return &csvReader{}
}

type CsvReader interface {
	GetHistory(source string) ([]entity.HistoryRecord, error)
}

type csvReader struct {
	reader *csv.Reader
}

func (r *csvReader) GetHistory(source string) ([]entity.HistoryRecord, error) {
	data, err := r.getRawData(source)
	if err != nil {
		return nil, err
	}
	historyRecords := r.convertRawDataToHistoryRecords(data)

	return historyRecords, err
}
func (r *csvReader) getRawData(source string) (data [][]string, err error) {
	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err = file.Close()
	}(file)

	r.reader = csv.NewReader(file)
	r.reader.FieldsPerRecord = fieldsPerRecord
	r.reader.Comma = comma
	r.reader.Comment = comment

	return r.reader.ReadAll()
}
func (r *csvReader) convertRawDataToHistoryRecords(data [][]string) []entity.HistoryRecord {
	historyRecords := make([]entity.HistoryRecord, 0, len(data)-1)

	for i, datum := range data {
		if i == 0 {
			continue
		}
		currentRecord := entity.HistoryRecord{
			Abonent:    datum[1],
			Operator:   datum[2],
			LineNumber: datum[3],
		}
		currentRecord.Date, _ = time.ParseInLocation(dateLayout, datum[0], time.Local)
		historyRecords = append(historyRecords, currentRecord)
	}
	return historyRecords
}
