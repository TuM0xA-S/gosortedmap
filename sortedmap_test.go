package gosortedmap

import (
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createIntMap() *SortedMap {
	return NewSortedMap(func(a, b interface{}) int {
		return a.(int) - b.(int)
	})
}

func TestDeleteNotExistingElement(t *testing.T) {
	req := require.New(t)

	sm := createIntMap()
	req.Equal(0, sm.Len())
	sm.Delete(1)
	req.Equal(0, sm.Len())
}

func TestDeleteExistingElement(t *testing.T) {
	req := require.New(t)
	key := 22
	sm := createIntMap()
	sm.Set(key, "hello")
	sm.Delete(key)
	req.Equal(0, sm.Len())
	val, ok := sm.Get(key)
	req.Nil(val)
	req.False(ok)
}

func TestSetElement(t *testing.T) {
	req := require.New(t)
	sm := createIntMap()
	key := 1
	value := "hello world"
	sm.Set(key, value)
	req.Equal(1, sm.Len())
	actual, ok := sm.Get(key)
	req.Equal(value, actual)
	req.True(ok)
}

func TestUpdateElement(t *testing.T) {
	req := require.New(t)
	sm := createIntMap()
	key := 1
	value := "updated"
	sm.Set(key, "initial")
	sm.Set(key, value)
	actual, ok := sm.Get(key)
	req.Equal(value, actual)
	req.True(ok)
}

func TestGetElement(t *testing.T) {
	ass := assert.New(t)
	sm := createIntMap()

	cnt := 16
	for i := 0; i < cnt; i++ {
		sm.Set(i, i)
	}
	ass.Equal(cnt, sm.Len())
	for i := 0; i < cnt; i++ {
		value, ok := sm.Get(i)
		ass.True(ok)
		ass.Equal(i, value)
	}
}

func TestCreateBunchAndDeleteHalf(t *testing.T) {
	ass := assert.New(t)
	req := require.New(t)
	sm := createIntMap()

	cnt := 16
	for i := 0; i < cnt; i++ {
		sm.Set(i, i)
	}

	for i := 0; i < cnt/2; i++ {
		sm.Delete(i * 2) // delete every second
	}

	req.Equal(sm.Len(), cnt-cnt/2)

	for i := 0; i < cnt/2; i++ {
		value, ok := sm.Get(2 * i)
		ass.False(ok)
		ass.Nil(value)
	}

}

func TestAsSlice(t *testing.T) {
	req := require.New(t)
	sm := createIntMap()

	data := []int{
		10, 1, 2, 7, 5, 12, 11, 15, 4,
	}

	for _, val := range data {
		sm.Set(val, val*2)
	}
	entries := sm.AsSlice()

	actual := [][2]int{}
	for _, ent := range entries {
		actual = append(actual, [2]int{ent.Key.(int), ent.Value.(int)})
	}

	sort.Ints(data)
	expected := [][2]int{}
	for _, ent := range data {
		expected = append(expected, [2]int{ent, ent * 2})
	}

	req.Equal(expected, actual)
}

func TestAsChan(t *testing.T) {
	req := require.New(t)
	sm := createIntMap()

	data := []int{
		10, 1, 2, 7, 5, 12, 11, 15, 4,
	}

	for _, val := range data {
		sm.Set(val, val*2)
	}
	entries := sm.AsChan()

	actual := [][2]int{}
	for ent := range entries {
		actual = append(actual, [2]int{ent.Key.(int), ent.Value.(int)})
	}

	sort.Ints(data)
	expected := [][2]int{}
	for _, ent := range data {
		expected = append(expected, [2]int{ent, ent * 2})
	}

	req.Equal(expected, actual)
}

// do operationCount operations with elements in [0; upperBound)
// operations: get, set, delete, check sorted
// use default seed, so fail can be restored
func TestRandomOperaions(t *testing.T) {
	req := require.New(t)
	operationCount := 10000
	upperBound := 512
	sm := createIntMap()
	hm := map[int]interface{}{}
	for operationCount > 0 {
		operationCount--

		switch rand.Intn(4) {
		case 0:
			key := rand.Intn(upperBound)
			valAct, okAct := sm.Get(key)
			valExp, okExp := hm[key]

			req.Equal(okExp, okAct)
			req.Equal(valExp, valAct)
		case 1:
			key := rand.Intn(upperBound)
			value := rand.Intn(upperBound)
			sm.Set(key, value)
			hm[key] = value
		case 2:
			key := rand.Intn(upperBound)
			sm.Delete(key)
			delete(hm, key)
		case 3:
			expected := [][2]int{}
			for k, v := range hm {
				expected = append(expected, [2]int{k, v.(int)})
			}
			sort.Slice(expected, func(i, j int) bool {
				return expected[i][0] < expected[j][0]
			})
			actual := [][2]int{}
			for _, e := range sm.AsSlice() {
				actual = append(actual, [2]int{e.Key.(int), e.Value.(int)})
			}

			req.Equal(expected, actual)

		}
	}

}

type comparableString string

func (cs comparableString) CompareTo(another Comparable) int {
	anothercs := another.(comparableString)
	return strings.Compare(string(cs), string(anothercs))
}

func TestComparableInterface(t *testing.T) {
	req := require.New(t)
	sm := NewSortedMap(nil)

	sm.Set(comparableString("world"), 2)
	sm.Set(comparableString("hello"), 1)
	sm.Set(comparableString("!!!"), 3)
	req.Equal([]Entry{{comparableString("!!!"), 3}, {comparableString("hello"), 1}, {comparableString("world"), 2}}, sm.AsSlice())

	val, ok := sm.Get(comparableString("world"))
	req.True(ok)
	req.Equal(2, val)

	sm.Delete(comparableString("!!!"))

	val, ok = sm.Get(comparableString("!!!"))
	req.False(ok)
	req.Equal(2, sm.Len())
}
