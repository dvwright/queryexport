package queryexport_test

import (
	"database/sql"
	"fmt"
	"github.com/dvwright/queryexport"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tealeg/xlsx"
	"os"
	"testing"
	// "os"
	// "path/filepath"
	// "strings"
)

var sqliteTestDB = "./test.db"
var testExcelFile = "test.xlsx"

func preSetupDb() {
	db, _ := sql.Open("sqlite3", sqliteTestDB)
	stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY, col1 TEXT, col2 TEXT, col3 INT, col4 INT)")
	stmt.Exec()
	stmt, _ = db.Prepare("INSERT INTO test (col1, col2, col3, col4) VALUES (?, ?, ?, ?), (?, ?, ?, ?), (?, ?, ?, ?)")
	stmt.Exec("test data row1 column1", "test data row1 column2", "3", "4", "test data row2 column1", "test data row2 column2", "3", "4", "test data row3 column1", "test data row3 column2", "3", "4")

}

func postCleanup() {
	_ = os.Remove(sqliteTestDB)
	_ = os.Remove(testExcelFile)
}

func TestQueryExportSqlite(t *testing.T) {
	preSetupDb()

	dbargs := []interface{}{"3", "4"}

	qe := queryexport.QEConf{
		DbName:      sqliteTestDB,
		DbType:      "sqlite3",
		DbQuery:     `SELECT * FROM test WHERE col3 = ? AND col4 = ?`,
		DbQueryArgs: dbargs,
	}

	err := queryexport.QueryExport(qe, testExcelFile)
	if err != nil {
		t.Errorf("Error writing Excel report file: %s", err.Error())
	}
	if _, err := os.Stat(testExcelFile); os.IsNotExist(err) {
		t.Errorf("Excel file not create")
	}
	postCleanup()
}

func TestQueryExportMysql(t *testing.T) {

	dbargs := []interface{}{"100"}

	// https://rfam.readthedocs.io/en/latest/database.html
	// confirm results
	qe := queryexport.QEConf{
		User:        "rfamro",
		Pass:        "",
		Host:        "mysql-rfam-public.ebi.ac.uk",
		Port:        "4497",
		DbName:      "Rfam",
		DbType:      "mysql",
		DbQuery:     `SELECT * FROM taxonomy WHERE ncbi_id < ? LIMIT 100`,
		DbQueryArgs: dbargs,
	}

	file := "rfamro.xlsx"
	err := queryexport.QueryExport(qe, file)
	if err != nil {
		t.Errorf("Error writing Excel report file: %s", err.Error())
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Errorf("Excel file not create")
	}
	_ = os.Remove(file)
}

func TestQueryExportRowsToFile(t *testing.T) {

	user := "rfamro"
	pass := ""
	host := "mysql-rfam-public.ebi.ac.uk"
	port := "4497"
	dbName := "Rfam"

	resource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, dbName)
	db, err := sql.Open("mysql", resource)
	if err != nil {
		t.Errorf("Database connection issues: %s", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT * FROM taxonomy WHERE ncbi_id < 100 LIMIT 100`)
	if err != nil {
		t.Errorf("Database query error: %s", err)
	}

	file := "rfamro.xlsx"
	err = queryexport.RowsToReport(rows, file)
	if err != nil {
		t.Errorf("Error writing Excel report file: %s", err.Error())
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Errorf("Excel file not create")
	}
	_ = os.Remove(file)
}

func TestQueryExportMapRows(t *testing.T) {
	preSetupDb()

	db, err := sql.Open("sqlite3", sqliteTestDB)
	if err != nil {
		t.Errorf("Database connection issues: %s", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT * FROM test`)
	if err != nil {
		t.Errorf("Database query error: %s", err)
	}

	headers_count, headers, results, err := queryexport.MapRows(rows)
	if err != nil {
		t.Errorf("Error mapping db rows: %s", err.Error())
	}

	if 5 != headers_count {
		t.Errorf("Unexpected header count")
	}

	if 3 != len(results) {
		t.Errorf("Unexpected total results")
	}

	var total_str_fields int
	var total_int_fields int
	var total_other_fields int
	for i := 0; i < len(results); i++ {
		for _, h := range headers {
			switch results[i][h].(type) {
			case string:
				total_str_fields++
			case int64:
				total_int_fields++
			default:
				total_other_fields++
			}
		}
	}

	if 6 != total_str_fields {
		t.Errorf("Unexpected database results, string fields")
	}

	if 9 != total_int_fields {
		t.Errorf("Unexpected database results, int64 fields")
	}

	if 0 != total_other_fields {
		t.Errorf("Unexpected database results, unknown fields")
	}

	postCleanup()
}

func TestQueryExportConfirmResults(t *testing.T) {
	preSetupDb()

	dbargs := []interface{}{"3"}

	qe := queryexport.QEConf{
		DbName:      sqliteTestDB,
		DbType:      "sqlite3",
		DbQuery:     `SELECT * FROM test WHERE col3 = ?`,
		DbQueryArgs: dbargs,
	}

	err := queryexport.QueryExport(qe, testExcelFile)
	if err != nil {
		t.Errorf("Error writing Excel report file: %s", err.Error())
	}
	if _, err := os.Stat(testExcelFile); os.IsNotExist(err) {
		t.Errorf("Excel file not create")
	}

	xlFile, err := xlsx.OpenFile(testExcelFile)
	if err != nil {
		t.Errorf("Cant open Excel file")
	}

	var data []string
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				data = append(data, cell.String())
			}
		}
	}
	if fmt.Sprintf("%s", data) != "[Id Col1 Col2 Col3 Col4 1 test data row1 column1 test data row1 column2 3 4 2 test data row2 column1 test data row2 column2 3 4 3 test data row3 column1 test data row3 column2 3 4]" {
		t.Errorf("Unexpected Excel file contents")

	}

	postCleanup()
}
