package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Database struct {
		Name     string `json:"Name"`
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"database"`
	Host string `json:"host"`
	Port string `json:"port"`
}

type SpellsList struct {
	ClassID int     `json:"class_id"`
	Spells  []Spell `json:"spells"`
}

type Spell struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ClassLevel int    `json:"class_level"`
}

func main() {
	dbConfig, err := LoadConfig()
	if err != nil {
		fmt.Println(err)
	}

	connectionString := fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%v",
		dbConfig.Database.User,
		dbConfig.Database.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database.Name,
	)

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Println(err)
	}

	defer db.Close()

	classSpellsData := [16]SpellsList{}

	b, err := ioutil.ReadFile("./templates/find-all-spells-page.template")
	if err != nil {
		fmt.Println(err)
	}

	spellsData := fmt.Sprintf("%v\n", string(b))

	for classID := 1; classID <= 16; classID++ {
		classIndex := classID - 1

		query := fmt.Sprintf(
			"SELECT id, `name`, classes%v FROM spells_new WHERE classes%v BETWEEN 71 AND 253",
			classID,
			classID,
		)

		res, err := db.Query(query)

		if err != nil {
			fmt.Println(err)
		}

		defer res.Close()

		for res.Next() {
			var currentSpell Spell
			err := res.Scan(&currentSpell.ID, &currentSpell.Name, &currentSpell.ClassLevel)

			if err != nil {
				fmt.Println(err)
			}

			classSpellsData[classIndex].ClassID = classID
			classSpellsData[classIndex].Spells = append(classSpellsData[classIndex].Spells, currentSpell)
		}
	}

	sort.Slice(classSpellsData[:], func(i, j int) bool {
		return classSpellsData[i].ClassID < classSpellsData[j].ClassID
	})

	for classID, currentSpellsList := range classSpellsData {
		spellsData += fmt.Sprintf("## %v\n", getClassName(currentSpellsList.ClassID))

		spellsData += "| ID | Name | Level |\n"

		spellsData += "| :--- | :--- | :--- |\n"

		for _, currentSpell := range currentSpellsList.Spells {
			spellsData += fmt.Sprintf("| %v | %v | %v |\n", currentSpell.ID, currentSpell.Name, currentSpell.ClassLevel)
		}

		if classID != 16 {
			spellsData += "\n"
		}
	}

	err = os.WriteFile("find_all_spells.md", []byte(spellsData), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
}

func LoadConfig() (Config, error) {
	var config Config

	configFile, err := os.Open("find_all_spells.json")

	if err != nil {
		return config, err
	}

	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return config, err
}

func getClassName(classID int) string {
	m := map[int]string{
		1:  "Warrior",
		2:  "Cleric",
		3:  "Paladin",
		4:  "Ranger",
		5:  "Shadow Knight",
		6:  "Druid",
		7:  "Monk",
		8:  "Bard",
		9:  "Rogue",
		10: "Shaman",
		11: "Necromancer",
		12: "Wizard",
		13: "Magician",
		14: "Enchanter",
		15: "Beastlord",
		16: "Berserker",
	}

	if _, ok := m[classID]; !ok {
		return "Unknown Class"
	}

	return m[classID]
}
