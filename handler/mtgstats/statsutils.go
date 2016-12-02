package mtgstats

import (
	"fmt"
	"unicode"

	sq "github.com/Masterminds/squirrel"
)

func joinAndWhere(search sq.SelectBuilder, query Query) sq.SelectBuilder {
	for k, v := range query {
		eq, not := splitNegatives(v)
		switch k {
		case "verb", "verbs":
			continue
		case "names", "name":
			break
		case "colors", "color":
			search = search.Join("card_color on cards.id=card_color.id").
				Where(colorQuery(eq, false))
			continue
		case "colorIDs", "colorID":
			search = search.Join("card_colorID on cards.id=card_colorID.id").
				Where(colorQuery(eq, true))
			continue
		case "supertypes", "supertype":
			search = search.Join("card_supertype on cards.id=card_supertype.id")
		case "types", "type":
			search = search.Join("card_type on cards.id=card_type.id")
		case "subtypes", "subtype":
			search = search.Join("card_subtype on cards.id=card_subtype.id")
		case "sets", "set", "set_codes", "set_code":
			search = search.Join("set_card on cards.id=set_card.id")
			if len(eq) != 0 {
				search = search.Where(sq.Eq{"set_code": eq})
			}
			if len(not) != 0 {
				search = search.Where(sq.NotEq{"set_code": not})
			}
			continue
		default:
			break
		}

		if len(eq) != 0 {
			search = search.Where(sq.Eq{k: eq})
		}
		if len(not) != 0 {
			search = search.Where(sq.NotEq{k: not})
		}
	}
	return search
}

func avg(search sq.SelectBuilder, arg string) string {
	var res float64
	err := queryOnSubSelect(sq.Select("avg("+arg+")"), search).
		QueryRow().
		Scan(&res)

	if err != nil {
		return err.Error() + "\n"
	}

	return fmt.Sprintf("Average %s: %f\n", arg, res)
}

func count(search sq.SelectBuilder) string {
	var res int
	err := queryOnSubSelect(sq.Select("count(id)"), search).
		QueryRow().
		Scan(&res)

	if err != nil {
		return err.Error() + "\n"
	}

	return fmt.Sprintf("Count: %d\n", res)
}

func sum(search sq.SelectBuilder, arg string) string {
	var res float64

	err := queryOnSubSelect(sq.Select("sum("+arg+")"), search).
		QueryRow().
		Scan(&res)

	if err != nil {
		return err.Error() + "\n"
	}

	return fmt.Sprintf("Sum %s: %f\n", arg, res)
}

func min(search sq.SelectBuilder, arg string) string {
	var (
		res  int
		id   string
		name string
	)
	err := queryOnSubSelect(sq.Select("min("+arg+")", "multiverse_id", "name"), search).
		QueryRow().
		Scan(&res, &id, &name)

	if err != nil {
		return err.Error() + "\n"
	}
	fmt.Println(search.ToSql())
	return fmt.Sprintf("Minimum %s: %s %s\n", arg, name, imageFromMID(id))
}

func max(search sq.SelectBuilder, arg string) string {
	var (
		res  int
		id   string
		name string
	)
	err := queryOnSubSelect(sq.Select("max("+arg+")", "multiverse_id", "name"), search).
		QueryRow().
		Scan(&res, &id, &name)

	if err != nil {
		return err.Error() + "\n"
	}

	return fmt.Sprintf("Maximum %s: %s %s\n", arg, name, imageFromMID(id))
}

func imageFromMID(id string) string {
	return fmt.Sprintf("http://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=%s&type=card", id)
}

func queryOnSubSelect(query sq.SelectBuilder, sub sq.SelectBuilder) sq.SelectBuilder {
	return query.FromSelect(sub.GroupBy("name"), "sub").
		Where("multiverse_id != 0").
		RunWith(db)
}

func colorQuery(colors []string, ID bool) string {
	table, result := "card_color", "1"
	if ID {
		table = "card_colorID"
	}

	for i, color := range colors {
		if i != 0 {
			result += "|1"
		}
		query := map[rune]bool{
			'w': false,
			'u': false,
			'b': false,
			'r': false,
			'g': false,
		}
		for _, rune := range color {
			query[unicode.ToLower(rune)] = true
		}

		for k, v := range query {
			if v {
				result += "&"
			} else {
				result += "&~"
			}
			result += table + "."
			if k != '0' {
				result += string(k)
			} else {
				result += "colorless"
			}
		}
	}
	return result
}

func strMap(ss []string, f func(string) string) []string {
	mapped := make([]string, len(ss))
	for i, s := range ss {
		mapped[i] = f(s)
	}

	return mapped
}

func strFilter(ss []string, f func(string) bool) []string {
	matches := []string{}
	for _, s := range ss {
		if f(s) {
			matches = append(matches, s)
		}
	}
	return matches
}

func splitNegatives(ss []string) ([]string, []string) {
	matches := strFilter(ss, func(s string) bool {
		return s[0] != '!'
	})

	non := strFilter(ss, func(s string) bool {
		return len(s) != 0 && s[0] == '!'
	})
	if len(non) != 0 {
		non = strMap(non, func(s string) string {
			return s[1:]
		})
	}
	return matches, non
}
