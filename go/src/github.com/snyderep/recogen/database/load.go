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
	db := OpenDB()
	defer db.Close()

	ch := make(chan string, 10)

	go LoadProducts(db, ch)
	go LoadUserProductViews(db, ch)
	go LoadUserProductPurchases(db, ch)
	go LoadProductConversionRates(db, ch)

	// drain the channel, there are 4 tasks to wait for
	for i := 0; i < 4; i++ {
		task := <-ch // wait for one task to complete
		fmt.Println(task + " done")
	}
}

func LoadUserProductViews(db *sql.DB, ch chan string) {
	var trans *sql.Tx
	var stmt *sql.Stmt
	var err error

	fmt.Println("loading user product views")

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
			if c%10000 == 0 {
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

		err = insertUserProduct(stmt, accountId, record[1], record[2], count)
		if err != nil {
			panic(err)
		}

		c += 1
	}

	trans.Commit()

	ch <- "user product views" // signal that we're done
}

func LoadUserProductPurchases(db *sql.DB, ch chan string) {
	var trans *sql.Tx
	var stmt *sql.Stmt
	var err error

	fmt.Println("loading user product purchases")
	trans, err = db.Begin()
	if err != nil {
		panic(err)
	}

	err = deleteAllUserProductPurchases(trans)
	if err != nil {
		panic(err)
	}

	file := openDataFile("user_products_purchased.txt")
	defer file.Close()

	stmt = getInsertUserProductPurchaseStmt(trans)
	defer stmt.Close()

	tabReader := getTabReader(file)

	c := 0

	for {
		record, err := tabReader.Read()
		if err == io.EOF {
			break
		} else if err == nil {
			if c%10000 == 0 {
				trans.Commit()
				stmt.Close()

				trans, err = db.Begin()
				if err != nil {
					panic(err)
				}
				stmt = getInsertUserProductPurchaseStmt(trans)
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

		err = insertUserProduct(stmt, accountId, record[1], record[2], count)
		if err != nil {
			panic(err)
		}

		c += 1
	}

	trans.Commit()

	ch <- "user product purchases" // signal that we're done
}

func LoadProductConversionRates(db *sql.DB, ch chan string) {
	var trans *sql.Tx
	var stmt *sql.Stmt
	var err error

	fmt.Println("loading product conversion rates")

	trans, err = db.Begin()
	if err != nil {
		panic(err)
	}

	err = deleteAllProductConversionRates(trans)
	if err != nil {
		panic(err)
	}

	file := openDataFile("global_conversion_rate.txt")
	defer file.Close()

	stmt = getInsertProductConversionRateStmt(trans)
	defer stmt.Close()

	tabReader := getTabReader(file)

	c := 0

	for {
		record, err := tabReader.Read()
		if err == io.EOF {
			break
		} else if err == nil {
			if c%10000 == 0 {
				trans.Commit()
				stmt.Close()

				trans, err = db.Begin()
				if err != nil {
					panic(err)
				}
				stmt = getInsertProductConversionRateStmt(trans)
			}
		} else {
			panic(err)
		}

		accountId, err := strconv.ParseInt(record[0], 10, 32)
		if err != nil {
			panic(err)
		}
		conversionRate, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			panic(err)
		}

		err = insertProductConversionRate(stmt, accountId, record[1], conversionRate)
		if err != nil {
			panic(err)
		}

		c += 1
	}

	trans.Commit()

	ch <- "product conversion rates" // signal that we're done
}

func LoadProducts(db *sql.DB, ch chan string) {
	fmt.Println("loading products")

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

		p := &Product{AccountId: accountId, Pid: record[1], Name: record[2], ProductUrl: record[3],
			ImageUrl: record[4], UnitCost: 0.0, UnitPrice: unitPrice, Margin: 0.0,
			MarginRate: 0.0}
		err = insertProduct(stmt, p)
		if err != nil {
			panic(err)
		}
	}

	trans.Commit()

	ch <- "products" // signal that we're done
}
