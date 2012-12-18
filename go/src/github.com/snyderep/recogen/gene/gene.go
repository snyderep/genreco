package gene

import (
    "database/sql"
    "fmt"
    "math/rand"
    "sort"
    "github.com/snyderep/recogen/database"
)

type Population struct {
    genomes []*Genome
}
// cheesy tournament selection - we consider everyone to be in the tournament.
// Alternatively we could select a random number of genomes from the population
// and select the fittest among those.
func (pop *Population) makeSelection() {
    scores := make([]float64, 0)    
    for i := 0; i < len(pop.genomes); i++ {
        scores = append(scores, pop.genomes[i].score)
    }
    sort.Float64s(scores)

    // choose the top 10% of scores
    topScores := make([]float64, 0)
    topN := int(float64(len(scores)) * 0.1);
    for i := len(scores) - 1; i >= len(scores) - topN; i-- {
        topScores = append(topScores, scores[i])        
    }

    // eliminate any genomes that are not in the topN scores
}

type RecoSet struct {
    products       []*database.Product
    people         []*database.Person
}

type Genome struct {
    rs      *RecoSet
    score   float64
    traits  []Trait
}
func (g *Genome) checkFitness(db *sql.DB, accountId int64) {
    var genScore float64

    // adjust for the number of products 
    if len(g.rs.products) == 0 {
        genScore -= 50.0
    } else if len(g.rs.products) < 15 {
        genScore += 10.0
    } else if len(g.rs.products) < 25 {
        genScore += 15.0
    } else {
        genScore -= 1.0
    }

    // adjust for global conversion
    for i := 0; i < len(g.rs.products); i++ {
        globalConv := database.QueryGlobalConversion(db, accountId, g.rs.products[i])
        if globalConv == 0.0 {
            genScore -= 0.0
        } else if globalConv <= 0.25 {
            genScore += 1.0 
        } else if globalConv <= 0.5 {
            genScore += 2.0
        } else if globalConv <= 0.75 {
            genScore += 3.0
        } else {
            genScore += 4.0
        }
    }
    
    g.score = genScore
}

type Trait interface {
    getName() (string)
    // *database.Person = the original person
    update(*RecoSet, int64, *database.Person) 
}

var allTraits []Trait;

func init() {
    allTraits = append(allTraits, &NopTrait{})
    allTraits = append(allTraits, &ResetWithViewedAndPurchasedTrait{})
    allTraits = append(allTraits, &ResetWithViewedTrait{})
    allTraits = append(allTraits, &ResetWithPurchasedTrait{})
}

func Evolve(maxPopulation int, maxGenerations int, accountId int64, monetateId string) {
    originalPerson := &database.Person{monetateId}

    pop := makeRandomPopulation(maxPopulation, accountId, originalPerson)

    db := database.OpenDB()
    defer db.Close()

    for g := 0; g < maxGenerations; g++ {
        fmt.Printf("processing generation %d\n", g)

        for i := 0; i < len(pop.genomes); i++ {
            // apply the update of the last (current) trait for each genome
            genome := pop.genomes[i]
            currentTrait := genome.traits[len(genome.traits) - 1]
            currentTrait.update(genome.rs, accountId, originalPerson)

            // score the genome (apply the fitness function)
            genome.checkFitness(db, accountId)
            fmt.Printf("genome: %d, score = %f\n", i, genome.score)
        }

        // select genomes to carry forward to the next generation 
        pop.makeSelection()

        // add new traits to the surviving genomes

        // fill in maxPopulation - the new number of genomes with a random population 
        // pop.mergePopulations?
        // the new population will already have a trait
    }
}

func makeRandomPopulation(size int, accountId int64, originalPerson *database.Person) (pop *Population) {
    db := database.OpenDB()
    defer db.Close()

    pop = &Population{}

    traits := make([]Trait, 1)
    traits = append(traits, selectRandomTrait())

    genomes := make([]*Genome, size)

    for i := 0; i < size; i++ {
        genome := &Genome{
            rs: &RecoSet{}, 
            score: 0.0,
            traits: traits}
        genomes[i] = genome
    }

    pop.genomes = genomes

    return
}

func selectRandomTrait() (trait Trait) {
    n := rand.Intn(len(allTraits))
    trait = allTraits[n]
    return
}

type NopTrait struct {}
func (t *NopTrait) getName() (string) {
    return "nop" 
}
func (t *NopTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    // do nothing, this is a nop after all
}

type ResetWithViewedAndPurchasedTrait struct {}
func (t *ResetWithViewedAndPurchasedTrait) getName() (string) {
    return "reset with viewed and purchased"
}
// ignore any existing products or people and bootstrap from the original person
func (t *ResetWithViewedAndPurchasedTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    db := database.OpenDB()
    defer db.Close()

    products := database.QueryProductsViewedAndPurchased(db, accountId, origPerson.MonetateId)
    people := make([]*database.Person, 1)
    people = append(people, origPerson)

    rs.people = people
    rs.products = products
}

type ResetWithViewedTrait struct {}
func (t *ResetWithViewedTrait) getName() (string) {
    return "reset with viewed"
}
// ignore any existing products or people and bootstrap from the original person
func (t *ResetWithViewedTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    db := database.OpenDB()
    defer db.Close()

    products := database.QueryProductsViewed(db, accountId, origPerson.MonetateId)
    people := make([]*database.Person, 1)
    people = append(people, origPerson)

    rs.people = people
    rs.products = products
}

type ResetWithPurchasedTrait struct {}
func (t *ResetWithPurchasedTrait) getName() (string) {
    return "reset with viewed"
}
// ignore any existing products or people and bootstrap from the original person
func (t *ResetWithPurchasedTrait) update(rs *RecoSet, accountId int64, origPerson *database.Person) {
    db := database.OpenDB()
    defer db.Close()

    products := database.QueryProductsPurchased(db, accountId, origPerson.MonetateId)
    people := make([]*database.Person, 1)
    people = append(people, origPerson)

    rs.people = people
    rs.products = products
}
