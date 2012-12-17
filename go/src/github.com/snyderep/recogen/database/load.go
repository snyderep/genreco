package database

import (
    "bufio"
    "database/sql"
    "encoding/csv"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strconv"
)

const dataBasePath = "/Users/esnyder/prj/w/genreco/data/" 

//func init() {
//    dataBasePath = "/Users/esnyder/prj/w/genreco/data/"
//}

func openDataFile(filename string) (file *os.File) {
    newPath := filepath.Join(dataBasePath, filename)
    file, err := os.Open(newPath)
    if err != nil {
        panic(err)
    }
    return
}

func getTabReader(file *os.File) (csvReader *csv.Reader) {
    reader := bufio.NewReader(file)
    csvReader = csv.NewReader(reader)
    csvReader.Comma = '\t'
    return
}

func LoadAllData() {
    db := openDB()
    defer db.Close() 

    fmt.Println("loading products")
    LoadProducts(db)
    fmt.Println("loading user product views")
    LoadUserProductViews(db)
}

func LoadUserProductViews(db *sql.DB) {
    var trans *sql.Tx
    var stmt *sql.Stmt
    var err error
    
    trans, err = db.Begin()
    if err != nil {
        panic(err)
    }

    err = deleteAllUserProductViews(trans)
    if err != nil {
        panic(err)
    }
  
    file := openDataFile("user_products_viewed.txt")
    defer file.Close()

    stmt = getInsertUserProductViewStmt(trans)
    defer stmt.Close()

    tabReader := getTabReader(file)

    c := 0

    for {
        record, err := tabReader.Read()
        if err == io.EOF {
            break
        } else if err == nil {
            if c % 10000 == 0 {
                trans.Commit()
                stmt.Close()

                trans, err = db.Begin()
                if err != nil {
                    panic(err)
                }
                stmt = getInsertUserProductViewStmt(trans)
            }
        } else {
            panic(err)
        }

        accountId, err := strconv.ParseInt(record[0], 10, 32) 
        if err != nil {
            panic(err)
        }
        count, err := strconv.ParseInt(record[3], 10, 32) 
        if err != nil {
            panic(err)
        }

        err = insertUserProductView(stmt, accountId, record[1], record[2], count)
        if err != nil {
            panic(err)
        }
        
        c += 1
    }

    trans.Commit()
}

func LoadProducts(db *sql.DB) {
    trans, err := db.Begin()
    if err != nil {
        panic(err)
    }

    err = deleteAllProducts(trans)
    if err != nil {
        panic(err)
    }

    file := openDataFile("products.txt")
    defer file.Close()

    stmt := getInsertProductStmt(trans)
    defer stmt.Close()

    tabReader := getTabReader(file)

    for {
        record, err := tabReader.Read()
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
        err = insertProduct(stmt, p)
        if err != nil {
            panic(err)
        }
    }

    trans.Commit()
}
