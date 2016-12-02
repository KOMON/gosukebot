package mtgstats

import (
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
)

func TestJoinAndWhere(t *testing.T) {
	fmt.Println(joinAndWhere(sq.Select("*").From("cards").Suffix("collate nocase"),
		Query{
			"name":      []string{"!Gleemax"},
			"color":     []string{"BR", "!U"},
			"supertype": []string{"legendary"},
			"set":       []string{"!UGL"},
		}).ToSql())
}

func TestSplitNegatives(t *testing.T) {
	fmt.Println(splitNegatives([]string{"!Gleemax"}))
}
