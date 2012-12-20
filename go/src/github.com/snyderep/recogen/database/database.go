package database

import (
	"database/sql"
    //"fmt"
    "strings"
	_ "github.com/bmizerany/pq"
)

func deleteAllProducts(trans *sql.Tx) (err error) {
	_, err = trans.Exec("DELETE FROM product")
	return
}
func deleteAllUserProductViews(trans *sql.Tx) (err error) {
	_, err = trans.Exec("DELETE FROM user_product_views")
	return
}
func deleteAllUserProductPurchases(trans *sql.Tx) (err error) {
	_, err = trans.Exec("DELETE FROM user_product_purchases")
	return
}
func deleteAllProductConversionRates(trans *sql.Tx) (err error) {
	_, err = trans.Exec("DELETE FROM product_conversion_rate")
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
func getInsertUserProductViewStmt(trans *sql.Tx) (stmt *sql.Stmt) {
	// note that postgresql uses $1, $2, etc while others use ?
	s := "INSERT INTO user_product_views (account_id, monetate_id, pid, count) " +
		"VALUES ($1, $2, $3, $4)"
	var err error
	stmt, err = trans.Prepare(s)
	if err != nil {
		panic(err)
	}
	return
}
func getInsertUserProductPurchaseStmt(trans *sql.Tx) (stmt *sql.Stmt) {
	// note that postgresql uses $1, $2, etc while others use ?
	s := "INSERT INTO user_product_purchases (account_id, monetate_id, pid, count) " +
		"VALUES ($1, $2, $3, $4)"
	var err error
	stmt, err = trans.Prepare(s)
	if err != nil {
		panic(err)
	}
	return
}
func getInsertProductConversionRateStmt(trans *sql.Tx) (stmt *sql.Stmt) {
	// note that postgresql uses $1, $2, etc while others use ?
	s := "INSERT INTO product_conversion_rate (account_id, pid, conversion_rate) " +
		"VALUES ($1, $2, $3)"
	var err error
	stmt, err = trans.Prepare(s)
	if err != nil {
		panic(err)
	}
	return
}

func insertProduct(stmt *sql.Stmt, p *Product) (err error) {
	_, err = stmt.Exec(p.AccountId, p.Pid, p.Name, p.ProductUrl, p.ImageUrl,
		p.UnitCost, p.UnitPrice, p.Margin, p.MarginRate)
	return
}
func insertUserProduct(stmt *sql.Stmt, accountId int64, monetateId string, pid string,
	count int64) (err error) {

	_, err = stmt.Exec(accountId, monetateId, pid, count)
	return
}
func insertProductConversionRate(stmt *sql.Stmt, accountId int64, pid string, 
    conversionRate float64) (err error) {

	_, err = stmt.Exec(accountId, pid, conversionRate)
	return
}

func QueryPeopleThatViewedProducts(db *sql.DB, accountId int64, products map[string]*Product) (people []*Person) {
    s := []string{}

    s = append(s, "SELECT DISTINCT monetate_id")
    s = append(s, "FROM user_product_views")
    s = append(s, "WHERE account_id = $1 AND pid = $2")
    s = append(s, "AND RANDOM() < 0.01")    
    s = append(s, "LIMIT 2")

    query := strings.Join(s, " ") 

    stmt, err := db.Prepare(query)
    defer stmt.Close()
    if err != nil {panic(err)}

    people = make([]*Person, 0)

    for pid, _ := range products {
        rows, err := stmt.Query(accountId, pid)
        if err != nil {panic(err)}

        for rows.Next() {
            p := &Person{}
            err = rows.Scan(&p.MonetateId)
            if err != nil {panic(err)}
            people = append(people, p) 
        }

        err = rows.Err()
        if err != nil {panic(err)}        
    }

    return
}

func QueryProductsViewedByPeople(db *sql.DB, accountId int64, people map[string]*Person) (products []*Product) {
    s := []string{}

    s = append(s, "SELECT")
    s = append(s, "p.account_id, p.pid, p.name, p.product_url, p.image_url, p.unit_cost,")
    s = append(s, "p.unit_price, p.margin, p.margin_rate")    
    s = append(s, "FROM user_product_views u JOIN product p ON (")        
    s = append(s, "u.account_id = p.account_id AND")        
    s = append(s, "u.pid = p.pid)")    
    s = append(s, "WHERE u.account_id = $1 AND u.monetate_id = $2")

    query := strings.Join(s, " ") 

    stmt, err := db.Prepare(query)
    defer stmt.Close()    
    if err != nil {panic(err)}

    products = make([]*Product, 0)

    for pid, _ := range people {
        rows, err := stmt.Query(accountId, pid)
        if err != nil {panic(err)}

        for rows.Next() {
            p := &Product{}
            err = rows.Scan(&p.AccountId, &p.Pid, &p.Name, &p.ProductUrl, &p.ImageUrl, &p.UnitCost, 
                            &p.UnitPrice, &p.Margin, &p.MarginRate)
            if err != nil {panic(err)}
            products = append(products, p) 
        }

        err = rows.Err()
        if err != nil {panic(err)}        
    }

    return
}

func QueryProductsViewed(db *sql.DB, accountId int64, person *Person) (products []*Product) {
    people := make(map[string]*Person)
    people[person.MonetateId] = person
    products = QueryProductsViewedByPeople(db, accountId, people)
    return
}

func QueryProductsPurchasedByPeople(db *sql.DB, accountId int64, people map[string]*Person) (products []*Product) {
    s := []string{}

    s = append(s, "SELECT")
    s = append(s, "p.account_id, p.pid, p.name, p.product_url, p.image_url, p.unit_cost,")
    s = append(s, "p.unit_price, p.margin, p.margin_rate")    
    s = append(s, "FROM user_product_purchases u JOIN product p ON (")        
    s = append(s, "u.account_id = p.account_id AND")        
    s = append(s, "u.pid = p.pid)")    
    s = append(s, "WHERE u.account_id = $1 AND u.monetate_id = $2")

    query := strings.Join(s, " ") 

    stmt, err := db.Prepare(query)
    defer stmt.Close()    
    if err != nil {panic(err)}

    products = make([]*Product, 0)

    for pid, _ := range people {
        rows, err := stmt.Query(accountId, pid)
        if err != nil {panic(err)}

        for rows.Next() {
            p := &Product{}
            err = rows.Scan(&p.AccountId, &p.Pid, &p.Name, &p.ProductUrl, &p.ImageUrl, &p.UnitCost, 
                            &p.UnitPrice, &p.Margin, &p.MarginRate)
            if err != nil {panic(err)}
            products = append(products, p) 
        }

        err = rows.Err()
        if err != nil {panic(err)}        
    }

    return
}

func QueryProductsPurchased(db *sql.DB, accountId int64, person *Person) (products []*Product) {
    people := make(map[string]*Person)
    people[person.MonetateId] = person
    products = QueryProductsPurchasedByPeople(db, accountId, people)
    return
}

func QueryProductsViewedAndPurchased(db *sql.DB, accountId int64, person *Person) (allProducts []*Product) {
    products := QueryProductsViewed(db, accountId, person)
    purchProducts := QueryProductsPurchased(db, accountId, person)

    // concatenate the slices, no there's no convenient way to do this
    allProducts = make([]*Product, len(products) + len(purchProducts))
    copy(allProducts, products)
    copy(allProducts[len(products):], purchProducts)

    return
}

func QueryRandomProduct(db *sql.DB, accountId int64, person *Person) (product *Product) {
    s := []string{}

    s = append(s, "SELECT")
    s = append(s, "account_id, pid, name, product_url, image_url, unit_cost,")
    s = append(s, "unit_price, margin, margin_rate")    
    s = append(s, "FROM product")      
    s = append(s, "WHERE account_id = $1")
    s = append(s, "AND RANDOM() < 0.01")    
    s = append(s, "LIMIT 1")
    
    query := strings.Join(s, " ") 

    row := db.QueryRow(query, accountId)
    product = &Product{}
    err := row.Scan(&product.AccountId, &product.Pid, &product.Name, &product.ProductUrl, &product.ImageUrl, 
                    &product.UnitCost, &product.UnitPrice, &product.Margin, &product.MarginRate)
    if err != nil {
        if err == sql.ErrNoRows {
            product = nil
        } else {
            panic(err)
        }
    } 

    return
}

func QuerySoundAlikeProduct(db *sql.DB, accountId int64, inProduct *Product) (product *Product) {
    s := []string{}

    s = append(s, "SELECT")
    s = append(s, "account_id, pid, name, product_url, image_url, unit_cost,")
    s = append(s, "unit_price, margin, margin_rate")    
    s = append(s, "FROM product")      
    s = append(s, "WHERE account_id = $1")
    s = append(s, "AND difference(name, $2) >= 3")    
    s = append(s, "LIMIT 1")
    
    query := strings.Join(s, " ") 

    row := db.QueryRow(query, accountId, inProduct.Name)
    product = &Product{}
    err := row.Scan(&product.AccountId, &product.Pid, &product.Name, &product.ProductUrl, &product.ImageUrl, 
                    &product.UnitCost, &product.UnitPrice, &product.Margin, &product.MarginRate)
    if err != nil {
        if err == sql.ErrNoRows {
            product = nil
        } else {
            panic(err)
        }
    } 

    return
}

func HasProductBeenSeenByPerson(db *sql.DB, accountId int64, person *Person, product *Product) (seen bool) {
    s := []string{}

    s = append(s, "SELECT 'x'")
    s = append(s, "FROM user_product_views u JOIN product p ON (")        
    s = append(s, "u.account_id = p.account_id AND")        
    s = append(s, "u.pid = p.pid)")    
    s = append(s, "WHERE u.account_id = $1 AND u.monetate_id = $2 AND u.pid = $3")
    
    query := strings.Join(s, " ") 

    row := db.QueryRow(query, accountId, person.MonetateId, product.Pid)
    var foo string
    err := row.Scan(&foo)
    if err == nil {
        seen = true
    } else {
        if err == sql.ErrNoRows {
            seen = false
        } else {
            panic(err)
        }
    } 

    return
}

func HasProductBeenPurchasedByPerson(db *sql.DB, accountId int64, person *Person, product *Product) (seen bool) {
    s := []string{}

    s = append(s, "SELECT 'x'")
    s = append(s, "FROM user_product_purchases u JOIN product p ON (")        
    s = append(s, "u.account_id = p.account_id AND")        
    s = append(s, "u.pid = p.pid)")    
    s = append(s, "WHERE u.account_id = $1 AND u.monetate_id = $2 AND u.pid = $3")
    
    query := strings.Join(s, " ") 

    row := db.QueryRow(query, accountId, person.MonetateId, product.Pid)
    var foo string
    err := row.Scan(&foo)
    if err == nil {
        seen = true
    } else {
        if err == sql.ErrNoRows {
            seen = false
        } else {
            panic(err)
        }
    } 

    return
}

func QueryGlobalConversion(db *sql.DB, accountId int64, product *Product) (conversionRate float64) {
    s := []string{}

    s = append(s, "SELECT conversion_rate")
    s = append(s, "FROM product_conversion_rate")
    s = append(s, "WHERE account_id = $1 AND pid = $2")
    
    query := strings.Join(s, " ") 

    row := db.QueryRow(query, accountId, product.Pid)
    err := row.Scan(&conversionRate)
    if err != nil {
        if err == sql.ErrNoRows {
            conversionRate = 0.0
        } else {
            panic(err)
        }
    } 

    return
}

func OpenDB() (db *sql.DB) {
	db, err := sql.Open("postgres", "dbname=recogen sslmode=disable")
	if err == nil {
		return
	}
	panic(err)
}
