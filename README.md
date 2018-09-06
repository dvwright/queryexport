# queryexport
Query Export provides a simple way to export the results of a database query to an Excel file

It's a fairly common request for 'Business' to want an Excel report from data stored in a database.

Rows are written out in the same order as received by the database query.

Query Export auto detects your database table schema no need to provide a struct of the schema.

# How To Use it?

Set your database connection details, provide a query and run QueryExport

```go
package main

import (
    "github.com/dvwright/queryexport"
    "fmt"
)

func main() {

        // the arguments (if any) to your query (i.e. the placeholder values '?')
        dbargs := []interface{}{"1", "business"}

        qe := queryexport.QEConf{
                User:        "my_mysql_user",
                Pass:        "my_mysql_pass
                Host:        "localhost",
                Port:        "3306",
                DbName:      "my_database",
                DbType:      "mysql",
                DbQuery:     `SELECT * FROM my_table WHERE id > ? AND name = ?`,
                DbQueryArgs: dbargs,
        }

        file := "my_excel_report.xlsx"
        err := queryexport.QueryExport(qe, file)
        if err != nil {
                fmt.Printf("Error writing Excel report file: %s", err.Error())
        }
}
```

# Additional Uses

Can be integrated/used from your go code to create an xlsx report from a database query.

```go 

rows, err := db.Query(`SELECT * FROM mytable`)
err := queryexport.RowsToReport(rows, "fileToWrite.xlsx") 

```

Can save time serializing database table schema data. Usually with go and sql, you provide a struct which represents your table schema.
Then `rows.scan` to your struct and again iterate over your struct to serialze to a file. On really big tables, 
or doing many tables this can get tedious, especially if getting the data to a file is your end goal.
i.e. not worrying about performance/memory, maintaining state or data integrity or additionally transforming your data.

Namely 'MapRows', creates a map from database query results without defining a struct.

Additionally, can be useful debugging table schema info. If for some reason you have no way of viewing your database table schema, you can run this and dump the headers.

```go
        // error handling omitted for clarity

        db, err := sql.Open("sqlite3", sqliteTestDB)
        rows, err := db.Query(`SELECT * FROM test`)

        headers_count, headers, results, err := queryexport.MapRows(rows)
	// show database table column names
        fmt.Printf("%#v\n", headers)

	// get total rows result count
        fmt.Printf("%d\n", len(results))

        // iterate values
        for i := 0; i < len(results); i++ {
                for _, h := range headers {
                        switch v := results[i][h].(type) {
                        case string:
		                fmt.Printf("String Data: %s\n", v)
                        case int64:
		                fmt.Printf("Int Data: %d\n", v)
                        case float64:
		                fmt.Printf("Float Data: %f\n", v)
                        default:
		                fmt.Printf("Other Data: %v\n", v)
                        }
                }
        }
```

A title case transform is performed on the field name for the Excel headers. i.e. 'field_name' becomes FieldName

# Data

# NOTE: This is Beta level code with minimal actual real world usage

## Contributing

You are welcome to contribute to this project.

Please update/add tests as appropriate, send pull request.

Please use the same formatting as the Go authors. Run code through gofmt before submitting.

Thanks

# Language

Query Export is written in [Go](https://golang.org/)


### Misc ###

Currently only supports mysql and sqlite3 databases
