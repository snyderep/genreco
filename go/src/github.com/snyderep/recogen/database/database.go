package database

import (
    "database/sql"
    _ "github.com/bmizerany/pq"
)

type Product struct {
    accountId  int64
    pid        string
    name       string
    productUrl string
    imageUrl   string
    unitCost   float64
    unitPrice  float64
    margin     float64
    marginRate float64
}

func deleteAllProducts(trans *sql.Tx) (err error) {
    _, err = trans.Exec("DELETE FROM product")
    return
}

func getInsertProductStmt(trans *sql.Tx) (stmt *sql.Stmt) {
    // note that postgresql uses $1, $2, etc while others use ?
    s := "INSERT INTO product (account_id, pid, name, product_url, image_url, unit_cost, " +
          "unit_price, margin, margin_rate) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
    var err error
    stmt, err = trans.Prepare(s)
    if err != nil {
        panic(err)
    }
    return 
}

func insertProduct(stmt *sql.Stmt, p *Product) (err error) {
    _, err = stmt.Exec(p.accountId, p.pid, p.name, p.productUrl, p.imageUrl, 
                                   p.unitCost, p.unitPrice, p.margin, p.marginRate)
    return
}

func openDB() (db *sql.DB) {
    db, err := sql.Open("postgres", "dbname=recogen sslmode=disable")
    if (err == nil) {
        return
    }
    panic(err)
}
