package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var zoneName string = ""
var zoneNumber string = ""

// 1 == PRINT PARSED ITEMS TO CONSOLE
var printZone int = 1
var printObjects int = 0
var printMobiles int = 0

const bitsPerInt32 = 32

type BitVector struct {
	array   []uint32
	length  int
	version int
}

func main() {
	//Hard coded path to Ronin world folder
	path := "../ronin/lib/world"
	files := listFiles(path)
	for _, v := range files {
		parseZON(v)
	}
}
func listFiles(dir string) []string {
	// Get a list of all .zon files in the specified directory
	root := os.DirFS(dir)
	mdFiles, err := fs.Glob(root, "*.zon")
	if err != nil {
		log.Fatal(err)
	}
	var files []string
	for _, v := range mdFiles {
		files = append(files, path.Join(dir, v))
	}
	return files
}

// PARSE ZONE FILE AND ALL ITS PARTS: .ZON, .WLD, .MOB, .OBJ =============
// DB.C, Line 1579, int read_zone(FILE *fl)
func parseZON(fileName string) {
	// Read the entire file into memory
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}
	//Init Variables before parsing each object
	var parseCount int = 1

	zoneName = ""
	zoneNumber = ""
	var lastRoomNum string = ""
	var respawnTimer string = ""
	var resetMode string = ""
	var zonCreationDate string = ""
	var zonUpdateDate string = ""
	var zonAuthor string = ""
	// Split the object into individual lines
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		//Parse all lines and store data in relivant variables
		var spawnMobileID string = ""
		var spawnMobileCount string = ""
		var spawnMobileRoomID string = ""
		var spawnMobileType string = ""
		var spawnItemID string = ""
		var spawnItemLocationID string = ""
		var spawnItemType string = ""
		var spawnDoorID string = ""
		var spawnDoorState string = ""

		var validData int = 0

		//Get Zone Number
		if parseCount == 1 {
			parseCount++
			validData = 1
			zoneNumber = strings.TrimSpace(strings.ReplaceAll(line, "#", ""))
			continue
		}
		//Get Zone Name
		if parseCount == 2 {
			zoneName = zoneName + " " + line

			if strings.Contains(line, "~") {
				validData = 1
				parseCount++
				zoneName = strings.TrimSpace(strings.ReplaceAll(line, "~", ""))
				continue
			}
		}
		//Get Highest room number, reset timer, reset mode
		if parseCount == 3 {
			parseCount++
			//Line aways has 3 space delimited values
			flags := strings.Fields(line)
			if len(flags) >= 3 {
				validData = 1
				lastRoomNum = strings.TrimSpace(flags[0])
				respawnTimer = strings.TrimSpace(flags[1])
				resetMode = getResetMode(strings.TrimSpace(flags[2]))
			}
			continue
		}
		if parseCount == 4 {
			//Freeform Parsing section. Look for first letter on line and parse accordingly
			//X ?, create date, revampDate, Author
			if strings.HasPrefix(line, "X ") {
				//X 0 17Jan2002 10Jul2021 NightInfinity
				flags := strings.Fields(line)
				if len(flags) >= 5 {
					validData = 1
					zonCreationDate = strings.TrimSpace(flags[2])
					zonUpdateDate = strings.TrimSpace(flags[3])
					zonAuthor = strings.TrimSpace(flags[4])
				}
			}
			//Need to parse E F G O D lines
			//'M': /* read a mobile */ db.c line 2623
			if strings.HasPrefix(line, "M ") {
				//M 0 27713 12 27769
				// Mobile# SpawnCount, Room
				flags := strings.Fields(line)
				if len(flags) >= 5 {
					validData = 1
					spawnMobileID = strings.TrimSpace(flags[2])
					spawnMobileCount = strings.TrimSpace(flags[3])
					spawnMobileRoomID = strings.TrimSpace(flags[4])
					spawnMobileType = "normal"
				}
			}
			//'F': /* follow a mobile */
			//MobileID, SpawnCount, RoomID
			if strings.HasPrefix(line, "F ") {
				flags := strings.Fields(line)
				if len(flags) >= 5 {
					validData = 1
					spawnMobileID = strings.TrimSpace(flags[2])
					spawnMobileCount = strings.TrimSpace(flags[3])
					spawnMobileRoomID = strings.TrimSpace(flags[4])
					spawnMobileType = "follow"
				}
			}
			//'R': /* add mount for M */
			//MobileID, SpawnCount, RoomID
			if strings.HasPrefix(line, "R ") {
				flags := strings.Fields(line)
				if len(flags) >= 5 {
					validData = 1
					spawnMobileID = strings.TrimSpace(flags[2])
					spawnMobileCount = strings.TrimSpace(flags[3])
					spawnMobileRoomID = strings.TrimSpace(flags[4])
					spawnMobileType = "mount"
				}
			}
			//'O': /* read an object */
			//Item#, ?, RoomID
			if strings.HasPrefix(line, "O ") {
				flags := strings.Fields(line)
				if len(flags) >= 5 {
					validData = 1
					spawnItemID = strings.TrimSpace(flags[2])
					spawnItemLocationID = strings.TrimSpace(flags[4])
					spawnItemType = "inRoom"
				}
			}
			//'P': /* object to object */
			if strings.HasPrefix(line, "P ") {
				flags := strings.Fields(line)
				if len(flags) >= 5 {
					validData = 1
					spawnItemID = strings.TrimSpace(flags[2])
					spawnItemLocationID = strings.TrimSpace(flags[4])
					spawnItemType = "inObject"
				}
			}
			//case 'T': /* take an object */
			if strings.HasPrefix(line, "T ") {
				flags := strings.Fields(line)
				if len(flags) >= 5 {
					validData = 1
					spawnItemID = strings.TrimSpace(flags[2])
					spawnItemLocationID = strings.TrimSpace(flags[4])
					spawnItemType = "takeObject"
				}
			}
			//'G': /* obj_to_char */
			//G 1 27742 0 0
			if strings.HasPrefix(line, "G ") {
				flags := strings.Fields(line)
				if len(flags) >= 3 {
					validData = 1
					spawnItemID = strings.TrimSpace(flags[2])
					spawnItemType = "ObjectToMob"
				}
			}
			//'E': /* object to equipment list */ Line 2813
			//E 1 27723 0 17
			if strings.HasPrefix(line, "E ") {
				flags := strings.Fields(line)
				if len(flags) >= 3 {
					validData = 1
					spawnItemID = strings.TrimSpace(flags[2])
					spawnItemType = "ObjectToMobEQ"
				}
			}
			//case 'D': /* set state of door */
			if strings.HasPrefix(line, "D ") {
				flags := strings.Fields(line)
				if len(flags) >= 5 {
					validData = 1
					spawnDoorID = strings.TrimSpace(flags[2])
					switch strings.TrimSpace(flags[4]) {
					case "0":
						spawnDoorState = "Open Unlocked"
					case "1":
						spawnDoorState = "Closed Unlocked"
					case "2":
						spawnDoorState = "Closed Locked"
					default:
						spawnDoorState = "invalid state " + strings.TrimSpace(flags[4])
					}
				}
			}
		}
		//Make sure data of some type was captured
		if validData == 1 {
			validData = 1
			//Item confirmed valid
			if printZone == 1 {
				fmt.Println("ZoneNumber: " + zoneNumber)
				fmt.Println("ZoneName: " + zoneName)
				fmt.Println("zonCreationDate: " + zonCreationDate)
				fmt.Println("zonUpdateDate: " + zonUpdateDate)
				fmt.Println("zonAuthor: " + zonAuthor)
				fmt.Println("respawnTimer: " + respawnTimer)
				fmt.Println("resetMode: " + resetMode)
				fmt.Println("lastRoomNum: " + lastRoomNum)
				fmt.Println("spawnMobileID: " + spawnMobileID)
				fmt.Println("spawnMobileCount: " + spawnMobileCount)
				fmt.Println("spawnMobileRoomID: " + spawnMobileRoomID)
				fmt.Println("spawnMobileType: " + spawnMobileType)
				fmt.Println("spawnItemID: " + spawnItemID)
				fmt.Println("spawnItemLocationID: " + spawnItemLocationID)
				fmt.Println("spawnItemType: " + spawnItemType)
				fmt.Println("spawnDoorID: " + spawnDoorID)
				fmt.Println("spawnDoorState: " + spawnDoorState)
			}
			//Write data to file for debugging and move to DB long term
			filename := "zon.csv"

			// Header row
			header := []string{
				"ZoneNumber",
				"ZoneName",
				"zonCreationDate",
				"zonUpdateDate",
				"zonAuthor",
				"respawnTimer",
				"resetMode",
				"lastRoomNum",
				"spawnMobileID",
				"spawnMobileCount",
				"spawnMobileRoomID",
				"spawnMobileType",
				"spawnItemID",
				"spawnItemLocationID",
				"spawnItemType",
				"spawnDoorID",
				"spawnDoorState",
			}

			// Check if file exists
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				// File does not exist, create it and write header
				file, err := os.Create(filename)
				if err != nil {
					log.Fatal(err)
				}
				writer := csv.NewWriter(file)
				if err := writer.Write(header); err != nil {
					log.Fatal(err)
				}
				writer.Flush()
				file.Close()
			}

			// Now open the file in append mode
			file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			writer := csv.NewWriter(file)
			defer writer.Flush()
			// Record row (convert all variables to strings)
			record := []string{
				zoneNumber,
				zoneName,
				zonCreationDate,
				zonUpdateDate,
				zonAuthor,
				respawnTimer,
				resetMode,
				lastRoomNum,
				spawnMobileID,
				spawnMobileCount,
				spawnMobileRoomID,
				spawnMobileType,
				spawnItemID,
				spawnItemLocationID,
				spawnItemType,
				spawnDoorID,
				spawnDoorState,
			}

			// Write record
			if err := writer.Write(record); err != nil {
				log.Fatal(err)
			}
		}
	}
	//Parse each file type
	fileName = strings.TrimRight(fileName, "zon")
	//parseWLD(fileName + "wld")
	//fileName = strings.TrimRight(fileName, "wld")
	parseOBJ(fileName + "obj")
	fileName = strings.TrimRight(fileName, "obj")
	parseMOB(fileName + "mob")
}

