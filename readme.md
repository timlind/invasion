# Simulation

This is a coding exercise which simulates an alien invasion of a given number of aliens on a world of cities specified by an input file.

## Execution

The executable takes two arguments, the world file, and a uint64 specifying the number of aliens to create.
```sh
go run main.go simulation/testdata/world.txt 2
```

## Testing

```bash
go test ./... 
```

## Map File

A map file has the format with one city per line. The first word is the city name, followed by 1-4 directions (north,south,west,east).
The city and each of the pairs are seperated by a single space, and the directions are seperated from their respective cities with an equal sign.

```
Foo north=Bar west=Baz south=Qu-ux
Bar south=Foo
Baz east=Foo
Qu-ux north=Foo
```

## API

ParseWorld() will parse in the given file with the city map, and the given number of aliens.
```go
ParseWorld(filename string, numAliens uint64) (*World, error)
```

### World

StartWar() will send the aliens to their initial cities then start a goroutine per alien to continue travelling and fighting as they go.
```go
StartWar()
```

String() will output the world map in the same format as input
```go
String() string
```

### Alien

Invade() will make a given number of movements via the movements channel and return on the done channel when done.
```go
Invade(moves int, movements chan Hovercraft, done chan bool)
```

Cancel() will cancel the invasion this alien is making.
```go
Cancel()
```