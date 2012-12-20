package gene

import (
    "database/sql"
    "fmt"
    "math/rand"
    "sort"
    "time"
    "github.com/snyderep/recogen/database"
)

func init() {
    rand.Seed( time.Now().UTC().UnixNano())    
}

type Population struct {
    genomes []*Genome
}
func (pop *Population) evolve(maxPopulation int, maxGenerations int, accountId int64, 
    originalPerson *database.Person) {

    db := database.OpenDB()
    defer db.Close()

    for g := 0; g < maxGenerations; g++ {
        fmt.Printf("processing generation %d\n", g)

        ch := make(chan bool)
        for i := 0; i < len(pop.genomes); i++ {
            go func(ch chan bool, genome *Genome) {
                // apply the update of the last (current) trait a genome                
                genome.getCurrentTrait().update(genome.rs, accountId, originalPerson) 
                genome.checkFitness(db, accountId, originalPerson)

                ch <- true
            }(ch, pop.genomes[i])
        }
        // drain the channel
        for i := 0; i < len(pop.genomes); i++ {<-ch}

        pop.display()                

        if g == (maxGenerations - 1) {
            pop.displayFinal()
        } else {
            // select genomes to carry forward to the next generation             
            pop.makeSelection()

            // add new traits to the surviving genomes
            for i := 0; i < len(pop.genomes); i++ {
                pop.genomes[i].addRandomTrait()
            }        

            // have the successful ones reproduce to fill out the remainder of the population
            childrenGenomes := make([]*Genome, 0)
            for i := 0; i < (maxPopulation - len(pop.genomes)); i++ {
                r1 := rand.Intn(len(pop.genomes))
                r2 := rand.Intn(len(pop.genomes))
                newGenome := reproduce(pop.genomes[r1], pop.genomes[r2])
                childrenGenomes = append(childrenGenomes, newGenome)
            }
            pop.appendGenomes(childrenGenomes)
        }
    }    
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

    // choose the top 50% of scores
    topScores := make([]float64, 0)
    topN := int(float64(len(scores)) * 0.5);
    for i := len(scores) - 1; i >= len(scores) - topN; i-- {
        topScores = append(topScores, scores[i])        
    }
    sort.Float64s(topScores)    
    min := topScores[0]
    max := topScores[len(topScores) - 1]

    // build a new list of genomes that are in topScores
    selectedGenomes := make([]*Genome, 0)
    fmt.Printf("min, max scores: %f, %f\n", min, max)
    for i := 0; i < len(pop.genomes); i++ {
        genome := pop.genomes[i]
        if genome.score >= min && genome.score <= max {
            selectedGenomes = append(selectedGenomes, genome)
        }
    }

    pop.genomes = selectedGenomes
}
func (pop *Population) display() {
    for i := 0; i < len(pop.genomes); i++ {
        genome := pop.genomes[i]
        fmt.Printf("genome %d, score: %f, products: %d\n", i, genome.score, 
            genome.getProductsCount())
    }
}
func (pop *Population) getHighestScoringGenome() (bestGenome *Genome) {
    for i := 0; i < len(pop.genomes); i++ {
        genome := pop.genomes[i]
        if bestGenome == nil || genome.score > bestGenome.score {
            bestGenome = genome
        }
    }
    return
}
func (pop *Population) displayFinal() {
    fmt.Println("********** DONE **********")

    bestGenome := pop.getHighestScoringGenome()
    for _, product := range bestGenome.getProducts() {
        fmt.Println(product.String())
        fmt.Println("**************************")    
    }
    fmt.Printf("Score: %f\n", bestGenome.score)
}
func (pop *Population) appendGenomes(genomes []*Genome) {
    for i := 0; i < len(genomes); i++ {
        pop.genomes = append(pop.genomes, genomes[i])
    }        
}

type RecoSet struct {
    products map[string]*database.Product
    people   map[string]*database.Person
}

