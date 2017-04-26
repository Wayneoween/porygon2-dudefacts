package bot

import (
	"encoding/json"
	"fmt"
	"github.com/0x263b/porygon2"
	"github.com/oxtoacart/go-ringbuffer/ringbuff"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
)

var dudefacts []DudeFact
var dudefactsloaded bool = false
var context = ringbuff.New(10)

type DudeFact struct {
	Nickname string `json:"nickname"`
	Facts    []Fact `json:"facts"`
}

type Fact struct {
	Content string            `json:"fact"`
	Votes   map[string]string `json:"votes"`
	Context []string          `json:"context,omitempty"`
	Meta    metaData          `json:"meta,omitempty"`
}

type metaData struct {
	Author string `json:"author"`
	Added  int    `json:"added"`
}

func addToFactContext(command *bot.PassiveCmd) (string, error) {
	matched, _ := regexp.MatchString(`^\.`, command.Raw)
	if !matched {
		line := "<" + command.Nick + "> " + command.Raw
		context.Add(line)
	}
	return "", nil
}

func loadFacts() {
	if dudefactsloaded == false {
		dudefactsfile, e := ioutil.ReadFile(".dudefacts.json")
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		json.Unmarshal(dudefactsfile, &dudefacts)
		dudefactsloaded = true
	}
	return
}

func writeFacts() {
	path := ".dudefacts.json"
	fo, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	e := json.NewEncoder(fo)
	if err := e.Encode(dudefacts); err != nil {
		panic(err)
	}
}

func buildLineItem(i int, dudefact Fact) string {
	var fct string = dudefact.Content
	var item string
	var idx int

	idx = i + 1

	item = fmt.Sprintf("[%02d", idx)

	if len(dudefact.Context) != 0 {
		item += "*"
	}

	item += "] " + fct

	return item
}

func voteFact(command *bot.Cmd, matches []string) (msg string, err error) {
	if dudefactsloaded == false {
		loadFacts()
	}

	votee := command.Nick
	voted_nick := matches[1]
	voted_fact, _ := strconv.Atoi(matches[2])
	vote_value, vote_err := strconv.ParseBool(matches[3])

	var output string

	if command.Nick == voted_nick {
		output = "du kek"
	} else if voted_fact <= 0 {
		output = "top kek kleiner null"
	} else if vote_err != nil {
		output = "wat"
	} else {
		voted_fact -= 1

		for idx, dude := range dudefacts {
			if voted_nick == dude.Nickname {
				if voted_fact <= len(dude.Facts) {
					dude.Facts[voted_fact].Votes[votee] = strconv.FormatBool(vote_value)
					output = "kek, ok\n"

					false_counter := 0
					for _, vote := range dude.Facts[voted_fact].Votes {
						// count false votes
						if vote == "false" {
							false_counter++
						}
					}
					// if 3 false counts on a fact
					if false_counter >= 3 {
						// remove fact
						dudefacts = append(dudefacts[:idx], dudefacts[idx+1:]...)
						output += "i bims der loescher, fact war zu kek"
					}
				}
			}
		}
	}

	writeFacts()

	return output, nil

}

func addFact(command *bot.Cmd, matches []string) (msg string, err error) {
	if dudefactsloaded == false {
		loadFacts()
	}

	input_nick := matches[1]
	input_fact := matches[2]
	author := command.Nick

	var dfact Fact
	var dfact_meta metaData

	var timestamp int
	var output string

	now := time.Now()
	timestamp = int(now.Unix())

	if command.Nick == input_nick {
		return "fick dich weg, du hurrens0hn", nil
	}

	// iterate over all dudes
	for idx, dude := range dudefacts {
		// if dude matches nickname from irc
		if input_nick == dude.Nickname {
			// create metadata
			dfact_meta.Added = timestamp
			dfact_meta.Author = author

			// create dudefact
			dfact.Content = input_fact
			dfact.Votes = make(map[string]string)
			dfact.Context = []string{}
			context.ForEach(func(line interface{}) {
				dfact.Context = append(dfact.Context, line.(string))
			})
			dfact.Meta = dfact_meta

			dudefacts[idx].Facts = append(dudefacts[idx].Facts, dfact)
			output = "Ok du kek"
		}
	}

	writeFacts()

	return output, nil
}

func printUserFactContext(command *bot.Cmd, matches []string) (msg string, err error) {
	if dudefactsloaded == false {
		loadFacts()
	}

	input_nick := matches[1]
	input_fact_num, err := strconv.Atoi(matches[2])
	fct_num := input_fact_num - 1
	var output string

	// iterate over all dudes
	for _, dude := range dudefacts {
		// if dude matches nickname from irc
		if input_nick == dude.Nickname {
			// and context array not empty
			if len(dude.Facts[fct_num].Context) == 0 {
				output = "k1 kontext fuer 1 kek wie dich"
			} else {
				// collect each context line for this fact
				for _, ctx := range dude.Facts[fct_num].Context {
					output += ctx + "\n"
				}
			}
			break
		}
	}

	return output, nil
}

func printAllUserFacts(command *bot.Cmd, matches []string) (msg string, err error) {
	if dudefactsloaded == false {
		loadFacts()
	}

	input_nick := matches[1]
	var output string

	// iterate over all dudes
	for _, dude := range dudefacts {
		// if dude matches nickname from irc
		if input_nick == dude.Nickname {
			// get the facts of this user
			var facts = dude.Facts
			// and concatenate all facts
			iter := 0
			line := ""
			for i, fact := range facts {
				if iter == 0 {
					line = buildLineItem(i, fact)
					iter++
				} else if iter < 3 {
					line += "; " + buildLineItem(i, fact)
					iter++
				}
				if iter == 3 {
					output += line + "\n"
					line = ""
					iter = 0
				}
			}
			// clean possibly remaining items
			if iter >= 1 {
				output += line + "\n"
				line = ""
				iter = 0
			}
		}
	}

	return output, nil
}

func printRandomFactUser(command *bot.Cmd, matches []string) (msg string, err error) {
	if dudefactsloaded == false {
		loadFacts()
	}

	input_nick := matches[1]
	var output string

	for _, dude := range dudefacts {
		if input_nick == dude.Nickname {
			rf := rand.Int() % len(dude.Facts)
			output = buildLineItem(rf, dude.Facts[rf])
		} else {
			output = "ENOSUCHKEK: [EE] babetibupi kaputti"
		}
	}

	return output, nil
}

func printRandomFact(command *bot.Cmd, matches []string) (msg string, err error) {
	if dudefactsloaded == false {
		loadFacts()
	}

	rd := rand.Int() % len(dudefacts)
	rf := rand.Int() % len(dudefacts[rd].Facts)
	var output string

	output = dudefacts[rd].Nickname + ": "
	output += buildLineItem(rf, dudefacts[rd].Facts[rf])

	return output, nil
}

func init() {
	bot.RegisterPassiveCommand(
		"addToFactContext",
		addToFactContext)

	bot.RegisterCommand(
		"^addfact (\\S+) (.+)$",
		addFact)

	bot.RegisterCommand(
		"^votefact (\\S+) (\\S+) (\\S+)$",
		voteFact)

	bot.RegisterCommand(
		"^(\\S+)$",
		printAllUserFacts)

	bot.RegisterCommand(
		"^(\\S+) (\\S+)$",
		printUserFactContext)

	bot.RegisterCommand(
		"^rf$",
		printRandomFact)

	bot.RegisterCommand(
		"^rf (\\S+)$",
		printRandomFactUser)
}
