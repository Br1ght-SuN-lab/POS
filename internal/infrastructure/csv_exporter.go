package infrastructure

import (
	"encoding/csv"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"pos/internal/application"
)

type CSVExporter struct{}

func (CSVExporter) Export(path string, report application.Report) error {
	if !strings.EqualFold(filepath.Ext(path), ".csv") {
		return fmt.Errorf("укажите имя файла с расширением .csv")
	}
	return writeFile(path, func(w io.Writer) error {
		writer := csv.NewWriter(w)
		if err := writer.Write([]string{"student_id", "name", "mode", "percent", "graded", "criteria"}); err != nil {
			return err
		}
		for _, row := range report.Rows {
			err := writer.Write([]string{
				row.Student.ID, row.Student.Name, report.Mode,
				strconv.FormatFloat(row.Result.Percent, 'f', 2, 64),
				strconv.Itoa(row.Result.Graded), strconv.Itoa(row.Result.Criteria),
			})
			if err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	})
}
