package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
)

var tables []string = []string{
	`create table sets (
     name varchar(50),
     code varchar(4),
     release_date varchar(10),
     type varchar(10),
     block varchar(50)
   )`,
	`create table cards (
     id varchar(40),
     layout varchar(10),
     name varchar(255),
     mana_cost varchar(10),
     cmc integer,
     type varchar(100),
     card_text text,
     flavor text,
     artist varchar(50),
     number varchar(20),
     power integer,
     toughness integer,
     loyalty integer,
     multiverse_id integer,
     timeshifted boolean,
     reserved boolean,
     release_date varchar(10),
     mci_number varchar(4)
   )`,
	`create table set_card (set_code varchar(4), id varchar(40))`,
	`create table card_color (
     id varchar(40), r boolean, g boolean, 
     u boolean, b boolean, w boolean, colorless boolean
   )`,
	`create table card_colorID (
     id varchar(40), r boolean, g boolean, 
     u boolean, b boolean, w boolean, colorless boolean
   )`,
	`create table card_supertype(id varchar(40), supertype varchar(10))`,
	`create table card_type(id varchar(40), type varchar(20))`,
	`create table card_rarity(id varchar(40), rarity varchar(12))`,
	`create virtual table virt_cards using fts3(id, name, multiverse_id)`,
}

var views []string = []string{
	`create view creatures (id) as select id from card_type where type = "Creature"`,
	`create view artifacts (id) as select id from card_type where type = "Artifact"`,
	`create view enchantments (id) as select id from card_type where type = "Enchantment"`,
	`create view lands (id) as select id from card_type where type = "Land"`,
	`create view planeswalkers (id) as select id from card_type where type = "Planeswalker"`,
	`create view instants (id) as select id from card_type where type = "Instant"`,
	`create view sorceries (id) as select id from card_type where type = "Sorcery"`,
	`create view tribals (id) as select id from card_type where type = "Tribal"`,

	`create view legendaries (id) as select id from card_supertype where supertype = "Legendary"`,
	`create view basics (id) as select id from card_supertype where supertype = "Basic"`,
	`create view ongoings (id) as select id from card_supertype where supertype = "Ongoing"`,
	`create view snows (id) as select id from card_supertype where supertype = "Snow"`,
	`create view worlds (id) as select id from card_supertype where supertype = "World"`,

	`create view commons (id) as select id from card_rarity where rarity = "Common"`,
	`create view uncommons (id) as select id from card_rarity where rarity = "Uncommon"`,
	`create view rares (id) as select id from card_rarity where rarity = "Rare"`,
	`create view mythics (id) as select id from card_rarity where rarity = "Mythic Rare"`,
	`create view specials (id) as select id from card_rarity where rarity = "Special"`,
}

var db *sql.DB

func main() {
	os.Remove("./mtg.db")

	db, _ = sql.Open("sqlite3", "./mtg.db")

	defer db.Close()

	for _, stmt := range tables {
		_, _ = db.Exec(stmt)
	}

	sets, _ := loadSets("AllSets.json")

	for _, s := range sets {
		fmt.Printf("\033[K Importing set: %s\r", s.Name)
		ImportSet(s)
	}

	for _, stmt := range views {
		_, _ = db.Exec(stmt)
	}
	_, _ = db.Exec("insert into virt_cards select id, name, multiverse_id from cards")
}

func ImportSet(s Set) {
	insertSet := sq.
		Insert("sets").
		Columns("name", "code", "release_date", "type", "block")

	_, _ = insertSet.
		Values(s.Name, s.Code, s.ReleaseDate, s.Type, s.Block).
		RunWith(db).Exec()

	for _, c := range s.Cards {
		if s.Type == "promo" {
			continue
		}
		ImportCard(c, s.ReleaseDate)
		ImportSetCard(s, c)
		ImportCardColor(c)
		ImportCardColorID(c)
		ImportCardSupertype(c)
		ImportCardType(c)
		ImportCardRarity(c)
	}
}

