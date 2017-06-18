package mtgstats

import (
	"database/sql"
	"log"
	"strings"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var matches []string

type MtgStatsHandler struct{}

type Query map[string][]string

func New() MtgStatsHandler {
	var err error
	db, err = sql.Open("sqlite3", "mtg.db")
	if err != nil {
		log.Fatal(err)
	}

	return MtgStatsHandler{}
}

func (msh MtgStatsHandler) Match(msg string) bool {
	matches = []string{}
	for i := strings.Index(msg, "#[["); i != -1; i = strings.Index(msg, "#[[") {
		if j := strings.Index(msg[i+3:], "]]"); j != -1 {
			matches = append(matches, msg[i+3:j+3])
			msg = msg[j+2:]
		} else {
			msg = msg[i+3:]
		}
	}
	return len(matches) > 0
}

func (msh MtgStatsHandler) Respond() (string, error) {
	response := ""
	for _, match := range matches {
		query := Query{}
		verbs := Query{}
		args := strings.Split(match, ",")
		for _, arg := range args {
			if len(arg) == 0 {
				break
			}
			kv := strings.Split(arg, ":")
			k, v := strings.TrimSpace(kv[0]), strMap(strings.Split(kv[1], "|"), strings.TrimSpace)
			if k == "avg" || k == "count" || k == "sum" ||
				k == "min" || k == "max" {
				verbs[k] = v
			} else {
				query[k] = v
			}
		}
		if len(verbs) == 0 {
			verbs["count"] = []string{"id"}
		}
		response += runSearch(query, verbs)
	}

	return response, nil
}

//check statsutils.go for the definitions of unfamiliar functions used here
func runSearch(query Query, verbs Query) string {
	response := ""
	search := joinAndWhere(sq.Select("*").From("cards").Suffix("collate nocase"), query)

	for verb, args := range verbs {
		for _, arg := range args {
			switch verb {
			case "avg":
				response += avg(search, arg)
			case "count":
				response += count(search, arg)
			case "sum":
				response += sum(search, arg)
			case "min":
				response += min(search, arg)
			case "max":
				response += max(search, arg)
			default:
				continue
			}
		}
	}
	return response
}
