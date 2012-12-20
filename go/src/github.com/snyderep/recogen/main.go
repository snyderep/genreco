package main

import (
    "flag"
	"github.com/snyderep/recogen/database"
	"github.com/snyderep/recogen/gene"
)

var loadData bool

func init() {
    flag.BoolVar(&loadData, "load", false, "load all data")
}

func main() {
    flag.Parse()

    if loadData {
	    database.LoadAllData()
    } else {
        gene.Evolve(10, 100, 321, "2.1001298975.1355107162879")
    }
}
