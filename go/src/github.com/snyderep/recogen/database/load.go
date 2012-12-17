package database

import (
    "bufio"
    "encoding/csv"
    "io"
    "os"
    "strconv"
)

func LoadProducts() {
    db := openDB()
    defer db.Close()
    
    trans, err := db.Begin()
    if err != nil {
        panic(err)
    }

    stmt := getInsertProductStmt(trans)
    defer stmt.Close()

    err := deleteAllProducts(trans)
    if err != nil {
        panic(err)
    }

    file, err := os.Open("/Users/esnyder/prj/w/genreco/data/products.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    reader := bufio.NewReader(file)
    csvReader := csv.NewReader(reader)
    csvReader.Comma = '\t'

    for {
        record, err := csvReader.Read()
        if err == io.EOF {
            break
        } else if err != nil {
            trans.Rollback()
            panic(err)
        }

        accountId, err := strconv.ParseInt(record[0], 10, 32) 
        if err != nil {
            trans.Rollback()
            panic(err)
        }
        unitPrice, err := strconv.ParseFloat(record[5], 32)
        if err != nil {
            trans.Rollback()
            panic(err)
        }

        p := &Product{accountId: accountId, pid: record[1], name: record[2], productUrl: record[3],
                      imageUrl: record[4], unitCost: 0.0, unitPrice: unitPrice, margin: 0.0,
                      marginRate: 0.0}
        insertProduct(stmt, p)
    }

    trans.Commit()
}