func ImportCard(c Card, releaseDate string) {
	p, _ := strconv.ParseInt(c.Power, 0, 0)
	t, _ := strconv.ParseInt(c.Toughness, 0, 0)
	cost := formatCost(c.ManaCost)
	_, _ = sq.
		Insert("cards").
		Columns("id", "name", "mana_cost", "cmc",
			"type", "card_text", "flavor", "artist",
			"number", "power", "toughness", "loyalty",
			"multiverse_id", "timeshifted", "reserved",
			"release_date", "mci_number").
		Values(c.ID, c.Name, cost, c.CMC, c.Type,
			c.Text, c.Flavor, c.Artist, c.Number, p, t, c.Loyalty,
			c.MultiverseID, c.Timeshifted, c.Reserved, releaseDate,
			c.MCINumber).
		RunWith(db).Exec()
}

func ImportSetCard(s Set, c Card) {
	_, _ = sq.
		Insert("set_card").
		Columns("set_code", "id").
		Values(s.Code, c.ID).
		RunWith(db).Exec()
}

func ImportCardColor(c Card) {
	r, g, u, b, w := false, false, false, false, false
	colorless := true

	if c.Colors != nil {
		colorless = false
		for _, color := range c.Colors {
			switch color {
			case "Red":
				r = true
			case "Blue":
				u = true
			case "Green":
				g = true
			case "Black":
				b = true
			case "White":
				w = true
			}
		}
	}

	_, _ = sq.
		Insert("card_color").
		Columns("id", "r", "g", "u", "b", "w", "colorless").
		Values(c.ID, r, g, u, b, w, colorless).
		RunWith(db).Exec()
}

func ImportCardColorID(c Card) {
	r, g, u, b, w := false, false, false, false, false
	colorless := true

	if c.ColorIdentity != nil {
		colorless = false
		for _, color := range c.ColorIdentity {
			switch color {
			case "Red":
				r = true
			case "Blue":
				u = true
			case "Green":
				g = true
			case "Black":
				b = true
			case "White":
				w = true
			}
		}
	}

	_, _ = sq.
		Insert("card_colorID").
		Columns("id", "r", "g", "u", "b", "w", "colorless").
		Values(c.ID, r, u, g, b, w, colorless).
		RunWith(db).Exec()
}

func ImportCardSupertype(c Card) {
	insertSupertype := sq.
		Insert("card_supertype").
		Columns("id", "supertype")
	for _, s := range c.Supertypes {
		_, _ = insertSupertype.
			Values(c.ID, s).
			RunWith(db).Exec()
	}
}

func ImportCardType(c Card) {
	insertType := sq.
		Insert("card_type").
		Columns("id", "type")
	for _, t := range c.Types {
		_, _ = insertType.
			Values(c.ID, t).
			RunWith(db).Exec()
	}
}

func ImportCardRarity(c Card) {
	insertRarity := sq.
		Insert("card_rarity").
		Columns("id", "rarity")
	for _, r := range c.Rarity.Rarities {
		_, _ = insertRarity.
			Values(c.ID, r).
			RunWith(db).Exec()
	}
}

func importVirtCards() {
	sq.
		Insert("virt_cards").
		Columns("id", "name", "multiverse_id")
}

func formatCost(cost string) string {
	fixSymbols := strings.NewReplacer(
		"{W}", ":ww:", "{U}", ":uu:", "{B}", ":bb:",
		"{R}", ":rr:", "{G}", ":gg:", "{1}", ":1:",
		"{2}", ":2:", "{3}", ":3:", "{4}", ":4:",
		"{5}", ":5:", "{6}", ":6:", "{7}", ":7:",
		"{8}", ":8:", "{9}", ":9:", "{10}", ":10:",
		"{11}", ":11:", "{12}", ":12:", "{13}", ":13:",
		"{14}", ":14:", "{15}", ":15:", "{20}", ":20:",
		"{X}", ":xx:", "{2w}", ":2w:", "{2u}", ":2u:",
		"{2b}", ":2b:", "{2r}", ":2r:", "{2g}", ":2g:",
		"{W/P}", ":wp:", "{U/P}", ":up:", "{B/P}", ":bp:",
		"{R/P}", ":rp:", "{G/P}", ":gp:", "{W/U}", ":wu:",
		"{W/B}", ":wb:", "{U/B}", ":ub:", "{U/R}", ":ur:",
		"{B/R}", ":br:", "{B/G}", ":bg:", "{R/G}", ":rg:",
		"{R/W}", ":rw:", "{G/W}", ":gw:", "{G/U}", ":gu:",
	)
	return fixSymbols.Replace(cost)
}
