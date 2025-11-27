package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

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
func parseOBJ(fileName string) {
	// Read the entire file into memory
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}
	// Split the file content into individual objects
	objects := strings.Split(string(data), "#")
	// Print each line
	for _, object := range objects {
		//Init Variables before parsing each object
		var parseCount int = 1

		var itemNumber string = ""
		var keywords string = ""
		var shortDesc string = ""
		var longDesc string = ""
		var actionDesc string = ""

		// Split the object into individual lines
		lines := strings.Split(string(object), "\n")
		for _, line := range lines {
			//Parse all lines and store data in relivant variables

			//Get Item Number
			if parseCount == 1 {
				parseCount++
				itemNumber = strings.TrimSpace(line)
				fmt.Println("Item Number: " + itemNumber)
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
					fmt.Println("Keywords: " + keywords)
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
					fmt.Println("ShortDesc: " + shortDesc)
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
					fmt.Println("LongDesc: " + longDesc)
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
						fmt.Println("ActionDesc: " + actionDesc)
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
					itemType := getItemType(flags[0])
					fmt.Println("Item Type: " + itemType)
					//Lookup extra flags from bitvector
					if flags[1] == "0" {
						extraFlags := "NONE"
						fmt.Println("Extra Flags: " + extraFlags)
					} else {
						bitvector, err := strconv.ParseUint(flags[1], 10, 64)
						if err != nil {
							extraFlags := "NONE"
							fmt.Println("Extra Flags: " + extraFlags)
						} else {
							bits := BitValues(bitvector, 30)
							extraFlags := getExtraFlags(bits)
							fmt.Println("Extra Flags:", extraFlags)
							//Look up each bit in the bitvector and print the flag name
						}
					}
					//Lookup wear flags from bitvector
					if flags[2] == "0" {
						wearFlags := "NONE"
						fmt.Println("Wear Flags: " + wearFlags)
					} else {
						bitvector, err := strconv.ParseUint(flags[2], 10, 64)
						if err != nil {
							extraFlags := "NONE"
							fmt.Println("Extra Flags: " + extraFlags)
						} else {
							bits := BitValues(bitvector, 21)
							wearFlags := getWearFlags(bits)
							fmt.Println("Wear Flags: " + wearFlags)
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
					fmt.Println("Value0: " + strings.TrimSpace(flags[0]))
					fmt.Println("Value1: " + strings.TrimSpace(flags[1]))
					fmt.Println("Value2: " + strings.TrimSpace(flags[2]))
					fmt.Println("Value3: " + strings.TrimSpace(flags[3]))
				}
				parseCount++
				continue
			}
			//Get weight, cost, rent
			if parseCount == 8 {
				flags := strings.Fields(line) //Line aways has 3 space delimited values
				if len(flags) >= 3 {
					fmt.Println("Weight: " + strings.TrimSpace(flags[0]))
					fmt.Println("Cost: " + strings.TrimSpace(flags[1]))
					fmt.Println("Rent: " + strings.TrimSpace(flags[2]))
				}
				parseCount++
				continue
			}
			//Get item load rate
			if parseCount == 9 {
				fmt.Println("LoadRate: " + strings.TrimSpace(line))
				parseCount++
				continue
			}
			//E X M T all indicate what comes on the next line and can occure more than once
			//T item timer value on next line
			//E extra descriptions
			//M wear_description
			//X action_description
			//A MAX_OBJ_AFFECT
			//B obj_flags.bitvector
			//C obj_flags.bitvector2
			//One more bitvector line always comes after X 0 0 0
			// obj_flags.extra_flags2, obj_flags.subclass_res, obj_flags.material

		}
	}
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
func parseMOB(fileName string) {
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
func parseZON(fileName string) {
	// Read the entire file into memory
	//data, err := ioutil.ReadFile(fileName)
	//if err != nil {
	//	log.Fatalf("failed to read file: %s", err)
	//}
	// Split the file content into lines
	//lines := strings.Split(string(data), "\n")
	// Print each line
	//for _, line := range lines {
	//	fmt.Println(line)
	//}
	fileName = strings.TrimRight(fileName, "zon")
	//parseWLD(fileName + "wld")
	//fileName = strings.TrimRight(fileName, "wld")
	parseOBJ(fileName + "obj")
	//fileName = strings.TrimRight(fileName, "obj")
	//parseMOB(fileName + "mob")
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
