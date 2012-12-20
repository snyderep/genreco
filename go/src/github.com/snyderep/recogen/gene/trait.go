package gene

import (
    "math/rand"
    "github.com/snyderep/recogen/database"
)

type Trait interface {
    String() (string)
    // *database.Person = the original person
    update(*RecoSet, int64, *database.Person) 
}

var allTraits []Trait;

func init() {
    allTraits = append(allTraits, &NopTrait{})
    allTraits = append(allTraits, &PeopleThatViewedProductsTrait{})
    allTraits = append(allTraits, &ProductsViewedByPeopleTrait{})  
    allTraits = append(allTraits, &RandomProductTrait{})
    allTraits = append(allTraits, &RandomProductDeleteTrait{})
    allTraits = append(allTraits, &SoundAlikeProductTrait{})
}

type NopTrait struct {}
func (t *NopTrait) String() (string) {
    return "nop" 
}
func (t *NopTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    // do nothing, this is a nop after all
}

type PeopleThatViewedProductsTrait struct {}
func (t *PeopleThatViewedProductsTrait) String() (string) {
    return "people that viewed products"
}
func (t *PeopleThatViewedProductsTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    db := database.OpenDB()
    defer db.Close()

    people := database.QueryPeopleThatViewedProducts(db, accountId, rs.products)
    peeps := make(map[string]*database.Person)
    for i := 0; i < len(people); i++ {
        //rs.people[people[i].MonetateId] = people[i]
        peeps[people[i].MonetateId] = people[i]
    }
    rs.people = peeps
}

type ProductsViewedByPeopleTrait struct {}
func (t *ProductsViewedByPeopleTrait) String() (string) {
    return "products viewed by people"
}
func (t *ProductsViewedByPeopleTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    db := database.OpenDB()
    defer db.Close()

    products := database.QueryProductsViewedByPeople(db, accountId, rs.people)
    for i := 0; i < len(products); i++ {
        rs.products[products[i].Pid] = products[i]
    }
}

type RandomProductTrait struct {}
func (t *RandomProductTrait) String() (string) {
    return "random product"
}
func (t *RandomProductTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    db := database.OpenDB()
    defer db.Close()

    product := database.QueryRandomProduct(db, accountId, origPerson)
    if product != nil {
        rs.products[product.Pid] = product
    }
}

type RandomProductDeleteTrait struct {}
func (t *RandomProductDeleteTrait) String() (string) {
    return "random product delete"
}
func (t *RandomProductDeleteTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    for pid, _ := range rs.products {
        coin := rand.Intn(10)
        if coin == 0 {delete(rs.products, pid)}
    }
}

type SoundAlikeProductTrait struct {}
func (t *SoundAlikeProductTrait) String() (string) {
    return "sound alike product"
}
func (t *SoundAlikeProductTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    db := database.OpenDB()
    defer db.Close()

    if len(rs.products) > 0 {
        // take advantage of the fact that go randomizes the iteration order of map items
        var inProduct *database.Product
        for _, p := range rs.products {
            inProduct = p
            break
        }
        outProduct := database.QuerySoundAlikeProduct(db, accountId, inProduct)
        if outProduct != nil {
            rs.products[outProduct.Pid] = outProduct
        }        
    }
}
