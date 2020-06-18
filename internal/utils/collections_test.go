package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapEquals(t *testing.T) {

	t.Run("two nil list should be equals", func(t *testing.T) {
		assert.True(t, MapEquals(nil, nil))
	})

	t.Run("two list with same content should be equals", func(t *testing.T) {

		left := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}

		right := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}
		assert.True(t, MapEquals(&left, &right))
	})

	t.Run("two list with same content with different order should be equals", func(t *testing.T) {

		left := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}

		right := map[string][]string{
			"groupA": []string{"a", "c", "b"},
		}
		assert.True(t, MapEquals(&left, &right))
	})

	t.Run("two list with different content not be equals", func(t *testing.T) {

		left := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}

		right := map[string][]string{
			"groupA": []string{"a", "b", "d"},
		}
		assert.False(t, MapEquals(&left, &right))
	})

	t.Run("two list with the first contained into the second should not be equals", func(t *testing.T) {

		left := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}

		right := map[string][]string{
			"groupA": []string{"a", "b", "c", "d"},
		}
		assert.False(t, MapEquals(&left, &right))
	})

	t.Run("two list with the second contained into the second should not be equals", func(t *testing.T) {

		left := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}

		right := map[string][]string{
			"groupA": []string{"a", "b", "c", "d"},
		}
		assert.False(t, MapEquals(&right, &left))
	})
}

func TestMapConcat(t *testing.T) {

	t.Run("two nil map should produce nil", func(t *testing.T) {
		assert.Nil(t, MapConcat(nil, nil))
	})

	t.Run("two map with same content should return the content of the first", func(t *testing.T) {

		left := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}

		right := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}
		assert.Equal(t, MapConcat(&left, &right), &left)
	})

	t.Run("two map with same key should return the union of values", func(t *testing.T) {

		left := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}

		right := map[string][]string{
			"groupA": []string{"d", "e", "f"},
		}
		concat := *MapConcat(&left, &right)
		list := concat["groupA"]
		assert.True(t, Equal(&[]string{"a", "b", "c", "d", "e", "f"}, &list))
	})

	t.Run("two map with same key should return the union of values", func(t *testing.T) {

		left := map[string][]string{
			"groupA": []string{"a", "b", "c"},
		}

		right := map[string][]string{
			"groupB": []string{"d", "e", "f"},
		}
		concat := *MapConcat(&left, &right)
		list := concat["groupB"]
		listA := concat["groupA"]
		assert.True(t, Equal(&[]string{"a", "b", "c"}, &listA))
		assert.True(t, Equal(&[]string{"d", "e", "f"}, &list))
	})

}
