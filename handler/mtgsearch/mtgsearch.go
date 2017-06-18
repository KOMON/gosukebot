package mtgsearch

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"unicode"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var matches []string

type mtgSearchResult struct {
	name string
	cost string
	text string
	id   int
	set  string
}

// MtgSearchHandler satisfies the handler.Handler interface
type MtgSearchHandler struct{}

// Returns a new MtgSearchHandler and initializes the package-level
// db connection
func New() MtgSearchHandler {
	var err error
	db, err = sql.Open("sqlite3", "mtg.db")
	if err != nil {
		log.Fatal(err)
	}

	return MtgSearchHandler{}
}

// Match searches a string for substrings [[inside double square brackets]]
// Returns true if any such substrings are found,  and stores the strings
// minus the brackets for the Response method
func (msh MtgSearchHandler) Match(msg string) bool {
	matches = []string{}
	// search for instances of opening brackets
	for i := strings.Index(msg, "[["); i != -1; i = strings.Index(msg, "[[") {
		// if there's a closing square bracket pair to match
		if j := strings.Index(msg, "]]"); j != -1 {
			// if the opening brackets aren't the start of the string and
			// the character before isn't whitespace
			if i != 0 && !unicode.IsSpace(rune(msg[i-1])) {
				//skip to just past the closing brackets
				msg = msg[j+2:]
				continue
			}
			matches = append(matches, msg[i+2:j])
			msg = msg[j+2:]
		} else {
			// skip to after the opening brackets to look for another pair
			msg = msg[i+2:]
		}
	}
	return len(matches) > 0
}

// Respond returns a string containing info about the card searched
// for, if no card is found it returns a message saying so
func (msh MtgSearchHandler) Respond() (string, error) {
	multi := len(matches) > 1
	response := ""

	for _, match := range matches {
		var (
			res mtgSearchResult
			err error
		)
		args := strings.Split(match, "|")

		if len(args) == 1 {
			res, err = runSearch(args[0], "")
		} else {
			res, err = runSearch(args[0], args[1])
		}

		if err != nil || res.name == "" {
			response += "Card Not Found!\n"
			log.Println(err)
			continue
		}
		if res.text == "" {
			res.text = " "
		}
		if multi {
			response += fmt.Sprintf("%s %s ```%s``` %s\n", res.name,
				res.cost, res.text, res.set)
		} else {
			response += fmt.Sprintf("%s %s ```%s``` %s", formatImageURL(res.id),
				res.cost, res.text, res.set)
		}
	}
	return response, nil
}

func runSearch(name string, set string) (mtgSearchResult, error) {
	var (
		rows *sql.Rows
		err  error
		id   []uint8
	)

	res := mtgSearchResult{}
	nameQuery := sq.
		Select("cards.id", "cards.name", "mana_cost", "card_text", "cards.multiverse_id").
		From("cards").
		Join("virt_cards on cards.id=virt_cards.id").
		Where("virt_cards.name match ? and cards.multiverse_id != 0", name)

	if strings.EqualFold(set, "ALL") {
		query := sq.Select("group_concat(n.set_code, ', ')").
			FromSelect(sq.
				Select("set_code").
				Options("distinct").
				FromSelect(nameQuery, "n").
				Join("set_card on n.id=set_card.id"), "n")
		err = query.RunWith(db).QueryRow().Scan(&res.set)
		if err != nil {
			return res, err
		}
	}
	if set != "" && !strings.EqualFold(set, "ALL") {
		query := sq.
			Select("n.id", "name", "mana_cost", "card_text", "multiverse_id").
			FromSelect(nameQuery, "n").
			Join("set_card on n.id=set_card.id").
			Where(sq.Eq{"set_code": strings.ToUpper(set)})
		rows, err = query.RunWith(db).Query()
	} else {
		rows, err = nameQuery.RunWith(db).Query()
	}
	if err != nil {
		log.Fatalf("Error in mtgsearch with query: %s, %s, %v\n", name, set, err)
	}
	defer rows.Close()

	if rows == nil || !rows.Next() {
		return res, err
	}

	err = rows.Scan(&id, &res.name, &res.cost, &res.text, &res.id)
	if err != nil {
		return mtgSearchResult{}, err
	}

	if !strings.EqualFold(name, res.name) {
		for rows.Next() {
			res := mtgSearchResult{}
			err := rows.Scan(&id, &res.name, &res.cost, &res.text, &res.id)
			if err != nil {
				log.Fatal(err)
			}
			if strings.EqualFold(res.name, name) {
				return res, err
			}
		}
	}
	if res.set == "" {
		res.set = set
	}
	return res, err
}

func formatImageURL(multiverseID int) string {
	return fmt.Sprintf("http://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=%d&type=card", multiverseID)
}
