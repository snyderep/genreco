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
func (pop *Population) Display() {
    for i := 0; i < len(pop.genomes); i++ {
        genome := pop.genomes[i]
        fmt.Printf("genome %d, score: %f, people: %d, products: %d\n", i, genome.score, 
            genome.getPeopleCount(), genome.getProductsCount())
    }
}
func (pop *Population) GetHighestScoringGenome() (bestGenome *Genome) {
    for i := 0; i < len(pop.genomes); i++ {
        genome := pop.genomes[i]
        if bestGenome == nil || genome.score > bestGenome.score {
            bestGenome = genome
        }
    }
    return
}
func (pop *Population) DisplayFinal() {
    fmt.Println("********** DONE **********")

    bestGenome := pop.GetHighestScoringGenome()
    for _, product := range bestGenome.getProducts() {
        fmt.Println(product.String())
    }
    fmt.Printf("Score: %f\n", bestGenome.score)
    for i := 0; i < len(bestGenome.traits); i++ {
        fmt.Println(bestGenome.traits[i].String())
    }
}
func (pop *Population) Append(genomes []*Genome) {
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

    // adjust for the number of products 
    productCount := g.getProductsCount()    
    if productCount == 0 {
        countScore -= 10.0
    } else if productCount < 15 {
        // fmt.Println("+= 5.0 for product count between 0 and 14")        
        countScore += 5.0
    } else if productCount < 25 {
        // fmt.Println("+= 10.0 for product count between 15 and 24")
        countScore += 10.0
    } else {
        // fmt.Println("-1.0 for product count >= 25")        
        countScore += 3.0
    }

    // adjust for global conversion and seen/unseen
    for _, prod := range g.rs.products {
        conv := database.QueryGlobalConversion(db, accountId, prod)
        if conv == 0.0 {
            convScore += 0.0
        } else if conv <= 0.25 {
            convScore += 1.0 
        } else if conv <= 0.5 {
            convScore += 3.0
        } else if conv <= 0.75 {
            convScore += 4.0
        } else {
            convScore += 6.0
        }

        if database.HasProductBeenSeenByPerson(db, accountId, originalPerson, prod) {
            seenScore -= 0.5
        } else {
            seenScore += 1.0
        }
    }
    convScore = convScore / float64(g.getProductsCount())
    seenScore = seenScore / float64(g.getProductsCount())

    // adjust for seen/unseen 


    g.score = (countScore * 0.3) + (convScore * 0.5) + (seenScore * 0.2)
}
func (g *Genome) addRandomTrait() {
    n := rand.Intn(len(allTraits))
    t := allTraits[n]
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
            genome.checkFitness(db, accountId, originalPerson)
        }

        pop.Display()                

        if g == (maxGenerations - 1) {
            pop.DisplayFinal()
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
            pop.Append(childrenGenomes)
        }
    }
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

    for i := 0; i < len(oneGenome.traits); i++ {
        coin := rand.Int31n(2)
        if coin == 1 {childGenome.traits = append(childGenome.traits, oneGenome.traits[i])}
    }
    for i := 0; i < len(anotherGenome.traits); i++ {
        coin := rand.Int31n(2)
        if coin == 1 {childGenome.traits = append(childGenome.traits, anotherGenome.traits[i])}
    }
    // we have to have at least one trait
    if len(childGenome.traits) == 0 {
        childGenome.traits = append(childGenome.traits, &NopTrait{})
    }

    childGenome.rs = &RecoSet{products: childProducts, people: childPeople,}

    return
}

