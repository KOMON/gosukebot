package mtgstats

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	sq "github.com/Masterminds/squirrel"
)

func joinAndWhere(search sq.SelectBuilder, query Query) sq.SelectBuilder {
	for k, v := range query {
		eq, not := splitNegatives(v)
		switch k {
		case "verb", "verbs":
		case "names", "name":
			if len(eq) != 0 {
				search = search.Where(sq.Eq{"name": strMap(eq, strings.Title)})
			}
			if len(not) != 0 {
				search = search.Where(sq.NotEq{"name": strMap(not, strings.Title)})
			}
		case "colors", "color":
			search = search.Join("card_color on cards.id=card_color.id").
				Where(colorQuery(eq, false))
		case "colorIDs", "colorID":
			search = search.Join("card_colorID on cards.id=card_colorID.id").
				Where(colorQuery(eq, true))
		case "supertypes", "supertype":
			search = filterSupertype(search, v)
		case "types", "type":
			search = filterType(search, v)
		case "subtypes", "subtype":
			for _, term := range eq {
				search = search.Where("type like '%" + strings.Title(term) + "%'")
			}
			for _, term := range not {
				search = search.Where("type not like '%" + strings.Title(term) + "%'")
			}
		case "sets", "set", "set_codes", "set_code":
			search = search.Join("set_card on cards.id=set_card.id")
			if len(eq) != 0 {
				search = search.Where(sq.Eq{"set_code": strMap(eq, strings.ToUpper)})
			}
			if len(not) != 0 {
				search = search.Where(sq.NotEq{"set_code": strMap(not, strings.ToUpper)})
			}
		case "rarities", "rarity", "rareness":
			search = filterRarity(search, v)
		default:
		}
	}
	return search
}

func filterSupertype(search sq.SelectBuilder, ts []string) sq.SelectBuilder {
	eq, not := splitNegatives(ts)
	viewFilter := func(s string, eq bool) {
		view := ""
		switch strings.Title(s) {
		case "Legendary", "Legendaries", "Legend", "Legends":
			view = "legendaries"
		case "Basic", "Basics", "Basic Land", "Basic Lands":
			view = "basics"
		case "Ongoing", "Ongoings":
			view = "ongoings"
		case "Snow", "Snows":
			view = "snows"
		case "World", "Worlds":
			view = "worlds"
		default:
			return
		}
		if eq {
			search = search.Where("id in " + view)
		} else {
			search = search.Where("id not in " + view)
		}
	}
	for _, e := range eq {
		viewFilter(e, true)
	}
	for _, n := range not {
		viewFilter(n, false)
	}
	return search
}

func filterType(search sq.SelectBuilder, ts []string) sq.SelectBuilder {
	eq, not := splitNegatives(ts)
	viewFilter := func(s string, eq bool) {
		view := ""
		switch strings.Title(s) {
		case "Creature", "Creatures":
			view = "creatures"
		case "Artifact", "Artifacts":
			view = "artifacts"
		case "Enchantment", "Echantments":
			view = "enchantments"
		case "Land", "Lands":
			view = "lands"
		case "Planeswalker", "Planeswalkers":
			view = "planeswalkers"
		case "Instant", "Instants":
			view = "instants"
		case "Sorcery", "Sorceries":
			view = "sorceries"
		case "Tribal", "Tribals":
			view = "tribals"
		default:
			return
		}
		if eq {
			search = search.Where("id in " + view)
		} else {
			search = search.Where("id not in " + view)
		}
	}

	for _, e := range eq {
		viewFilter(e, true)
	}

	for _, n := range not {
		viewFilter(n, false)
	}
	return search
}

func filterRarity(search sq.SelectBuilder, rs []string) sq.SelectBuilder {
	eq, not := splitNegatives(rs)
	viewFilter := func(s string, eq bool) {
		view := ""
		switch strings.Title(s) {
		case "Common", "Commons", "C":
			view = "commons"
		case "Uncommon", "Uncommons", "U":
			view = "uncommons"
		case "Rare", "Rares", "R":
			view = "rares"
		case "Mythic", "Mythics", "Mythic Rare", "Mythic Rares", "MR":
			view = "mythics"
		case "Special", "Specials", "S":
			view = "specials"
		default:
			return
		}
		if eq {
			search = search.Where("id in " + view)
		} else {
			search = search.Where("id not in " + view)
		}
	}
	for _, e := range eq {
		viewFilter(e, true)
	}
	for _, n := range not {
		viewFilter(n, false)
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

func count(search sq.SelectBuilder, arg string) string {
	var res int
	err := queryOnSubSelect(sq.Select("count("+arg+")"), search).
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

func debugSqlizer(s sq.Sqlizer) string {
	sql, args, err := s.ToSql()
	if err != nil {
		return fmt.Sprintf("[ToSql error: %s]", err)
	}

	// TODO: dedupe this with placeholder.go
	buf := &bytes.Buffer{}
	i := 0
	for {
		p := strings.Index(sql, "?")
		if p == -1 {
			break
		}
		if len(sql[p:]) > 1 && sql[p:p+2] == "??" { // escape ?? => ?
			buf.WriteString(sql[:p])
			buf.WriteString("?")
			if len(sql[p:]) == 1 {
				break
			}
			sql = sql[p+2:]
		} else {
			if i+1 > len(args) {
				return fmt.Sprintf(
					"[DebugSqlizer error: too many placeholders in %#v for %d args]",
					sql, len(args))
			}
			buf.WriteString(sql[:p])
			fmt.Fprintf(buf, "'%v'", args[i])
			sql = sql[p+1:]
			i++
		}
	}
	if i < len(args) {
		return fmt.Sprintf(
			"[DebugSqlizer error: not enough placeholders in %#v for %d args]",
			sql, len(args))
	}
	buf.WriteString(sql)
	return buf.String()
}
