package queryexport

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tealeg/xlsx"
	"os"
	"path/filepath"
	"strings"
)

type QEConf struct {
	User        string
	Pass        string
	Host        string
	Port        string
	DbName      string
	DbType      string
	DbQuery     string
	DbQueryArgs []interface{}
}

// TODO: Additional database support: Postgress, Oracle, etc
func QueryExport(qec QEConf, creFile string) error {
	var resource string
	var err error

	if qec.DbType == "" || qec.DbQuery == "" || creFile == "" {
		return errors.New("Database Type/Query and file argument cannot be empty")
	}

	if qec.DbType == "mysql" {
		if qec.User == "" || qec.Host == "" || qec.Port == "" || qec.DbName == "" {
			return errors.New("User/Host/Port/DbName cannot be empty")
		}
		resource = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", qec.User, qec.Pass, qec.Host, qec.Port, qec.DbName)
	}

	if qec.DbType == "sqlite3" {
		if qec.DbName == "" {
			return errors.New("DbName cannot be empty")
		}
		resource = fmt.Sprintf("%s", qec.DbName)
	}

	db, err := sql.Open(qec.DbType, resource)
	if err != nil {
		return err
	}
	defer db.Close()

	var rows *sql.Rows

	if len(qec.DbQueryArgs) > 0 {
		rows, err = db.Query(qec.DbQuery, qec.DbQueryArgs...)
	} else {
		rows, err = db.Query(qec.DbQuery)
	}
	if err != nil {
		return err
	}

	err = RowsToReport(rows, creFile)
	if err != nil {
		return err
	}

	return nil
}

// thanks to https://yuks.me/blog/go-lang-database-results-to-map-without-structure
// for indexed map[imt] idea, so we can iterate through hasmap query rows and produces
// results as received from query.
func MapRows(rows *sql.Rows) (int, []string, map[int]map[string]interface{}, error) {
	columns, _ := rows.Columns()
	amtCols := len(columns)
	colValKey := make([]interface{}, amtCols)
	colValKeyPtr := make([]interface{}, amtCols)

	indxMapRes := map[int]map[string]interface{}{}
	result_id := 0
	for rows.Next() {
		for i, _ := range columns {
			colValKeyPtr[i] = &colValKey[i]
		}
		err := rows.Scan(colValKeyPtr...)
		if err != nil {
			var s []string
			return 0, s, indxMapRes, err
		}

		buildRes := map[string]interface{}{}

		for i, col := range columns {
			buildRes[col] = colValKey[i]
			b, ok := colValKey[i].([]byte)
			if ok {
				buildRes[col] = string(b)
			}
		}

		indxMapRes[result_id] = buildRes
		result_id++
	}
	defer rows.Close()

	return amtCols, columns, indxMapRes, nil
}

/////////////////////////////////////////////////////////////////////////////
// Case 1.
// transform mysql type database table field name to a common xlsx header type name
// e.g. remove '_', change name to Title Case. i.e.  "create_date" => "CreateDate"
// Other cases todo
/////////////////////////////////////////////////////////////////////////////
func ToReportHeaderCase(fieldName string) string {
	if fieldName == "" {
		return fieldName
	}

	// case 1 - check for underscore return fielName (to Title) if not found
	if strings.Index(fieldName, "_") == -1 {
		return strings.Title(fieldName)
	}

	var reportCase bytes.Buffer
	for _, word := range strings.Split(fieldName, "_") {
		reportCase.WriteString(strings.Title(word))
	}
	return reportCase.String()
}

func mkdirIfNotExist(fileAbsPath string) error {
	fileDir := filepath.Dir(fileAbsPath)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		err := os.MkdirAll(fileDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func RowsToReport(rows *sql.Rows, writeFile string) error {
	//fmt.Println("Writing Earnings to XLSX")
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		return err
	}

	// process db query rows
	_, headers, results, err := MapRows(rows)

	// write headers
	row = sheet.AddRow()
	for i, _ := range headers {
		cell = row.AddCell()
		cell.Value = ToReportHeaderCase(headers[i])
	}

	for i := 0; i < len(results); i++ {
		row = sheet.AddRow()
		for _, h := range headers {
			cell = row.AddCell()
			switch v := results[i][h].(type) {
			case string:
				cell.Value = v
			case int64:
				//fmt.Println("is int64")
				cell.Value = fmt.Sprintf("%d", v)
			case float64:
				//fmt.Println("is float64")
				cell.Value = fmt.Sprintf("%f", v)
			case nil:
				//fmt.Println("is nil")
			default:
				//fmt.Println("default")
			}
		}
	}

	err = mkdirIfNotExist(writeFile)
	if err != nil {
		return fmt.Errorf("Error creating directory path for XLSX file: %v", err)
	}

	err = file.Save(writeFile)
	if err != nil {
		return fmt.Errorf("Error writing XLSX file: %v", err)
	}

	return nil
}