// PARSE WORLD FILES ===========================================
// DB.C, Line 765, void read_rooms(FILE *fl)
func parseWLD(fileName string) {
	// Read the entire file into memory
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}
	// Split the file content into lines
	lines := strings.Split(string(data), "\n")
	// Print each line
	for _, line := range lines {
		fmt.Println(line)
	}
}

// PARSE MOBILE FILES ===========================================
func parseMOB(fileName string) {
	// Read the entire file into memory
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}
	// Split the file content into individual objects
	objects := strings.Split(string(data), "#")
	for _, object := range objects {
		//Init Variables before parsing each object
		var parseCount int = 1

		var mobNumber string = ""

		// Split the object into individual lines
		lines := strings.Split(string(object), "\n")
		for _, line := range lines {
			//Parse all lines and store data in relivant variables

			//Get Mobile Number
			if parseCount == 1 {
				parseCount++
				mobNumber = strings.TrimSpace(line)
				continue
			}
		}
		//Make sure item number is actually a number and confirm its not 0 EOF
		if _, err := strconv.Atoi(mobNumber); err == nil {
			if mobNumber != "0" {
				//Item confirmed valid
				if printMobiles == 1 {
					fmt.Println("Zone: " + zoneName)
					fmt.Println("MobileNumber: " + mobNumber)

				}

			}
		}
	}
}

