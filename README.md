# Sql-Glue
SQL Glue is library for GO which speeding up constructing SQL queries. It generate input for sqlx or other libraries for SQL databases.

## Features
- Escaping variables and identifiers
- Joining AND and OR conditions
- Generate SET part
- Splitting Structs and Maps

## Basic example
There is just basic example how it is work. For more examples look to CookBook(in progress).
[Try in go playground](https://play.golang.com/p/JGPFw-L-j9h)

```go
package main

import (
    "strings"
    "fmt"

    "github.com/lib/pq"
    sqlg "github.com/mzahradnicek/sql-glue"
)

type filter map[string]interface{}

var builder *sqlg.Builder

func generateQuery(f filter) (string, []interface{}, error) {
    where := sqlg.Qg{"deleted = 0"}

    if v, ok := f["id"]; ok {
        where.Append("id = %v", v)  // automatically escaped value if string
    }

    if v, ok := f["name"]; ok {
        where.Append("name = %v", v)
    }

    q := &sqlg.Qg{"SELECT * FROM users WHERE %and", where}
    return builder.Glue(q)
}

func main() {
    builder = sqlg.NewBuilder(sqlg.Config{
        KeyModifier:      strings.ToLower,
        IdentifierEscape: pq.QuoteIdentifier,
        PlaceholderInit:  sqlg.PqPlaceholder,
        Tag:              "sqlg",
    })

    query, data, err := generateQuery(filter{"id": 5, "name": "John Smith"})
    fmt.Printf("%#v\n\n%#v\n\n%#v", query, data, err)
	// then use with sqlx.Select(dest, query, data...)
}
```
