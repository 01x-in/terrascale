package ui

import (
	"fmt"
	"io"

	"github.com/rodaine/table"
)

// PrintTable renders a table with the given headers and rows.
func PrintTable(w io.Writer, headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Convert headers to []interface{}
	iHeaders := make([]interface{}, len(headers))
	for i, h := range headers {
		iHeaders[i] = h
	}

	tbl := table.New(iHeaders...)
	tbl.WithWriter(w)

	for _, row := range rows {
		iRow := make([]interface{}, len(row))
		for i, cell := range row {
			iRow[i] = cell
		}
		tbl.AddRow(iRow...)
	}

	tbl.Print()
	fmt.Fprintln(w)
}