// PARSE OBJECT FILES ===========================================
func parseOBJ(fileName string) {
	// Read the entire file into memory
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}
	// Split the file content into individual objects
	objects := strings.Split(string(data), "#")
	for _, object := range objects {
		//Init Variables before parsing each object
		var parseCount int = 1
		var freeFormParse string = ""
		var captureNext string = "false"
		var affSlotNumber int = 0

		var itemNumber string = ""
		var keywords string = ""
		var shortDesc string = ""
		var longDesc string = ""
		var actionDesc string = ""
		var value0 string = ""
		var value1 string = ""
		var value2 string = ""
		var value3 string = ""
		var itemType string = ""
		var extraFlags string = ""
		var wearFlags string = ""
		var weight string = ""
		var cost string = ""
		var rent string = ""
		var loadrate string = ""
		var extraFlags2 string = ""
		var subclassFlags string = ""
		var materialFlags string = ""
		var objTimer string = ""
		var objAffFlags string = ""
		var objAffFlags2 string = ""
		var Affect0 string = ""
		var Affect1 string = ""
		var Affect2 string = ""
		var AffectModifier0 string = ""
		var AffectModifier1 string = ""
		var AffectModifier2 string = ""
		var lightColor string = ""
		var lightType string = ""
		var lightHours string = ""
		var recipeCreates string = ""
		var recipeRequires1 string = ""
		var recipeRequires2 string = ""
		var recipeRequires3 string = ""
		var aqOrderRequires1 string = ""
		var aqOrderRequires2 string = ""
		var aqOrderRequires3 string = ""
		var aqOrderRequires4 string = ""
		var spellLevel string = ""
		var spell string = ""
		var chargesCurrent string = ""
		var chargesMax string = ""
		var weaponSpecial string = ""
		var diceNumber string = ""
		var diceSize string = ""
		var damageMin string = ""
		var damageMax string = ""
		var damageAve string = ""
		var weaponType string = ""
		var gunLicense string = ""
		var bulletsLeft string = ""
		var affAC string = ""
		var maxContains string = ""
		var lockType string = ""
		var keyType string = ""
		var assignedKey string = ""
		var currentContains string = ""
		var liquidType string = ""
		var poisoned string = ""
		var gunNumber string = ""
		var foodSatiation string = ""
		var treasureCoins string = ""
		var picksCurrent string = ""
		var picksMax string = ""
		var boardReadLvl string = ""
		var boardWriteLvl string = ""
		var boardRemoveLvl string = ""
		var subclassPointValue string = ""

		// Split the object into individual lines
		lines := strings.Split(string(object), "\n")
		for _, line := range lines {
			//Parse all lines and store data in relivant variables

			//Get Item Number
			if parseCount == 1 {
				parseCount++
				itemNumber = strings.TrimSpace(line)
				continue
			}
			//Get Keywords
			if parseCount == 2 {
				//Check for ~ end of value symbol and remove it from keywords
				if strings.Contains(line, "~") == true {
					keywords = keywords + " "
					keywords = keywords + strings.TrimSpace(line)
					keywords = strings.ReplaceAll(keywords, "~", "")
					parseCount++
					continue
				} else {
					keywords = keywords + " "
					keywords = keywords + strings.TrimSpace(line)
					continue
				}

			}
			//Get Item Short Description
			if parseCount == 3 {
				//Check for ~ end of value symbol and remove it from keywords
				if strings.Contains(line, "~") == true {
					shortDesc = shortDesc + " "
					shortDesc = shortDesc + strings.TrimSpace(line)
					shortDesc = strings.ReplaceAll(shortDesc, "~", "")
					parseCount++
					continue
				} else {
					shortDesc = shortDesc + " "
					shortDesc = shortDesc + strings.TrimSpace(line)
					continue
				}

			}
			//Get Item Long Description
			if parseCount == 4 {
				//Check for ~ end of value symbol and remove it from keywords
				if strings.Contains(line, "~") == true {
					longDesc = longDesc + " "
					longDesc = longDesc + strings.TrimSpace(line)
					longDesc = strings.ReplaceAll(longDesc, "~", "")
					parseCount++
					continue
				} else {
					longDesc = longDesc + " "
					longDesc = longDesc + strings.TrimSpace(line)
					continue
				}
			}
			//Get Action Description
			if parseCount == 5 {
				if line == "~" {
					parseCount++
					fmt.Println("ActionDesc: " + actionDesc)
					continue
				} else {
					//Check for ~ end of value symbol and remove it from keywords
					if strings.Contains(line, "~") == true {
						actionDesc = actionDesc + " "
						actionDesc = actionDesc + strings.TrimSpace(line)
						actionDesc = strings.ReplaceAll(actionDesc, "~", "")
						parseCount++
						continue
					} else {
						actionDesc = actionDesc + " "
						actionDesc = actionDesc + strings.TrimSpace(line)
						continue
					}
				}
			}
			//Get item typeFlags, extraFlags, wearFlags
			if parseCount == 6 {
				//Line aways has 3 space delimited values
				flags := strings.Fields(line)
				if len(flags) >= 3 {
					itemType = getItemType(flags[0])
					//Lookup extra flags from bitvector
					if flags[1] == "0" {
						extraFlags = ""
					} else {
						bitvector, err := strconv.ParseUint(flags[1], 10, 64)
						if err != nil {
							extraFlags = ""
						} else {
							bits := BitValues(bitvector, 30)
							extraFlags = getExtraFlags(bits)
							//Look up each bit in the bitvector and print the flag name
						}
					}
					//Lookup wear flags from bitvector
					if flags[2] == "0" {
						wearFlags = ""
					} else {
						bitvector, err := strconv.ParseUint(flags[2], 10, 64)
						if err != nil {
							wearFlags = ""
						} else {
							bits := BitValues(bitvector, 21)
							wearFlags = getWearFlags(bits)
						}
					}
				}
				parseCount++
				continue
			}
			//Get values 0-3. These change based on item type
			//https://github.com/RoninMUD/ronin/blob/dev/src/create.c
			//line 10916
			if parseCount == 7 {
				flags := strings.Fields(line) //Line aways has 4 space delimited values
				if len(flags) >= 4 {
					value0 = strings.TrimSpace(flags[0])
					value1 = strings.TrimSpace(flags[1])
					value2 = strings.TrimSpace(flags[2])
					value3 = strings.TrimSpace(flags[3])
				}
				//Lookup specific data type and value based on item type
				switch itemType {
				case "LIGHT":
					//send_to_char("For light sources: <color> <type> <hours> <not used>\n\r", ch);
					lightColor = value0
					lightType = value1
					lightHours = value2
				case "RECIPE":
					//send_to_char("For Recipes: <Creates> <Requires> <Requires> <Requires> (-1 for none)\n\r", ch);
					recipeCreates = value0
					recipeRequires1 = value1
					recipeRequires2 = value2
					recipeRequires3 = value3
				case "AQ_ORDER":
					//send_to_char("For AQ Orders: <Requires> <Requires> <Requires> <Requires> (-1 for none)\n\r", ch);
					aqOrderRequires1 = value0
					aqOrderRequires2 = value1
					aqOrderRequires3 = value2
					aqOrderRequires4 = value3
				case "SCROLL", "POTION":
					//send_to_char("For Scrolls and Potions: <Level> <Spell1|0> <Spell2|0> <Spell3|0>\n\r", ch);
					spellLevel = value0
					spell = getSpell(value1)
					spell = spell + " " + getSpell(value2)
					spell = spell + " " + getSpell(value3)
					spell = strings.TrimSpace(spell)
				case "WAND", "STAFF":
					//send_to_char("For Staves and Wands: <Level> <Max Charges> <Charges Left> <Spell>\n\r", ch);
					spellLevel = value0
					chargesMax = value1
					chargesCurrent = value2
					spell = getSpell(value3)
				case "WEAPON", "2H-WEAPON":
					//send_to_char("For weapons: <olchelp weapon attacks> <dice damage> <dice size> <type>\n\r", ch);
					weaponSpecial = getWeaponSpecial(value0)
					diceNumber = value1
					diceSize = value2
					weaponType = getWeaponType(value3)
				case "FIRE WEAPON": //FIREARM
					//send_to_char("For guns: <license number> <bullets left> <dam dice number> <dam dice size>\n\r", ch);
					gunLicense = value0
					bulletsLeft = value1
					diceNumber = value2
					diceSize = value3
				case "MISSILE":
					//send_to_char("For thrown weapons: <dam dice number> <dam dice size> <unused> <unused>\n\r", ch);
					diceNumber = value0
					diceSize = value1
				case "ARMOR":
					//send_to_char("For armor: <AC apply (positive is better)> <unused> <unused> <unused>\n\r", ch);
					affAC = value0
				case "TRAP":
					//send_to_char("For traps: <spell> <damage> <unused> <unused>\n\r", ch);
					spell = value0
				case "CONTAINER":
					//send_to_char("For containers: <max contains> <how locked> <key #> <corpse>\n\r", ch);
					maxContains = value0
					lockType = value1
					assignedKey = value2
				case "LIQUID CONTAINER":
					//send_to_char("For drink containers: <max contains> <current contains> <liquid> <poisoned (0 = False, 1 = True)>\n\r", ch);
					maxContains = value0
					currentContains = value1
					liquidType = getLiquidType(value2)
					if value3 == "1" {
						poisoned = "TRUE"
					} else {
						poisoned = "FALSE"
					}
				case "BULLET":
					//send_to_char("For bullets: <unused> <unused> <gun #> <unused>\n\r", ch);
					gunNumber = value2
				case "KEY":
					//send_to_char("For keys: <keytype> <unused> <unused> <unused>\n\r", ch);
					keyType = value0
				case "FOOD":
					//send_to_char("For food: <feeds> <unused> <unused> <poisoned (0 = False, 1 = True)>\n\r", ch);
					foodSatiation = value0
					if value3 == "1" {
						poisoned = "TRUE"
					} else {
						poisoned = "FALSE"
					}
				case "MONEY":
					//send_to_char("For money: <coins> <unused> <unused> <unused>\n\r", ch);
					treasureCoins = value0
				case "LOCKPICK":
					//send_to_char("For lockpicks: <# picks> <max # picks> <unused> <unused>\n\r", ch);
					picksCurrent = value0
					picksMax = value1
				case "BOARD":
					//send_to_char("For boards: <min read level> <min write level> <min remove level> <unused>\n\r", ch);
					boardReadLvl = value0
					boardWriteLvl = value1
					boardRemoveLvl = value2
				case "SC_TOKEN":
					//send_to_char("For subclass tokens: <Subclass Points Given> <unused> <unused> <unused>\n\r", ch);
					subclassPointValue = value0
				}

				parseCount++
				continue
			}
			//Get weight, cost, rent
			if parseCount == 8 {
				flags := strings.Fields(line) //Line aways has 3 space delimited values
				if len(flags) >= 3 {
					weight = strings.TrimSpace(flags[0])
					cost = strings.TrimSpace(flags[1])
					rent = strings.TrimSpace(flags[2])
				}
				parseCount++
				continue
			}
			//Get item load rate
			if parseCount == 9 {
				loadrate = strings.TrimSpace(line)
				parseCount++
				continue
			}
			//Free form parsing after this point
			if parseCount >= 10 {
				//Look for free form markers and enable the proper parsing mode
				if strings.TrimSpace(line) == "A" {
					freeFormParse = "A"
					captureNext = "false"
				}
				if strings.TrimSpace(line) == "B" {
					freeFormParse = "B"
					captureNext = "false"
				}
				if strings.TrimSpace(line) == "C" {
					freeFormParse = "C"
					captureNext = "false"
				}
				if strings.TrimSpace(line) == "T" {
					freeFormParse = "T"
					captureNext = "false"
				}
				if strings.TrimSpace(line) == "X" {
					freeFormParse = "X"
					captureNext = "false"
				}
			}
			//A MAX_OBJ_AFFECT
			if freeFormParse == "A" {
				//Get apply types and values for AFF0 - AFF2
				if captureNext == "true" {
					captureNext = "false"
					flags := strings.Fields(line)
					if len(flags) >= 2 {
						//Lookup apply type from table
						flags[0] = strings.TrimSpace(flags[0])
						switch affSlotNumber {
						case 0:
							Affect0 = getApplyType(flags[0])
							AffectModifier0 = strings.TrimSpace(flags[1])
						case 1:
							Affect1 = getApplyType(flags[0])
							AffectModifier1 = strings.TrimSpace(flags[1])
						case 2:
							Affect2 = getApplyType(flags[0])
							AffectModifier2 = strings.TrimSpace(flags[1])
						}
						freeFormParse = ""
						affSlotNumber++
					}
				}
				captureNext = "true"
			}
			//B obj_flags.bitvector
			if freeFormParse == "B" {
				if captureNext == "true" {
					captureNext = "false"
					//Lookup objectAFF flags from bitvector
					bitvector, err := strconv.ParseUint(strings.TrimSpace(line), 10, 64)
					if err != nil {
						objAffFlags = ""
					} else {
						bits := BitValues(bitvector, 30)
						objAffFlags = getAffFlags(bits)
					}
					objAffFlags = strings.TrimSpace(objAffFlags)
					freeFormParse = ""
					continue
				}
				captureNext = "true"
			}
			//C obj_flags.bitvector2
			if freeFormParse == "C" {
				if captureNext == "true" {
					captureNext = "false"
					//Lookup objectAFF2 flags from bitvector
					bitvector, err := strconv.ParseUint(strings.TrimSpace(line), 10, 64)
					if err != nil {
						objAffFlags2 = ""
					} else {
						bits := BitValues(bitvector, 7)
						objAffFlags2 = getAFF2Flags(bits)
					}
					objAffFlags2 = strings.TrimSpace(objAffFlags2)
					freeFormParse = ""
					continue
				}
				captureNext = "true"
			}
			//T item timer value on next line
			if freeFormParse == "T" {
				if captureNext == "true" {
					captureNext = "false"
					objTimer = strings.TrimSpace(line)
					freeFormParse = ""
				}
				captureNext = "true"
			}
			//X action_description
			//One more bitvector line always comes after X 0 0 0
			// obj_flags.extra_flags2, obj_flags.subclass_res, obj_flags.material
			if freeFormParse == "X" {
				if captureNext == "true" {
					captureNext = "false"
					flags := strings.Fields(line) //Line aways has 3 space delimited values
					if len(flags) >= 3 {
						//Get extra flags 2 bitvector
						if flags[0] == "0" {
							extraFlags2 = ""
						} else {
							bitvector, err := strconv.ParseUint(flags[0], 10, 64)
							if err != nil {
								extraFlags2 = ""
							} else {
								bits := BitValues(bitvector, 12)
								extraFlags2 = getExtraFlags2(bits)
							}
						}
						//Lookup subclass restriction flags
						if flags[1] == "0" {
							subclassFlags = ""
						} else {
							bitvector, err := strconv.ParseUint(flags[1], 10, 64)
							if err != nil {
								subclassFlags = ""
							} else {
								bits := BitValues(bitvector, 20)
								subclassFlags = getSubclassFlags(bits)
							}
						}
						//Lookup material flags (UNUSED in Ronin)
						materialFlags = strings.TrimSpace(flags[2])
					}
					freeFormParse = ""
				}
				if strings.Contains(line, "~") == true {
					captureNext = "true"
				}
			}

		}
		//All data collected.
		//Calculate damage if dice exist
		if diceNumber != "" {
			var damageMinInt float64 = 0
			var damageMaxInt float64 = 0
			var damageAveInt float64 = 0
			var AFFmodifier float64 = 0
			//Convert strings to int
			diceNumberInt, err := strconv.ParseFloat(diceNumber, 64)
			if err != nil {
				// handle error
			}
			diceSizeInt, err := strconv.ParseFloat(diceSize, 64)
			if err != nil {
				// handle error
			}
			//Get AFF modifier
			if Affect0 == "DAMROLL" {
				affMod0int, err := strconv.ParseFloat(AffectModifier0, 64)
				if err != nil {
					// handle error
				}
				AFFmodifier = affMod0int
			}
			if Affect1 == "DAMROLL" {
				affMod1int, err := strconv.ParseFloat(AffectModifier1, 64)
				if err != nil {
					// handle error
				}
				AFFmodifier = affMod1int
			}
			if Affect2 == "DAMROLL" {
				affMod2int, err := strconv.ParseFloat(AffectModifier2, 64)
				if err != nil {
					// handle error
				}
				AFFmodifier = affMod2int
			}
			//Calculate min damage
			damageMinInt = diceNumberInt + AFFmodifier
			//Check for RANDOM flags +0-2 damroll
			if Affect0 == "DAMROLL" {
				matchedRandom, _ := regexp.MatchString(`(?:RANDOM\s|RANDOM$|RANDOM_0)`, extraFlags2)
				if matchedRandom {
					AFFmodifier = AFFmodifier + 2
				}
			}
			if Affect1 == "DAMROLL" {
				matchedRandom, _ := regexp.MatchString(`(?:RANDOM\s|RANDOM$|RANDOM_1)`, extraFlags2)
				if matchedRandom {
					AFFmodifier = AFFmodifier + 2
				}
			}
			if Affect2 == "DAMROLL" {
				matchedRandom, _ := regexp.MatchString(`(?:RANDOM\s|RANDOM$|RANDOM_2)`, extraFlags2)
				if matchedRandom {
					AFFmodifier = AFFmodifier + 2
				}
			}
			//Calculate damage numbers

			damageMaxInt = (diceNumberInt * diceSizeInt) + AFFmodifier
			// Average of one die
			avgDie := float64(diceSizeInt+1) / 2.0
			damageAveInt = float64(diceNumberInt)*avgDie + float64(AFFmodifier)
			//Convert values to strings
			// 'f' = decimal format, -1 = use necessary digits, 64 = float64
			damageMin = strconv.FormatFloat(damageMinInt, 'f', -1, 64)
			damageMax = strconv.FormatFloat(damageMaxInt, 'f', -1, 64)
			damageAve = strconv.FormatFloat(damageAveInt, 'f', -1, 64)
		}
		//Add max possible bonus from RANDOM flags to all AFF modifiers
		switch Affect0 {
		case "DAMROLL", "HITROLL", "HP_REGEN", "MANA_REGEN", "ARMOR", "MANA", "HIT", "MOVE":
			matchedRandom, _ := regexp.MatchString(`(?:RANDOM\s|RANDOM$|RANDOM_2)`, extraFlags2)
			if matchedRandom {
				tmp, err := strconv.Atoi(AffectModifier0)
				if err == nil {
					switch Affect0 {
					case "DAMROLL", "HITROLL": //Add +2 for random damage
						tmp = tmp + 2
						AffectModifier0 = strconv.Itoa(tmp)
					case "HP_REGEN": //Add +30 for HP REGEN
						tmp = tmp + 30
						AffectModifier0 = strconv.Itoa(tmp)
					case "MANA_REGEN": //Add +6 for MANA REGEN
						tmp = tmp + 6
						AffectModifier0 = strconv.Itoa(tmp)
					case "ARMOR": //Subtract 10 for ARMOR
						tmp = tmp - 10
						AffectModifier0 = strconv.Itoa(tmp)
					case "MANA", "HIT", "MOVE": //Add 100 for MANA HP MANA
						tmp = tmp + 100
						AffectModifier0 = strconv.Itoa(tmp)
					}
				}
			}
		}
		switch Affect1 {
		case "DAMROLL", "HITROLL", "HP_REGEN", "MANA_REGEN", "ARMOR", "MANA", "HIT", "MOVE":
			matchedRandom, _ := regexp.MatchString(`(?:RANDOM\s|RANDOM$|RANDOM_2)`, extraFlags2)
			if matchedRandom {
				tmp, err := strconv.Atoi(AffectModifier1)
				if err == nil {
					switch Affect1 {
					case "DAMROLL", "HITROLL": //Add +2 for random damage
						tmp = tmp + 2
						AffectModifier1 = strconv.Itoa(tmp)
					case "HP_REGEN": //Add +30 for HP REGEN
						tmp = tmp + 30
						AffectModifier1 = strconv.Itoa(tmp)
					case "MANA_REGEN": //Add +6 for MANA REGEN
						tmp = tmp + 6
						AffectModifier1 = strconv.Itoa(tmp)
					case "ARMOR": //Subtract 10 for ARMOR
						tmp = tmp - 10
						AffectModifier1 = strconv.Itoa(tmp)
					case "MANA", "HIT", "MOVE": //Add 100 for MANA HP MANA
						tmp = tmp + 100
						AffectModifier1 = strconv.Itoa(tmp)
					}
				}
			}
		}
		switch Affect2 {
		case "DAMROLL", "HITROLL", "HP_REGEN", "MANA_REGEN", "ARMOR", "MANA", "HIT", "MOVE":
			matchedRandom, _ := regexp.MatchString(`(?:RANDOM\s|RANDOM$|RANDOM_2)`, extraFlags2)
			if matchedRandom {
				tmp, err := strconv.Atoi(AffectModifier2)
				if err == nil {
					switch Affect2 {
					case "DAMROLL", "HITROLL": //Add +2 for random damage
						tmp = tmp + 2
						AffectModifier2 = strconv.Itoa(tmp)
					case "HP_REGEN": //Add +30 for HP REGEN
						tmp = tmp + 30
						AffectModifier2 = strconv.Itoa(tmp)
					case "MANA_REGEN": //Add +6 for MANA REGEN
						tmp = tmp + 6
						AffectModifier2 = strconv.Itoa(tmp)
					case "ARMOR": //Subtract 10 for ARMOR
						tmp = tmp - 10
						AffectModifier2 = strconv.Itoa(tmp)
					case "MANA", "HIT", "MOVE": //Add 100 for MANA HP MANA
						tmp = tmp + 100
						AffectModifier2 = strconv.Itoa(tmp)
					}
				}
			}
		}
		//Make sure item number is actually a number and confirm its not 0 EOF
		if _, err := strconv.Atoi(itemNumber); err == nil {
			if itemNumber != "0" {
				//Combine objAffFlags and objAffFlags2
				objAffFlags = objAffFlags + " " + objAffFlags2
				objAffFlags = strings.TrimSpace(objAffFlags)
				//Combine extra flags 1, 2, and subclass flags
				extraFlags = extraFlags + " " + subclassFlags
				extraFlags = strings.TrimSpace(extraFlags)
				extraFlags = extraFlags + " " + extraFlags2
				extraFlags = strings.TrimSpace(extraFlags)

				if printObjects == 1 {
					fmt.Println("Zone: " + zoneName)
					fmt.Println("ZoneID: " + zoneNumber)
					fmt.Println("Item Number: " + itemNumber)
					fmt.Println("Keywords: " + keywords)
					fmt.Println("ShortDesc: " + shortDesc)
					fmt.Println("LongDesc: " + longDesc)
					//fmt.Println("ActionDesc: " + actionDesc)
					fmt.Println("ItemType: " + itemType)
					fmt.Println("WearFlags: " + wearFlags)
					fmt.Println("ExtraFlags: " + extraFlags)
					fmt.Println("objAffFlags: " + objAffFlags)
					fmt.Println("Affect0: " + Affect0 + " " + AffectModifier0)
					fmt.Println("Affect1: " + Affect1 + " " + AffectModifier1)
					fmt.Println("Affect2: " + Affect2 + " " + AffectModifier2)
					fmt.Println("Value0: " + value0)
					fmt.Println("Value1: " + value1)
					fmt.Println("Value2: " + value2)
					fmt.Println("Value3: " + value3)
					fmt.Println("LightColor: " + lightColor)
					fmt.Println("LightType: " + lightType)
					fmt.Println("LightHours: " + lightHours)
					fmt.Println("RecipeCreates: " + recipeCreates)
					fmt.Println("RecipeRequires1: " + recipeRequires1)
					fmt.Println("RecipeRequires2: " + recipeRequires2)
					fmt.Println("RecipeRequires3: " + recipeRequires3)
					fmt.Println("aqOrderRequires1: " + aqOrderRequires1)
					fmt.Println("aqOrderRequires2: " + aqOrderRequires2)
					fmt.Println("aqOrderRequires3: " + aqOrderRequires3)
					fmt.Println("aqOrderRequires4: " + aqOrderRequires4)
					fmt.Println("maxContains: " + maxContains)
					fmt.Println("currentContains: " + currentContains)
					fmt.Println("liquidType: " + liquidType)
					fmt.Println("foodSatiation: " + foodSatiation)
					fmt.Println("poisoned: " + poisoned)
					fmt.Println("lockType: " + lockType)
					fmt.Println("assignedKey: " + assignedKey)
					fmt.Println("keyType: " + keyType)
					fmt.Println("picksCurrent: " + picksCurrent)
					fmt.Println("picksMax: " + picksMax)
					fmt.Println("affAC: " + affAC)
					fmt.Println("itemSpell: " + spell)
					fmt.Println("SpellLevel: " + spellLevel)
					fmt.Println("ChargesMax: " + chargesMax)
					fmt.Println("ChargesCurrent: " + chargesCurrent)
					fmt.Println("WeaponSpecial: " + weaponSpecial)
					fmt.Println("WeaponType: " + weaponType)
					fmt.Println("DiceNumber: " + diceNumber)
					fmt.Println("DiceSize: " + diceSize)
					fmt.Println("damageMin: " + damageMin)
					fmt.Println("damageMax: " + damageMax)
					fmt.Println("damageAve: " + damageAve)
					fmt.Println("GunLicense: " + gunLicense)
					fmt.Println("BulletsLeft: " + bulletsLeft)
					fmt.Println("GunNumber: " + gunNumber)
					fmt.Println("Weight: " + weight)
					fmt.Println("Cost: " + cost)
					fmt.Println("Rent: " + rent)
					fmt.Println("treasureCoins: " + treasureCoins)
					fmt.Println("subclassPointValue: " + subclassPointValue)
					fmt.Println("LoadRate: " + loadrate)
					fmt.Println("boardReadLvl: " + boardReadLvl)
					fmt.Println("boardWriteLvl: " + boardWriteLvl)
					fmt.Println("boardRemoveLvl: " + boardRemoveLvl)
					if materialFlags != "0" {
						fmt.Println("MaterialFlags: " + materialFlags)
					}
					fmt.Println("ObjTimer: " + objTimer)
					fmt.Println("-----------------------------------------------------")
				}
				//Write data to a CSV file =========================================================
				filename := "obj.csv"

				// Header row
				header := []string{
					"zone", "zoneNumber", "itemNumber", "keywords", "shortDesc", "longDesc", "itemType", "wearFlags", "extraFlags", "objAffFlags",
					"Affect0", "AffectModifier0", "Affect1", "AffectModifier1", "Affect2", "AffectModifier2",
					"value0", "value1", "value2", "value3",
					"lightColor", "lightType", "lightHours",
					"recipeCreates", "recipeRequires1", "recipeRequires2", "recipeRequires3",
					"aqOrderRequires1", "aqOrderRequires2", "aqOrderRequires3", "aqOrderRequires4",
					"maxContains", "currentContains", "liquidType", "foodSatiation", "poisoned",
					"lockType", "assignedKey", "keyType", "picksCurrent", "picksMax",
					"affAC", "spell", "spellLevel", "chargesMax", "chargesCurrent",
					"weaponSpecial", "weaponType", "diceNumber", "diceSize",
					"damageMin", "damageMax", "damageAve",
					"gunLicense", "bulletsLeft", "gunNumber",
					"weight", "cost", "rent", "treasureCoins", "subclassPointValue",
					"loadrate", "boardReadLvl", "boardWriteLvl", "boardRemoveLvl", "objTimer",
				}

				// Check if file exists
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					// File does not exist, create it and write header
					file, err := os.Create(filename)
					if err != nil {
						log.Fatal(err)
					}
					writer := csv.NewWriter(file)
					if err := writer.Write(header); err != nil {
						log.Fatal(err)
					}
					writer.Flush()
					file.Close()
				}

				// Now open the file in append mode
				file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				defer writer.Flush()
				// Record row (convert all variables to strings)
				record := []string{
					zoneName, zoneNumber, itemNumber, keywords, shortDesc, longDesc, itemType, wearFlags, extraFlags, objAffFlags,
					Affect0, AffectModifier0, Affect1, AffectModifier1, Affect2, AffectModifier2,
					value0, value1, value2, value3,
					lightColor, lightType, lightHours,
					recipeCreates, recipeRequires1, recipeRequires2, recipeRequires3,
					aqOrderRequires1, aqOrderRequires2, aqOrderRequires3, aqOrderRequires4,
					maxContains, currentContains, liquidType, foodSatiation, poisoned,
					lockType, assignedKey, keyType, picksCurrent, picksMax,
					affAC, spell, spellLevel, chargesMax, chargesCurrent,
					weaponSpecial, weaponType, diceNumber, diceSize,
					damageMin, damageMax, damageAve,
					gunLicense, bulletsLeft, gunNumber,
					weight, cost, rent, treasureCoins, subclassPointValue,
					loadrate, boardReadLvl, boardWriteLvl, boardRemoveLvl, objTimer,
				}

				// Write record
				if err := writer.Write(record); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
func getResetMode(mode string) string {
	switch mode {
	case "0":
		return "Never reset"
	case "1":
		return "Reset when empty"
	case "2":
		return "Reset on timer"
	case "3":
		return "Reset blocked"
	case "4":
		return "Reset locked"
	case "5":
		return "Doors only"
	default:
		return "UNKNOWN: " + mode
	}
}
func getSpell(spellNum string) string {
	switch spellNum {
	case "0":
		return ""
	case "1":
		return "armor"
	case "2":
		return "teleport"
	case "3":
		return "bless"
	case "4":
		return "blindness"
	case "5":
		return "burning_hands"
	case "6":
		return "call_lightning"
	case "7":
		return "charm_person"
	case "8":
		return "chill_touch"
	case "9":
		return "clone"
	case "10":
		return "colour_spray"
	case "11":
		return "control_weather"
	case "12":
		return "create_food"
	case "13":
		return "create_water"
	case "14":
		return "cure_blind"
	case "15":
		return "cure_critic"
	case "16":
		return "cure_light"
	case "17":
		return "curse"
	case "18":
		return "detect_alignment"
	case "19":
		return "detect_invisibility"
	case "20":
		return "detect_magic"
	case "21":
		return "detect_poison"
	case "22":
		return "dispel_evil"
	case "23":
		return "earthquake"
	case "24":
		return "enchant_weapon"
	case "25":
		return "energy_drain"
	case "26":
		return "fireball"
	case "27":
		return "harm"
	case "28":
		return "heal"
	case "29":
		return "invisibility"
	case "30":
		return "lightning_bolt"
	case "31":
		return "locate_object"
	case "32":
		return "magic_missile"
	case "33":
		return "poison"
	case "34":
		return "protection_from_evil"
	case "35":
		return "remove_curse"
	case "36":
		return "sanctuary"
	case "37":
		return "shocking_grasp"
	case "38":
		return "sleep"
	case "39":
		return "strength"
	case "40":
		return "summon"
	case "41":
		return "ventriloquate"
	case "42":
		return "word_of_recall"
	case "43":
		return "remove_poison"
	case "44":
		return "sense_life"
	case "71":
		return "identify"
	case "72":
		return "cure_serious"
	case "73":
		return "infravision"
	case "74":
		return "regeneration"
	case "75":
		return "vitality"
	case "76":
		return "cure_light_spray"
	case "77":
		return "cure_serious_spray"
	case "78":
		return "cure_critic_spray"
	case "79":
		return "heal_spray"
	case "80":
		return "death_spray"
	case "81":
		return "holy_word"
	case "82":
		return "iceball"
	case "83":
		return "total_recall"
	case "84":
		return "recharge"
	case "85":
		return "miracle"
	case "86":
		return "fly"
	case "87":
		return "mana_transfer"
	case "88":
		return "holy_bless"
	case "89":
		return "evil_bless"
	case "90":
		return "satiate"
	case "91":
		return "animate_dead"
	case "92":
		return "great_miracle"
	case "93":
		return "flamestrike"
	case "94":
		return "spirit_levy"
	case "95":
		return "lethal_fire"
	case "96":
		return "hold"
	case "97":
		return "sphere"
	case "98":
		return "imp_invisibility"
	case "99":
		return "invulnerability"
	case "100":
		return "fear"
	case "101":
		return "forget"
	case "102":
		return "fury"
	case "103":
		return "endure"
	case "104":
		return "blindness_dust"
	case "105":
		return "poison_smoke"
	case "106":
		return "hell_fire"
	case "107":
		return "hypnotize"
	case "108":
		return "recover_mana"
	case "109":
		return "thunderball"
	case "110":
		return "electric_shock"
	case "111":
		return "paralyze"
	case "112":
		return "remove_paralysis"
	case "113":
		return "dispel_good"
	case "114":
		return "evil_word"
	case "115":
		return "reappear"
	case "116":
		return "reveal"
	case "117":
		return "relocation"
	case "118":
		return "locate_character"
	case "119":
		return "super_harm"
	case "122":
		return "great_mana"
	case "124":
		return "perceive"
	case "127":
		return "haste"
	case "128":
		return "power_word_kill"
	case "129":
		return "dispel_magic"
	case "130":
		return "conflagration"
	case "132":
		return "convergence"
	case "133":
		return "enchant_armour"
	case "134":
		return "disintegrate"
	case "136":
		return "vampiric_touch"
	case "137":
		return "searing_orb"
	case "138":
		return "clairvoyance"
	case "139":
		return "firebreath"
	case "140":
		return "layhands"
	case "141":
		return "dispel_sanctuary"
	case "142":
		return "disenchant"
	case "143":
		return "petrify"
	case "145":
		return "protection_from_good"
	case "146":
		return "remove_improved_invisibility"
	case "149":
		return "quick"
	case "150":
		return "divine_intervention"
	case "151":
		return "rush"
	case "152":
		return "blood_lust"
	case "154":
		return "mystic_swiftness"
	case "165":
		return "wind_slash"
	case "172":
		return "debilitate"
	case "175":
		return "blur"
	case "179":
		return "tremor"
	case "180":
		return "shadow_wraith"
	case "181":
		return "devastation"
	case "187":
		return "power_of_faith"
	case "188":
		return "incendiary_cloud"
	case "189":
		return "power_of_devotion"
	case "190":
		return "wrath_of_god"
	case "191":
		return "disrupt_sanctuary"
	case "192":
		return "fortification"
	case "193":
		return "degenerate"
	case "194":
		return "magic_armament"
	case "195":
		return "ethereal_nature"
	case "196":
		return "engage"
	case "202":
		return "aid"
	case "206":
		return "desecrate"
	case "209":
		return "rimefang"
	case "210":
		return "wither"
	case "211":
		return "blackmantle"
	case "212":
		return "divine_wind"
	case "216":
		return "rejuvenation"
	case "217":
		return "wall_of_thorns"
	case "218":
		return "meteor"
	case "224":
		return "luck"
	case "225":
		return "warchant"
	case "226":
		return "rally"
	case "236":
		return "cloud_of_confusion"
	case "238":
		return "rage"
	case "239":
		return "righteousness"
	case "241":
		return "wrath_of_ancients"
	case "244":
		return "divine_hammer"
	case "245":
		return "camaraderie"
	case "246":
		return "orb_of_protection"
	case "247":
		return "dusk_requiem"
	case "248":
		return "frost_bolt"
	case "249":
		return "iron_skin"
	case "250":
		return "distortion"
	case "251":
		return "passdoor"
	case "252":
		return "blade_barrier"
	case "253":
		return "might"
	default:
		return "UNKNOWN: " + spellNum
	}
}
func getLiquidType(liquidType string) string {
	switch liquidType {
	case "0":
		return "water"
	case "1":
		return "water"
	case "2":
		return "beer"
	case "3":
		return "wine"
	case "4":
		return "ale"
	case "5":
		return "ale"
	case "6":
		return "whisky"
	case "7":
		return "lemonade"
	case "8":
		return "firebreather"
	case "9":
		return "local"
	case "10":
		return "juice"
	case "11":
		return "milk"
	case "12":
		return "tea"
	case "13":
		return "coffee"
	case "14":
		return "blood"
	case "15":
		return "salt"
	case "16":
		return "cola"
	case "17":
		return "stout"
	case "18":
		return "vodka"
	case "19":
		return "rum"
	case "20":
		return "liquor"
	case "21":
		return "champagne"
	case "22":
		return "bourbon"
	case "23":
		return "tequila"
	case "24":
		return "cider"
	case "25":
		return "urine"
	case "26":
		return "gin"
	case "27":
		return "merlot"
	case "28":
		return "schnapps"
	case "29":
		return "moonshine"
	case "30":
		return "pus"
	case "31":
		return "sherbet"
	case "32":
		return "cognac"
	case "33":
		return "brandy"
	case "34":
		return "scotch"
	case "35":
		return "kefir"
	case "36":
		return "ouzo"
	case "37":
		return "saki"
	case "38":
		return "lager"
	default:
		return "UNKNOWN"
	}
}
func getWeaponSpecial(weaponSpecial string) string {
	switch weaponSpecial {
	case "0":
		return ""
	case "1":
		return "Blind"
	case "2":
		return "Poison"
	case "3":
		return "Vampiric Touch"
	case "4":
		return "Chill Touch"
	case "5":
		return "Forget"
	case "6":
		return "Curse"
	case "7":
		return "Drain Mana"
	case "8":
		return "None"
	case "9":
		return "Power Word Kill"
	case "10":
		return "None"
	case "11":
		return "None"
	case "12":
		return "None"
	case "13":
		return "None"
	case "14":
		return "None"
	case "15":
		return "None"
	case "16":
		return "None"
	case "17":
		return "None"
	case "18":
		return "None"
	case "19":
		return "None"
	case "20":
		return "None"
	case "21":
		return "Slay Evil Beings"
	case "22":
		return "Slay Neutral Beings"
	case "23":
		return "Slay Good Beings"
	case "24":
		return "None"
	case "25":
		return "None"
	case "26":
		return "None"
	case "27":
		return "None"
	case "28":
		return "None"
	case "29":
		return "None"
	case "30":
		return "Chaotic"
	case "31":
		return "Slay Magic-Users"
	case "32":
		return "Slay Clerics"
	case "33":
		return "Slay Thieves"
	case "34":
		return "Slay Warriors"
	case "35":
		return "Slay Ninjas"
	case "36":
		return "Slay Nomads"
	case "37":
		return "Slay Paladins"
	case "38":
		return "Slay Anti-Paladins"
	case "39":
		return "Slay Avatars"
	case "40":
		return "Slay Bards"
	case "41":
		return "Slay Commandos"
	case "42":
		return "None"
	case "43":
		return "None"
	case "44":
		return "None"
	case "45":
		return "None"
	case "46":
		return "None"
	case "47":
		return "None"
	case "48":
		return "None"
	case "49":
		return "None"
	case "50":
		return "None"
	case "51":
		return "Slay Liches"
	case "52":
		return "Slay Lesser Undead"
	case "53":
		return "Slay Greater Undead"
	case "54":
		return "Slay Lesser Vampires"
	case "55":
		return "Slay Greater Vampires"
	case "56":
		return "Slay Lesser Dragons"
	case "57":
		return "Slay Greater Dragons"
	case "58":
		return "Slay Lesser Giants"
	case "59":
		return "Slay Greater Giants"
	case "60":
		return "Slay Lesser Lycanthropes"
	case "61":
		return "Slay Greater Lycanthropes"
	case "62":
		return "Slay Lesser Demons"
	case "63":
		return "Slay Greater Demons"
	case "64":
		return "Slay Lesser Elementals"
	case "65":
		return "Slay Greater Elementals"
	case "66":
		return "Slay Lesser Planars"
	case "67":
		return "Slay Greater Planars"
	case "68":
		return "Slay Humanoids"
	case "69":
		return "Slay Humans"
	case "70":
		return "Slay Halflings"
	case "71":
		return "Slay Dwarfs"
	case "72":
		return "Slay Elves"
	case "73":
		return "Slay Berserkers"
	case "74":
		return "Slay Kenders"
	case "75":
		return "Slay Trolls"
	case "76":
		return "Slay Insectoids"
	case "77":
		return "Slay Arachnoids"
	case "78":
		return "Slay Fungi"
	case "79":
		return "Slay Golems"
	case "80":
		return "Slay Reptiles"
	case "81":
		return "Slay Amphibians"
	case "82":
		return "Slay Dinosaurs"
	case "83":
		return "Slay Avians"
	case "84":
		return "Slay Fish"
	case "85":
		return "Slay Doppelgangers"
	case "86":
		return "Slay Animals"
	case "87":
		return "Slay Automatons"
	case "88":
		return "Slay Simians"
	case "89":
		return "Slay Canines"
	case "90":
		return "Slay Felines"
	case "91":
		return "Slay Bovines"
	case "92":
		return "Slay Plants"
	case "93":
		return "Slay Rodents"
	case "94":
		return "Slay Blobs"
	case "95":
		return "Slay Ghosts"
	case "96":
		return "Slay Orcs"
	case "97":
		return "Slay Gargoyles"
	case "98":
		return "Slay Invertibrates"
	case "99":
		return "Slay Drows"
	case "100":
		return "Slay Statues"
	default:
		return "UNKNOWN: " + weaponSpecial
	}
}
func getWeaponType(weaponType string) string {
	switch weaponType {
	case "0":
		return "whip"
	case "1":
		return "whip"
	case "2":
		return "whip"
	case "3":
		return "Slashing"
	case "4":
		return "Whip"
	case "5":
		return "Sting"
	case "6":
		return "Crush"
	case "7":
		return "Bludgeon"
	case "8":
		return "Claw"
	case "9":
		return "Pierce"
	case "10":
		return "Pierce"
	case "11":
		return "Pierce"
	case "12":
		return "Hack"
	case "13":
		return "Chop"
	case "14":
		return "Slice"
	default:
		return "UNKNOWN: " + weaponType
	}
}
func getApplyType(ApplyTypeCode string) string {
	switch ApplyTypeCode {
	case "0":
		return "NONE"
	case "1":
		return "STR"
	case "2":
		return "DEX"
	case "3":
		return "INT"
	case "4":
		return "WIS"
	case "5":
		return "CON"
	case "6":
		return "APPLY_6"
	case "7":
		return "APPLY_7"
	case "8":
		return "APPLY_8"
	case "9":
		return "AGE"
	case "10":
		return "APPLY_10"
	case "11":
		return "APPLY_10"
	case "12":
		return "MANA"
	case "13":
		return "HIT"
	case "14":
		return "MOVE"
	case "15":
		return "SAVING_ALL"
	case "16":
		return "APPLY_16"
	case "17":
		return "ARMOR"
	case "18":
		return "HITROLL"
	case "19":
		return "DAMROLL"
	case "20":
		return "SAVING_PARA"
	case "21":
		return "SAVING_ROD"
	case "22":
		return "SAVING_PETRI"
	case "23":
		return "SAVING_BREATH"
	case "24":
		return "SAVING_SPELL"
	case "25":
		return "SKILL_SNEAK"
	case "26":
		return "SKILL_HIDE"
	case "27":
		return "SKILL_STEAL"
	case "28":
		return "SKILL_BACKSTAB"
	case "29":
		return "SKILL_PICKLOCK"
	case "30":
		return "SKILL_KICK"
	case "31":
		return "SKILL_BASH"
	case "32":
		return "SKILL_RESCUE"
	case "33":
		return "SKILL_BLOCK"
	case "34":
		return "SKILL_KNOCK"
	case "35":
		return "SKILL_PUNCH"
	case "36":
		return "SKILL_PARRY"
	case "37":
		return "SKILL_DUAL"
	case "38":
		return "SKILL_THROW"
	case "39":
		return "SKILL_DODGE"
	case "40":
		return "SKILL_PEEK"
	case "41":
		return "SKILL_BUTCHER"
	case "42":
		return "SKILL_TRAP"
	case "43":
		return "SKILL_DISARM"
	case "44":
		return "SKILL_SUBDUE"
	case "45":
		return "SKILL_CIRCLE"
	case "46":
		return "SKILL_TRIPLE"
	case "47":
		return "SKILL_PUMMEL"
	case "48":
		return "SKILL_AMBUSH"
	case "49":
		return "SKILL_ASSAULT"
	case "50":
		return "SKILL_DISEMBOWEL"
	case "51":
		return "SKILL_TAUNT"
	case "52":
		return "HP_REGEN"
	case "53":
		return "MANA_REGEN"
	default:
		return "UNKNOWN: " + ApplyTypeCode
	}
}
func getAFF2Flags(bits []int) string {
	var af2Flags string = ""
	for i, v := range bits {
		if i == 0 && v == 1 {
			af2Flags = af2Flags + " " + "TRIPLE"
		}
		if i == 1 && v == 1 {
			af2Flags = af2Flags + " " + "IMMINENT-DEATH"
		}
		if i == 2 && v == 1 {
			af2Flags = af2Flags + " " + "SEVERED"
		}
		if i == 3 && v == 1 {
			af2Flags = af2Flags + " " + "QUAD"
		}
		if i == 4 && v == 1 {
			af2Flags = af2Flags + " " + "FORTIFICATION"
		}
		if i == 5 && v == 1 {
			af2Flags = af2Flags + " " + "PERCEIVE"
		}
		if i == 6 && v == 1 {
			af2Flags = af2Flags + " " + "RAGE"
		}
	}
	af2Flags = strings.TrimSpace(af2Flags)
	return af2Flags
}
func getAffFlags(bits []int) string {
	var afFlags string = ""
	for i, v := range bits {
		if i == 0 && v == 1 {
			afFlags = afFlags + " " + "BLIND"
		}
		if i == 1 && v == 1 {
			afFlags = afFlags + " " + "INVISIBLE"
		}
		if i == 2 && v == 1 {
			afFlags = afFlags + " " + "DETECT-ALIGNMENT"
		}
		if i == 3 && v == 1 {
			afFlags = afFlags + " " + "DETECT-INVISIBLE"
		}
		if i == 4 && v == 1 {
			afFlags = afFlags + " " + "DETECT-MAGIC"
		}
		if i == 5 && v == 1 {
			afFlags = afFlags + " " + "SENSE-LIFE"
		}
		if i == 6 && v == 1 {
			afFlags = afFlags + " " + "HOLD"
		}
		if i == 7 && v == 1 {
			afFlags = afFlags + " " + "SANCTUARY"
		}
		if i == 8 && v == 1 {
			afFlags = afFlags + " " + "GROUP"
		}
		if i == 9 && v == 1 {
			afFlags = afFlags + " " + "CONFUSION"
		}
		if i == 10 && v == 1 {
			afFlags = afFlags + " " + "CURSE"
		}
		if i == 11 && v == 1 {
			afFlags = afFlags + " " + "SPHERE"
		}
		if i == 12 && v == 1 {
			afFlags = afFlags + " " + "POISON"
		}
		if i == 13 && v == 1 {
			afFlags = afFlags + " " + "PROTECT-EVIL"
		}
		if i == 14 && v == 1 {
			afFlags = afFlags + " " + "PARALYSIS"
		}
		if i == 15 && v == 1 {
			afFlags = afFlags + " " + "INFRAVISION"
		}
		if i == 16 && v == 1 {
			afFlags = afFlags + " " + "STATUE"
		}
		if i == 17 && v == 1 {
			afFlags = afFlags + " " + "SLEEP"
		}
		if i == 18 && v == 1 {
			afFlags = afFlags + " " + "DODGE"
		}
		if i == 19 && v == 1 {
			afFlags = afFlags + " " + "SNEAK"
		}
		if i == 20 && v == 1 {
			afFlags = afFlags + " " + "HIDE"
		}
		if i == 21 && v == 1 {
			afFlags = afFlags + " " + "ANIMATE"
		}
		if i == 22 && v == 1 {
			afFlags = afFlags + " " + "CHARM"
		}
		if i == 23 && v == 1 {
			afFlags = afFlags + " " + "PROTECT-GOOD"
		}
		if i == 24 && v == 1 {
			afFlags = afFlags + " " + "FLY"
		}
		if i == 25 && v == 1 {
			afFlags = afFlags + " " + "SUBDUE"
		}
		if i == 26 && v == 1 {
			afFlags = afFlags + " " + "IMINV"
		}
		if i == 27 && v == 1 {
			afFlags = afFlags + " " + "INVUL"
		}
		if i == 28 && v == 1 {
			afFlags = afFlags + " " + "DUAL"
		}
		if i == 29 && v == 1 {
			afFlags = afFlags + " " + "FURY"
		}
	}
	afFlags = strings.TrimSpace(afFlags)
	return afFlags
}
func getSubclassFlags(bits []int) string {
	var subclassFlags string = ""
	for i, v := range bits {
		if i == 0 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_ENCHANTER"
		}
		if i == 1 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_ARCHMAGE"
		}
		if i == 2 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_DRUID"
		}
		if i == 3 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_TEMPLAR"
		}
		if i == 4 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_ROGUE"
		}
		if i == 5 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_BANDIT"
		}
		if i == 6 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_WARLORD"
		}
		if i == 7 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_GLADIATOR"
		}
		if i == 8 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_RONIN"
		}
		if i == 9 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_MYSTIC"
		}
		if i == 10 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_RANGER"
		}
		if i == 11 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_TRAPPER"
		}
		if i == 12 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_CAVALIER"
		}
		if i == 13 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_CRUSADER"
		}
		if i == 14 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_DEFILER"
		}
		if i == 15 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_INFIDEL"
		}
		if i == 16 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_BLADESINGER"
		}
		if i == 17 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_CHANTER"
		}
		if i == 18 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_LEGIONNAIRE"
		}
		if i == 19 && v == 1 {
			subclassFlags = subclassFlags + " " + "ANTI_SC_MERCENARY"
		}
	}
	subclassFlags = strings.TrimSpace(subclassFlags)
	return subclassFlags
}
func getExtraFlags2(bits []int) string {
	var extraFlags2 string = ""
	for i, v := range bits {
		if i == 0 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "RANDOM"
		}
		if i == 1 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "ALL_DECAY"
		}
		if i == 2 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "EQ_DECAY"
		}
		if i == 3 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "NO_GIVE"
		}
		if i == 4 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "NO_GIVE_MOB"
		}
		if i == 5 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "NO_PUT"
		}
		if i == 6 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "NO_TAKE_MOB"
		}
		if i == 7 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "NO_SCAVENGE"
		}
		if i == 8 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "NO_LOCATE"
		}
		if i == 9 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "RANDOM_0"
		}
		if i == 10 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "RANDOM_1"
		}
		if i == 11 && v == 1 {
			extraFlags2 = extraFlags2 + " " + "RANDOM_2"
		}
	}
	extraFlags2 = strings.TrimSpace(extraFlags2)
	return extraFlags2
}
func getWearFlags(bits []int) string {
	var wearFlags string = ""
	for i, v := range bits {
		if i == 0 && v == 1 {
			wearFlags = wearFlags + " " + "TAKE"
		}
		if i == 1 && v == 1 {
			wearFlags = wearFlags + " " + "FINGER"
		}
		if i == 2 && v == 1 {
			wearFlags = wearFlags + " " + "NECK"
		}
		if i == 3 && v == 1 {
			wearFlags = wearFlags + " " + "BODY"
		}
		if i == 4 && v == 1 {
			wearFlags = wearFlags + " " + "HEAD"
		}
		if i == 5 && v == 1 {
			wearFlags = wearFlags + " " + "LEGS"
		}
		if i == 6 && v == 1 {
			wearFlags = wearFlags + " " + "FEET"
		}
		if i == 7 && v == 1 {
			wearFlags = wearFlags + " " + "HANDS"
		}
		if i == 8 && v == 1 {
			wearFlags = wearFlags + " " + "ARMS"
		}
		if i == 9 && v == 1 {
			wearFlags = wearFlags + " " + "SHIELD"
		}
		if i == 10 && v == 1 {
			wearFlags = wearFlags + " " + "ABOUT"
		}
		if i == 11 && v == 1 {
			wearFlags = wearFlags + " " + "WAIST"
		}
		if i == 12 && v == 1 {
			wearFlags = wearFlags + " " + "WRIST"
		}
		if i == 13 && v == 1 {
			wearFlags = wearFlags + " " + "WIELD"
		}
		if i == 14 && v == 1 {
			wearFlags = wearFlags + " " + "HOLD"
		}
		if i == 15 && v == 1 {
			wearFlags = wearFlags + " " + "THROW"
		}
		if i == 16 && v == 1 {
			wearFlags = wearFlags + " " + "LIGHT-SOURCE"
		}
		if i == 17 && v == 1 {
			wearFlags = wearFlags + " " + "NO_REMOVE"
		}
		if i == 18 && v == 1 {
			wearFlags = wearFlags + " " + "NO_SCAVENGR"
		}
		if i == 19 && v == 1 {
			wearFlags = wearFlags + " " + "QUESTWEAR"
		}
		if i == 20 && v == 1 {
			wearFlags = wearFlags + " " + "2NECK"
		}
	}
	wearFlags = strings.TrimSpace(wearFlags)
	return wearFlags
}
func getExtraFlags(bits []int) string {
	var extraFlags string = ""
	for i, v := range bits {
		if i == 0 && v == 1 {
			extraFlags = extraFlags + " " + "GLOW"
		}
		if i == 1 && v == 1 {
			extraFlags = extraFlags + " " + "HUM"
		}
		if i == 2 && v == 1 {
			extraFlags = extraFlags + " " + "DARK"
		}
		if i == 3 && v == 1 {
			extraFlags = extraFlags + " " + "CLONED"
		}
		if i == 4 && v == 1 {
			extraFlags = extraFlags + " " + "EVIL"
		}
		if i == 5 && v == 1 {
			extraFlags = extraFlags + " " + "INVISIBLE"
		}
		if i == 6 && v == 1 {
			extraFlags = extraFlags + " " + "MAGICAL"
		}
		if i == 7 && v == 1 {
			extraFlags = extraFlags + " " + "NO_DROP"
		}
		if i == 8 && v == 1 {
			extraFlags = extraFlags + " " + "BLESSED"
		}
		if i == 9 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-GOOD"
		}
		if i == 10 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-EVIL"
		}
		if i == 11 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-NEUTRAL"
		}
		if i == 12 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-WARRIOR"
		}
		if i == 13 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-THIEF"
		}
		if i == 14 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-CLERIC"
		}
		if i == 15 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-MAGIC_USER"
		}
		if i == 16 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-MORTAL"
		}
		if i == 17 && v == 1 {
			extraFlags = extraFlags + " " + "NO_DISINTEGRATE"
		}
		if i == 18 && v == 1 {
			extraFlags = extraFlags + " " + "DISPELLED"
		}
		if i == 19 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-RENT"
		}
		if i == 20 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-NINJA"
		}
		if i == 21 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-NOMAD"
		}
		if i == 22 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-PALADIN"
		}
		if i == 23 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-ANTI-PALADIN"
		}
		if i == 24 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-AVATAR"
		}
		if i == 25 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-BARD"
		}
		if i == 26 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-COMMANDO"
		}
		if i == 27 && v == 1 {
			extraFlags = extraFlags + " " + "LIMITED"
		}
		if i == 28 && v == 1 {
			extraFlags = extraFlags + " " + "ANTI-AUCTION"
		}
		if i == 29 && v == 1 {
			extraFlags = extraFlags + " " + "CHAOTIC"
		}
	}
	extraFlags = strings.TrimSpace(extraFlags)
	return extraFlags
}
func BitValues(x uint64, width int) []int {
	bits := make([]int, width)
	for i := 0; i < width; i++ {
		if x&(1<<i) != 0 {
			bits[i] = 1
		}
	}
	return bits
}
func getItemType(typeFlag string) string {
	switch typeFlag {
	case "0":
		return "UNDEFINED"
	case "1":
		return "LIGHT"
	case "2":
		return "SCROLL"
	case "3":
		return "WAND"
	case "4":
		return "STAFF"
	case "5":
		return "WEAPON"
	case "6":
		return "FIRE WEAPON"
	case "7":
		return "MISSILE"
	case "8":
		return "TREASURE"
	case "9":
		return "ARMOR"
	case "10":
		return "POTION"
	case "11":
		return "WORN"
	case "12":
		return "OTHER"
	case "13":
		return "TRASH"
	case "14":
		return "TRAP"
	case "15":
		return "CONTAINER"
	case "16":
		return "NOTE"
	case "17":
		return "LIQUID CONTAINER"
	case "18":
		return "KEY"
	case "19":
		return "FOOD"
	case "20":
		return "MONEY"
	case "21":
		return "PEN"
	case "22":
		return "BOAT"
	case "23":
		return "BULLET"
	case "24":
		return "MUSICAL"
	case "25":
		return "LOCKPICK"
	case "26":
		return "2H-WEAPON"
	case "27":
		return "BOARD"
	case "28":
		return "TICKET"
	case "29":
		return "SC_TOKEN"
	case "30":
		return "SKIN"
	case "31":
		return "TROPHY"
	case "32":
		return "RECIPE"
	case "33":
		return "UNUSED"
	case "34":
		return "UNUSED"
	case "35":
		return "UNUSED"
	case "36":
		return "AQ_ORDER"
	default:
		return "UNKNOWN"
	}
}
