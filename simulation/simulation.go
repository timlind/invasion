package v2

import (
	"bufio"
	"context"
	"math/rand"
	"os"
	"strings"
)

type World struct {
	activeAliens uint64
	aliens       map[uint64]*Alien
	cities       map[string]*City
}

type City struct {
	name  string
	roads map[string]string
	alien *Alien
}

type Alien struct {
	id     uint64
	ctx    context.Context
	cancel context.CancelFunc
	city   *City
}

type Hovercraft struct {
	alienId     uint64
	direction   int
	initialCity string
}

type Direction int

const (
	NORTH = Direction(0)
	SOUTH = Direction(1)
	WEST  = Direction(2)
	EAST  = Direction(3)
)

func (d Direction) String() string {
	switch d {
	case NORTH:
		return "north"
	case SOUTH:
		return "south"
	case WEST:
		return "west"
	case EAST:
		return "east"
	}

	return ""
}

func NewAlien(id uint64) *Alien {
	ctx, cancel := context.WithCancel(context.Background())
	return &Alien{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		city:   nil,
	}
}

func (alien *Alien) Cancel() {
	alien.cancel()
}

func (world *World) StartWar() {
	// start occupying initial cities
	for _, alien := range world.aliens {
		city := world.chooseCity()
		if city == "" {
			break
		}
		world.occupy(Hovercraft{
			alienId:     alien.id,
			initialCity: city,
		})
	}

	if len(world.cities) == 0 {
		return
	}

	// start the invasion
	activeAliens := len(world.aliens)
	movements := make(chan Hovercraft)
	done := make(chan bool)
	for _, alien := range world.aliens {
		go alien.Invade(10000, movements, done)
	}

	// spin the world,
	// processing the fighting and travelling
	for activeAliens > 0 {
		select {
		case hovercraft := <-movements:
			world.spin(hovercraft)
		case <-done:
			activeAliens--
		}
	}
}

func (world *World) chooseCity() string {
	if len(world.cities) == 0 {
		return ""
	}

	var cityNames []string
	for cityName := range world.cities {
		cityNames = append(cityNames, cityName)
	}
	cityPosition := rand.Intn(len(cityNames))
	return cityNames[cityPosition]
}

func (alien Alien) Invade(moves int, movements chan Hovercraft, done chan bool) {
	for move := 0; move < moves; move++ {
		select {
		case <-alien.ctx.Done():
			done <- true
			return
		default:
			// choose a direction
			movements <- Hovercraft{
				alienId:   alien.id,
				direction: rand.Intn(4),
			}
		}
	}

	done <- true
}

// occupy() will make an alien's initial move into a city,
// if there is already an alien there they fight and the city and aliens are destroyed,
// otherwise the alien occupies the city exclusively.
func (world *World) occupy(hovercraft Hovercraft) {
	city := world.cities[hovercraft.initialCity]
	if city.alien != nil {
		// destroy the city and both aliens
		world.fight(hovercraft.initialCity, world.aliens[hovercraft.alienId])
	} else {
		alien := world.aliens[hovercraft.alienId]
		alien.city = world.cities[hovercraft.initialCity]
		world.cities[hovercraft.initialCity].alien = alien
	}
}

// fight() will trigger fighting in city between the given alien and the alien currently occupying it.
// The aliens will be cancelled, deleted, and the city destroyed along with all roads to it in neighbouring cities.
func (world *World) fight(city string, alien *Alien) {
	if world.cities[city].alien != nil {
		// cancel both aliens
		alien.Cancel()
		world.cities[city].alien.Cancel()
		delete(world.aliens, alien.id)
		delete(world.aliens, world.cities[city].alien.id)
		// destroy the city and all roads
		for direction := range []int{0, 1, 2, 3} {
			dest, ok := world.cities[city].roads[Direction(direction).String()]
			if ok {
				for d, to := range world.cities[dest].roads {
					if to == city {
						delete(world.cities[dest].roads, d)
					}
				}
			}
		}
		delete(world.cities, city)
	}
}

// spin() will process the movement of an alien from one city to another.
// If the alien is trying to move a direction there is no road it will forfeit this move.
func (world *World) spin(hovercraft Hovercraft) {
	alienId := hovercraft.alienId
	alien, ok := world.aliens[alienId]
	if !ok {
		return
	}
	direction := hovercraft.direction
	from := alien.city

	// cancel trapped alien if there are no roads
	if from != nil && len(from.roads) == 0 {
		alien.Cancel()
		return
	}

	// forfeit this move if there is no road in the chosen direction.
	to, hasRoad := world.cities[from.name].roads[Direction(direction).String()]
	if !hasRoad {
		return
	}

	from.alien = nil

	if world.cities[to].alien != nil {
		// destroy destination
		world.fight(to, alien)
	} else {
		// move alien to destination
		alien.city = world.cities[to]
		world.cities[to].alien = alien
	}
}

func (world *World) String() string {
	output := ""
	for c, city := range world.cities {
		output += c + " "
		for direction, toCity := range city.roads {
			output += direction + "=" + toCity + " "
		}
		output += "\n"
	}
	return output
}

func ParseWorld(filename string, numAliens uint64) (*World, error) {
	aliens := make(map[uint64]*Alien, numAliens)
	for a := uint64(0); a < numAliens; a++ {
		aliens[a] = NewAlien(a)
	}

	world := &World{
		aliens: aliens,
		cities: make(map[string]*City),
	}

	file, _ := os.Open(filename)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		cityName := parts[0]

		city := &City{
			name:  cityName,
			roads: make(map[string]string),
		}

		for a := 1; a < len(parts); a++ {
			road := parts[a]
			link := strings.Split(road, "=")
			direction := link[0]
			to := link[1]
			city.roads[direction] = to
		}

		world.cities[cityName] = city
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return world, nil
}