type Genome struct {
    rs      *RecoSet
    score   float64
    traits  []Trait
}
func (g *Genome) checkFitness(db *sql.DB, accountId int64, originalPerson *database.Person) {
    countScore := float64(0.0)
    convScore := float64(0.0)
    seenScore := float64(0.0)
    purchScore := float64(0.0)

    // adjust for the number of products 
    productCount := g.getProductsCount()    
    switch {
    case productCount == 0:
        countScore = 0.0
    case productCount > 0 && productCount <= 5:
        countScore = 5.0
    case productCount > 5 && productCount <= 10:
        countScore = 10.0
    case productCount > 10 && productCount <= 20:
        countScore = 15.0        
    case productCount > 20 && productCount <= 50:
        countScore = 0.0                
    default:
        countScore = -5.0
    }

    for _, prod := range g.rs.products {
        var score float64

        score = float64(0.0)
        conv := database.QueryGlobalConversion(db, accountId, prod)
        switch {
            case conv == 0.0:
                score = 0.0
            case conv > 0.0 && conv <= 0.25:
                score = 1.0
            case conv > 0.25 && conv <= 0.5:
                score = 3.0
            case conv > 0.5 && conv <= 0.75:
                score = 4.0
            default:
                score = 5.0
        }
        convScore += score

        score = float64(0.0)
        if database.HasProductBeenSeenByPerson(db, accountId, originalPerson, prod) {
            score = 0.0
        } else {
            score = 5.0
        }            
        seenScore += score

        score = float64(0.0)
        if database.HasProductBeenPurchasedByPerson(db, accountId, originalPerson, prod) {
            score = -10.0
        } else {
            score = 0.0
        }       
        purchScore += score
    }

    pCount := float64(g.getProductsCount())
    convScore = convScore / pCount
    seenScore = seenScore / pCount
    purchScore = purchScore / pCount

    g.score = (countScore * 0.4) + (convScore * 0.2) + (seenScore * 0.1) + (purchScore * 0.3)
}
func (g *Genome) getCurrentTrait() (trait Trait) {
    if len(g.traits) == 0 {
        trait = nil
    } else {
        trait = g.traits[len(g.traits) - 1]
    }
    return
}
func (g *Genome) addRandomTrait() {
    var t Trait
    traitsCount := len(allTraits)
    for i := 0; i < traitsCount; i++ {
        n := rand.Intn(traitsCount)
        t = allTraits[n]
        if t != g.getCurrentTrait() {
            break
        }
    }
    g.traits = append(g.traits, t)
}
func (g *Genome) getPeopleCount() (count int) {
    count = len(g.rs.people)
    return
}
func (g * Genome) getPeople() (products map[string]*database.Person) {
    return g.rs.people
}
func (g *Genome) getProductsCount() (count int) {
    count = len(g.rs.products)
    return
}
func (g * Genome) getProducts() (products map[string]*database.Product) {
    return g.rs.products
}

func Run(maxPopulation int, maxGenerations int, accountId int64, monetateId string) {
    originalPerson := &database.Person{monetateId}

    pop := makeRandomPopulation(maxPopulation, accountId, originalPerson)
    pop.evolve(maxPopulation, maxGenerations, accountId, originalPerson)
}

func makeRandomPopulation(size int, accountId int64, originalPerson *database.Person) (pop *Population) {
    db := database.OpenDB()
    defer db.Close()

    pop = &Population{}

    genomes := make([]*Genome, size)

    for i := 0; i < size; i++ {
        // seed with the original person
        persMap := make(map[string]*database.Person)
        persMap[originalPerson.MonetateId] = originalPerson

        // seed with the original person's products
        prodMap := make(map[string]*database.Product)
        // TODO: Don't requery for every genome
        products := database.QueryProductsViewedAndPurchased(db, accountId, originalPerson)
        for i := 0; i < len(products); i++ {
            prodMap[products[i].Pid] = products[i]
        }        

        rs := &RecoSet{products: prodMap, people: persMap,}
        genome := &Genome{rs: rs,  score: 0.0}
        genome.addRandomTrait()
        genomes[i] = genome
    }

    pop.genomes = genomes

    return
}

func reproduce(oneGenome *Genome, anotherGenome *Genome) (childGenome *Genome) {
    childGenome = &Genome{score: 0.0}
    childProducts := make(map[string]*database.Product)
    childPeople := make(map[string]*database.Person)
    childGenome.traits = make([]Trait, 0)

    for _, p := range oneGenome.getProducts() {
        coin := rand.Int31n(2)
        if coin == 1 {childProducts[p.Pid] = p}
    }
    for _, p := range anotherGenome.getProducts() {
        coin := rand.Int31n(2)
        if coin == 1 {childProducts[p.Pid] = p}
    }

    for _, p := range oneGenome.getPeople() {
        coin := rand.Int31n(2)
        if coin == 1 {childPeople[p.MonetateId] = p}
    }
    for _, p := range anotherGenome.getPeople() {
        coin := rand.Int31n(2)
        if coin == 1 {childPeople[p.MonetateId] = p}
    }

    childGenome.traits = append(childGenome.traits, &NopTrait{})

    childGenome.rs = &RecoSet{products: childProducts, people: childPeople,}

    return
}

