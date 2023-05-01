package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_StartWar(t *testing.T) {
	tt := []struct {
		name              string
		filename          string
		numAliens         uint64
		numExpectedCities int
	}{
		{"trapped", "testdata/world.txt", 1, 4},
		{"one fight", "testdata/world.txt", 2, 3},
		{"apocolypse", "testdata/world.txt", 10, 0},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			world, err := ParseWorld(tc.filename, tc.numAliens)
			require.NoError(t, err)
			world.StartWar()
			require.Equal(t, tc.numExpectedCities, len(world.cities))
		})
	}
}

func Test_ParseWorld(t *testing.T) {
	t.Run("basic example", func(t *testing.T) {
		world, err := ParseWorld("testdata/world.txt", 1)
		require.NoError(t, err)
		require.Len(t, world.cities, 4)
		// Foo
		require.Equal(t, world.cities["Foo"].name, "Foo")
		// Foo paths
		require.Len(t, world.cities["Foo"].roads, 3)
		require.Equal(t, world.cities["Foo"].roads["north"], "Bar")
		require.Equal(t, world.cities["Foo"].roads["west"], "Baz")
		require.Equal(t, world.cities["Foo"].roads["south"], "Qu-ux")
		// Bar
		require.Equal(t, world.cities["Bar"].name, "Bar")
		require.Equal(t, world.cities["Bar"].roads["south"], "Foo")
		// Baz
		require.Equal(t, world.cities["Baz"].name, "Baz")
		require.Equal(t, world.cities["Baz"].roads["east"], "Foo")
		// Qu-ux
		require.Equal(t, world.cities["Qu-ux"].name, "Qu-ux")
		require.Equal(t, world.cities["Qu-ux"].roads["north"], "Foo")
	})
}

func TestDirection_String(t *testing.T) {
	tests := []struct {
		name string
		d    Direction
		want string
	}{
		{"north", NORTH, "north"},
		{"south", SOUTH, "south"},
		{"west", WEST, "west"},
		{"east", EAST, "east"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorld_occupy(t *testing.T) {
	t.Run("one alien", func(t *testing.T) {
		world, err := ParseWorld("testdata/world.txt", 1)
		require.NoError(t, err)
		world.occupy(Hovercraft{alienId: 0, initialCity: "Foo"})
		require.Equal(t, uint64(0), world.cities["Foo"].alien.id)
	})

	t.Run("fighting", func(t *testing.T) {
		world, err := ParseWorld("testdata/world.txt", 2)
		require.NoError(t, err)
		world.occupy(Hovercraft{alienId: 0, initialCity: "Foo"})
		require.Equal(t, uint64(0), world.cities["Foo"].alien.id)
		world.occupy(Hovercraft{alienId: 1, initialCity: "Foo"})
		require.Nil(t, world.cities["Foo"])
	})
}

func TestWorld_fight(t *testing.T) {
	world, err := ParseWorld("testdata/world.txt", 2)
	require.NoError(t, err)
	world.occupy(Hovercraft{alienId: 0, initialCity: "Foo"})
	world.fight("Foo", world.aliens[1])
	require.Equal(t, 0, len(world.aliens))
	require.Equal(t, 3, len(world.cities))
	require.Nil(t, world.cities["Foo"])
	require.Empty(t, world.cities["Bar"].roads["south"])
}

func TestWorld_spin(t *testing.T) {
	t.Run("move alien", func(t *testing.T) {
		world, err := ParseWorld("testdata/world.txt", 1)
		require.NoError(t, err)
		world.occupy(Hovercraft{alienId: 0, initialCity: "Foo"})
		world.spin(Hovercraft{alienId: 0, direction: int(NORTH)})
		require.Equal(t, uint64(0), world.cities["Bar"].alien.id)
		require.Equal(t, world.cities["Bar"], world.aliens[0].city)
	})

	t.Run("destroy destination", func(t *testing.T) {
		world, err := ParseWorld("testdata/world.txt", 2)
		require.NoError(t, err)
		world.occupy(Hovercraft{alienId: 0, initialCity: "Foo"})
		world.occupy(Hovercraft{alienId: 1, initialCity: "Bar"})
		world.spin(Hovercraft{alienId: 0, direction: int(NORTH)})
		require.Nil(t, world.cities["Bar"])
		require.Equal(t, 0, len(world.aliens))
	})
}

func TestAlien_Invade(t *testing.T) {
	t.Run("move twice", func(t *testing.T) {
		world, err := ParseWorld("testdata/world.txt", 1)
		require.NoError(t, err)

		movements := make(chan Hovercraft)
		done := make(chan bool)

		go world.aliens[0].Invade(2, movements, done)

		move := <-movements
		require.Equal(t, uint64(0), move.alienId)
		move = <-movements
		require.Equal(t, uint64(0), move.alienId)
		now := <-done
		require.True(t, now)
	})

	t.Run("cancel", func(t *testing.T) {
		world, err := ParseWorld("testdata/world.txt", 1)
		require.NoError(t, err)

		movements := make(chan Hovercraft)
		done := make(chan bool)
		alien := world.aliens[0]

		alien.Cancel()
		go alien.Invade(2, movements, done)

		cancelled := <-done
		require.True(t, cancelled)
	})
}

func TestWorld_String(t *testing.T) {
	expected := "Foo north=Bar west=Baz south=Qu-ux \nBar south=Foo \nBaz east=Foo \nQu-ux north=Foo \n"
	world, err := ParseWorld("testdata/world.txt", 1)
	require.NoError(t, err)
	require.Equal(t, expected, world.String())
}

func TestWorld_chooseCity(t *testing.T) {
	world, err := ParseWorld("testdata/world.txt", 1)
	require.NoError(t, err)
	name := world.chooseCity()
	_, ok := world.cities[name]
	require.True(t, ok)
}
