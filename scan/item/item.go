package analyzer

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xackery/eqitemsniff/analyzer"
)

var fieldNames = []string{"itemclass", "name", "lore", "idfile", "lorefile", "id", "weight", "norent", "nodrop", "attunable", "size", "slots", "price", "icon", "benefitflag", "tradeskills", "cr", "dr", "pr", "mr", "fr", "svcorruption", "astr", "asta", "aagi", "adex", "acha", "aint", "awis", "hp", "mana", "endurance", "ac", "regen", "manaregen", "enduranceregen", "classes", "races", "deity", "skillmodvalue", "skillmodmax", "skillmodtype", "skillmodextra", "banedmgrace", "banedmgbody", "banedmgraceamt", "banedmgamt", "magic", "foodduration", "reqlevel", "reclevel", "reqskill", "bardtype", "bardvalue", "UNKNOWN02", "UNKNOWN03", "UNKNOWN04", "light", "delay", "elemdmgtype", "elemdmgamt", "therange", "damage", "color", "prestige", "UNKNOWN06", "UNKNOWN07", "UNKNOWN08", "itemtype", "material", "materialunk1", "elitematerial", "heroforge1", "heroforge2", "sellrate", "extradmgskill", "extradmgamt", "charmfileid", "mounteffect", "mountlevel2", "mounteffecttype", "mountlevel", "mountunk1", "mountunk2", "mountunk3", "mountunk4", "charmfile", "augtype", "augstricthidden", "augrestrict", "augslot1type", "augslot1visible", "augslot1unk", "augslot2type", "augslot2visible", "augslot2unk", "augslot3type", "augslot3visible", "augslot3unk", "augslot4type", "augslot4visible", "augslot4unk", "augslot5type", "augslot5visible", "augslot5unk", "augslot6type", "augslot6visible", "augslot6unk", "pointtype", "ldontheme", "ldonprice", "ldonsellbackrate", "ldonsold", "bagtype", "bagslots", "bagsize", "bagwr", "booktype", "booklang", "filename", "loregroup", "artifactflag", "summonedflag", "favor", "fvnodrop", "attack", "haste", "guildfavor", "augdistiller", "UNKNOWN09", "UNKNOWN10", "nopet", "UNKNOWN11", "stacksize", "notransfer", "expendablearrow", "UNKNOWN12", "UNKNOWN13", "clickeffect", "clicklevel2", "clicktype", "clicklevel", "maxcharges", "casttime", "recastdelay", "recasttype", "clickunk5", "clickname", "clickunk7", "proceffect", "proclevel2", "proctype", "proclevel", "prockunk1", "procunk2", "procunk3", "procunk4", "procrate", "procname", "procunk7", "worneffect", "wornlevel2", "worntype", "wornlevel", "wornunk1", "wornunk2", "wornunk3", "wornunk4", "wornunk5", "wornname", "wornunk7", "focuseffect", "focuslevel2", "focustype", "focuslevel", "focusunk1", "focusunk2", "focusunk3", "focusunk4", "focusunk5", "focusname", "focusunk7", "scrolleffect", "scrolllevel2", "scrolleffecttype", "scrolllevel", "scrollunk1", "scrollunk2", "scrollunk3", "scrollunk4", "scrollunk5", "scrollname", "scrollunk7", "bardeffect", "bardlevel2", "bardeffecttype", "bardlevel", "bardunk1", "bardunk2", "bardunk3", "bardunk4", "bardunk5", "bardname", "bardunk7", "mountunk5", "blessingeffect", "blessingname", "mountname", "mountunk7", "blessinglevel2", "blessingeffecttype", "blessinglevel", "blessingunk1", "rightclickscriptid", "questitemflag", "powersourcecap", "purity", "epicitem", "backstabdmg", "heroic_str", "heroic_int", "heroic_wis", "heroic_agi", "heroic_dex", "heroic_sta", "heroic_cha", "UNKNOWN29", "healamt", "spelldmg", "clairvoyance", "UNKNOWN30", "UNKNOWN31", "UNKNOWN32", "UNKNOWN33", "UNKNOWN34", "UNKNOWN35", "UNKNOWN36", "UNKNOWN37", "heirloom", "placeablebitfield", "UNKNOWN38", "UNKNOWN39", "UNKNOWN40", "UNKNOWN41", "UNKNOWN42", "UNKNOWN43", "UNKNOWN44", "placeablenpcname", "UNKNOWN46", "UNKNOWN47", "UNKNOWN48", "UNKNOWN49", "UNKNOWN50", "UNKNOWN51", "UNKNOWN52", "UNKNOWN53", "UNKNOWN54", "UNKNOWN55", "UNKNOWN56", "UNKNOWN57", "UNKNOWN58", "UNKNOWN59", "UNKNOWN60", "UNKNOWN61", "UNKNOWN62", "UNKNOWN63", "collectible", "nodestroy", "nonpc", "nozone", "UNKNOWN68", "UNKNOWN69", "UNKNOWN70", "UNKNOWN71", "noground", "UNKNOWN73", "marketplace", "freestorage", "UNKNOWN76", "UNKNOWN77", "UNKNOWN78", "UNKNOWN79", "evoitem", "evoid", "evolvl", "evomax", "convertitem", "convertid", "convertname", "blessingunk2", "blessingunk3", "blessingunk4", "blessingunk5", "blessingunk7", "familiareffect", "familiarlevel2", "familiareffecttype", "familiarlevel", "familiarunk1", "familiarunk2", "familiarunk3", "familiarunk4", "familiarunk5", "familiarname", "familiarunk7", "UNKNOWN80", "minluck", "maxluck", "loreequippedgroup", "updated", "created", "submitter", "verified", "verifiedby", "collectversion"}

// Scan returns an item
func Scan(packet *analyzer.EQPacket) *Item {
	dataSize := len(packet.Data)
	itOffset := -1
	for i, d := range packet.Data {
		if i >= dataSize-1 {
			return nil
		}
		if d == 0x49 && packet.Data[i+1] == 0x54 {
			itOffset = i
			break
		}
	}
	if itOffset == -1 {
		return nil
	}

	item := new(Item)
	length := 0
	behavior := 0

	idx := itOffset
	zeroCount := 0
	//a pattern regularly used to go from IT to beginning is, look for the 3 zero set values
	for ; idx > 0; idx-- {
		if packet.Data[idx] == 0 {
			zeroCount++
			if zeroCount == 3 {
				break
			}
		}
	}
	if idx == 0 {
		return nil
	}

	// the issue with this pattern, however, is that the if itemclass is set or not
	// it can cause a value to be off. So, let's see if next value is alphabetic in nature
	if packet.Data[idx+1] < 0x11 { //if low number, shift over one
		idx++
	}

	var val uint16
	var val32 uint32
	chars := []byte{}

	for _, fieldName := range fieldNames {
		length = item.length(fieldName)
		if length < 1 {
			//log.Warn().Msgf("length of %s not set", fieldName)
			continue
		}
		behavior = item.behavior(fieldName)

		if dataSize < idx+length {
			log.Debug().Msgf("ran out of packet at length %d field %s", idx+length, fieldName)
			return item
		}
		switch behavior {
		case 0: //simple uint16 parse
			switch length {
			case 1:
				val = uint16(packet.Data[idx])
				item.set(fieldName, fmt.Sprintf("%d", val))
				idx++
			case 2:
				val = binary.LittleEndian.Uint16(packet.Data[idx : idx+length])
				item.set(fieldName, fmt.Sprintf("%d", val))
				idx += length
			case 4:
				val32 = binary.LittleEndian.Uint32(packet.Data[idx : idx+length])
				item.set(fieldName, fmt.Sprintf("%d", val32))
				idx += length
			default:
			}
		case 1: //zero terminated string
			chars = []byte{}
			for ; packet.Data[idx] != 0x00 && idx < len(packet.Data); idx++ {
				chars = append(chars, packet.Data[idx])
			}
			idx++
			item.set(fieldName, fmt.Sprintf("%s", chars))
		default:
			//log.Warn().Msgf("unsupported behavior %d for %s", behavior, fieldName)
			continue
		}
		//if fieldName == "itemclass" {
		//	return item
		//}
	}

	return item
}

// InsertQuery returns a sqlx-compatible INSERT clause
func (item *Item) InsertQuery() string {
	fields := []string{}
	st := reflect.TypeOf(*item)

	preps := []string{}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag, ok := field.Tag.Lookup("db")
		if !ok {
			continue
		}
		fields = append(fields, fmt.Sprintf("`%s`", tag))
		preps = append(preps, fmt.Sprintf(":%s", tag))
	}

	return fmt.Sprintf("INSERT INTO items (%s) VALUES (%s);", strings.Join(fields, ", "), strings.Join(preps, ", "))
}

// newItem constructs an item struct based on a csv entry
func newItem(header []string, record []string) (*Item, error) {
	item := new(Item)

	if len(header) != len(record) {
		return nil, fmt.Errorf("header count (%d) does not match record count (%d)", len(header), len(record))
	}

	for i, field := range header {
		item.set(field, record[i])

		err := item.set(field, record[i])
		if err != nil {
			return nil, errors.Wrapf(err, "field %s", field)
		}

	}

	return item, nil
}

// set sets a property in an item using sodaeq struct tags
func (item *Item) set(fieldName string, value string) error {

	st := reflect.TypeOf(*item)
	sv := reflect.ValueOf(item)
	s := sv.Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag, ok := field.Tag.Lookup("sodaeq")
		if !ok {
			continue
		}
		if tag != fieldName {
			continue
		}
		pf := s.Field(i)

		if !pf.IsValid() {
			return fmt.Errorf("invalid value")
		}
		if !pf.CanSet() {
			return fmt.Errorf("cannot set")
		}
		switch pf.Kind() {
		case reflect.Int64:
			if strings.Contains(value, ".") {
				log.Debug().Msgf("field %s has value %s, converting to int will lose decimal", fieldName, value)
				value = value[0:strings.Index(value, ".")]
			}
			if value == "" {
				value = "0"
			}
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			pf.SetInt(val)
		case reflect.Float64:
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			pf.SetFloat(val)
		case reflect.String:
			pf.SetString(value)
		default:
			return fmt.Errorf("unknown type: %s", pf.Kind())
		}
		return nil
	}
	return fmt.Errorf("no sodaeq tag found")
}

// length returns a property in an item using length struct tags
func (item *Item) length(fieldName string) int {
	st := reflect.TypeOf(*item)

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag, ok := field.Tag.Lookup("sodaeq")
		if !ok {
			continue
		}
		if tag != fieldName {
			continue
		}
		tag, ok = field.Tag.Lookup("length")
		if !ok {
			return 0
		}

		length, _ := strconv.Atoi(tag)
		return length
	}
	return -1
}

// behavior returns a property in an item using type struct tags
func (item *Item) behavior(fieldName string) int {
	st := reflect.TypeOf(*item)

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag, ok := field.Tag.Lookup("sodaeq")
		if !ok {
			continue
		}
		if tag != fieldName {
			continue
		}
		tag, ok = field.Tag.Lookup("behavior")
		if !ok {
			return 0
		}
		length, _ := strconv.Atoi(tag)
		return length
	}
	return -1
}

// Item struct is used for in game items, and maps eqemu fields to sodaeq fields
type Item struct {
	ID                  int64          `db:"id" sodaeq:"id" length:"4"`                             // int(11) NOT NULL DEFAULT 0,
	Name                string         `db:"Name" sodaeq:"name" length:"10" behavior:"1"`           // varchar(64) NOT NULL DEFAULT '',
	Aagi                int64          `db:"aagi" sodaeq:"aagi" length:"1"`                         // int(11) NOT NULL DEFAULT 0,
	Ac                  int64          `db:"ac" sodaeq:"ac" length:"4"`                             // int(11) NOT NULL DEFAULT 0,
	Accuracy            int64          `db:"accuracy" sodaeq:"accuracy"`                            // int(11) NOT NULL DEFAULT 0,
	Acha                int64          `db:"acha" sodaeq:"acha" length:"1"`                         // int(11) NOT NULL DEFAULT 0,
	Adex                int64          `db:"adex" sodaeq:"adex" length:"1"`                         // int(11) NOT NULL DEFAULT 0,
	Aint                int64          `db:"aint" sodaeq:"aint" length:"1"`                         // int(11) NOT NULL DEFAULT 0,
	Artifactflag        int64          `db:"artifactflag" sodaeq:"artifactflag" length:"4"`         // tinyint(3) unsigned NOT NULL DEFAULT 0,
	Asta                int64          `db:"asta" sodaeq:"asta" length:"1"`                         // int(11) NOT NULL DEFAULT 0,
	Astr                int64          `db:"astr" sodaeq:"astr" length:"1"`                         // int(11) NOT NULL DEFAULT 0,
	Attack              int64          `db:"attack" sodaeq:"attack" length:"1"`                     // int(11) NOT NULL DEFAULT 0,
	Attuneable          int64          `db:"attuneable" sodaeq:"attunable" length:"1"`              // int(11) NOT NULL DEFAULT 0,
	Augdistiller        int64          `db:"augdistiller" sodaeq:"augdistiller" length:"10"`        // int(11) NOT NULL DEFAULT 0,
	Augrestrict         int64          `db:"augrestrict" sodaeq:"augrestrict" length:"1"`           // int(11) NOT NULL DEFAULT 0,
	Augslot1type        int64          `db:"augslot1type" sodaeq:"augslot1type" length:"1"`         // tinyint(3) NOT NULL DEFAULT 0,
	Augslot1unk2        int64          `db:"augslot1unk2" sodaeq:"augslot1unk" length:"1"`          // int(11) NOT NULL DEFAULT 0,
	Augslot1visible     int64          `db:"augslot1visible" sodaeq:"augslot1visible" length:"4"`   // tinyint(3) NOT NULL DEFAULT 0,
	Augslot2type        int64          `db:"augslot2type" sodaeq:"augslot2type" length:"1"`         // tinyint(3) NOT NULL DEFAULT 0,
	Augslot2unk2        int64          `db:"augslot2unk2" sodaeq:"augslot2unk" length:"1"`          // int(11) NOT NULL DEFAULT 0,
	Augslot2visible     int64          `db:"augslot2visible" sodaeq:"augslot2visible" length:"4"`   // tinyint(3) NOT NULL DEFAULT 0,
	Augslot3type        int64          `db:"augslot3type" sodaeq:"augslot3type" length:"1"`         // tinyint(3) NOT NULL DEFAULT 0,
	Augslot3unk2        int64          `db:"augslot3unk2" sodaeq:"augslot3unk" length:"1"`          // int(11) NOT NULL DEFAULT 0,
	Augslot3visible     int64          `db:"augslot3visible" sodaeq:"augslot3visible" length:"4"`   // tinyint(3) NOT NULL DEFAULT 0,
	Augslot4type        int64          `db:"augslot4type" sodaeq:"augslot4type" length:"1"`         // tinyint(3) NOT NULL DEFAULT 0,
	Augslot4unk2        int64          `db:"augslot4unk2" sodaeq:"augslot4unk" length:"1"`          // int(11) NOT NULL DEFAULT 0,
	Augslot4visible     int64          `db:"augslot4visible" sodaeq:"augslot4visible" length:"4"`   // tinyint(3) NOT NULL DEFAULT 0,
	Augslot5type        int64          `db:"augslot5type" sodaeq:"augslot5type" length:"1"`         // tinyint(3) NOT NULL DEFAULT 0,
	Augslot5unk2        int64          `db:"augslot5unk2" sodaeq:"augslot5unk" length:"1"`          // int(11) NOT NULL DEFAULT 0,
	Augslot5visible     int64          `db:"augslot5visible" sodaeq:"augslot5visible" length:"4"`   // tinyint(3) NOT NULL DEFAULT 0,
	Augslot6type        int64          `db:"augslot6type" sodaeq:"augslot6type" length:"1"`         // tinyint(3) NOT NULL DEFAULT 0,
	Augslot6unk2        int64          `db:"augslot6unk2" sodaeq:"augslot6unk" length:"1"`          // int(11) NOT NULL DEFAULT 0,
	Augslot6visible     int64          `db:"augslot6visible" sodaeq:"augslot6visible" length:"4"`   // tinyint(3) NOT NULL DEFAULT 0,
	Augstricthidden     int64          `sodaeq:"augstricthidden" length:"4"`                        // --- not supported ---
	Augtype             int64          `db:"augtype" sodaeq:"augtype" length:"1"`                   // int(11) NOT NULL DEFAULT 0,
	Avoidance           int64          `db:"avoidance" sodaeq:"avoidance"`                          // int(11) NOT NULL DEFAULT 0,
	Awis                int64          `db:"awis" sodaeq:"awis" length:"1"`                         // int(11) NOT NULL DEFAULT 0,
	Backstabdmg         int64          `db:"backstabdmg" sodaeq:"backstabdmg"`                      // smallint(6) NOT NULL DEFAULT 0,
	Bagsize             int64          `db:"bagsize" sodaeq:"bagsize" length:"4"`                   // int(11) NOT NULL DEFAULT 0,
	Bagslots            int64          `db:"bagslots" sodaeq:"bagslots" length:"1"`                 // int(11) NOT NULL DEFAULT 0,
	Bagtype             int64          `db:"bagtype" sodaeq:"bagtype" length:"1"`                   // int(11) NOT NULL DEFAULT 0,
	Bagwr               int64          `db:"bagwr" sodaeq:"bagwr" length:"1"`                       // int(11) NOT NULL DEFAULT 0,
	Banedmgamt          int64          `db:"banedmgamt" sodaeq:"banedmgamt" length:"4"`             // int(11) NOT NULL DEFAULT 0,
	Banedmgbody         int64          `db:"banedmgbody" sodaeq:"banedmgbody" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Banedmgrace         int64          `db:"banedmgrace" sodaeq:"banedmgrace" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Banedmgraceamt      int64          `db:"banedmgraceamt" sodaeq:"banedmgraceamt" length:"4"`     // int(11) NOT NULL DEFAULT 0,
	Bardeffect          int64          `db:"bardeffect" sodaeq:"bardeffect"`                        // smallint(6) NOT NULL DEFAULT 0,
	Bardeffecttype      int64          `db:"bardeffecttype" sodaeq:"bardeffecttype"`                // smallint(6) NOT NULL DEFAULT 0,
	Bardlevel           int64          `db:"bardlevel" sodaeq:"bardlevel"`                          // smallint(6) NOT NULL DEFAULT 0,
	Bardlevel2          int64          `db:"bardlevel2" sodaeq:"bardlevel2"`                        // smallint(6) NOT NULL DEFAULT 0,
	Bardname            string         `db:"bardname" sodaeq:"bardname"`                            // varchar(64) NOT NULL DEFAULT '',
	Bardtype            int64          `db:"bardtype" sodaeq:"bardtype" length:"4"`                 // int(11) NOT NULL DEFAULT 0,
	Bardunk1            int64          `db:"bardunk1" sodaeq:"bardunk1"`                            // smallint(6) NOT NULL DEFAULT 0,
	Bardunk2            int64          `db:"bardunk2" sodaeq:"bardunk2"`                            // smallint(6) NOT NULL DEFAULT 0,
	Bardunk3            int64          `db:"bardunk3" sodaeq:"bardunk3"`                            // smallint(6) NOT NULL DEFAULT 0,
	Bardunk4            int64          `db:"bardunk4" sodaeq:"bardunk4"`                            // smallint(6) NOT NULL DEFAULT 0,
	Bardunk5            int64          `db:"bardunk5" sodaeq:"bardunk5"`                            // smallint(6) NOT NULL DEFAULT 0,
	Bardunk7            int64          `db:"bardunk7" sodaeq:"bardunk7"`                            // smallint(6) NOT NULL DEFAULT 0,
	Bardvalue           int64          `db:"bardvalue" sodaeq:"bardvalue" length:"1"`               // int(11) NOT NULL DEFAULT 0,
	Benefitflag         int64          `db:"benefitflag" sodaeq:"benefitflag" length:"1"`           // int(11) NOT NULL DEFAULT 0,
	Blessingeffect      int64          `sodaeq:"blessingeffect"`                                    // --- not supported ---
	Blessingname        string         `sodaeq:"blessingname"`                                      // --- not supported ---
	Book                int64          `db:"book" sodaeq:"booklang"`                                // int(11) NOT NULL DEFAULT 0,
	Booktype            int64          `db:"booktype" sodaeq:"booktype" length:"4"`                 // int(11) NOT NULL DEFAULT 0,
	Casttime            int64          `db:"casttime" sodaeq:"casttime" length:"1"`                 // int(11) NOT NULL DEFAULT 0,
	Casttime2           int64          `db:"casttime_" sodaeq:"casttime_"`                          // int(11) NOT NULL DEFAULT 0,
	Charmfile           string         `db:"charmfile" sodaeq:"charmfile" length:"1"`               // varchar(32) NOT NULL DEFAULT '',
	Charmfileid         string         `db:"charmfileid" sodaeq:"charmfileid" length:"4"`           // varchar(32) NOT NULL DEFAULT '',
	Clairvoyance        int64          `db:"clairvoyance" sodaeq:"clairvoyance"`                    // smallint(6) NOT NULL DEFAULT 0,
	Classes             int64          `db:"classes" sodaeq:"classes" length:"4"`                   // int(11) NOT NULL DEFAULT 0,
	Clickeffect         int64          `db:"clickeffect" sodaeq:"clickeffect" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Clicklevel          int64          `db:"clicklevel" sodaeq:"clicklevel" length:"4"`             // int(11) NOT NULL DEFAULT 0,
	Clicklevel2         int64          `db:"clicklevel2" sodaeq:"clicklevel2" length:"1"`           // int(11) NOT NULL DEFAULT 0,
	Clickname           string         `db:"clickname" sodaeq:"clickname" length:"4"`               // varchar(64) NOT NULL DEFAULT '',
	Clicktype           int64          `db:"clicktype" sodaeq:"clicktype" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Clickunk5           int64          `db:"clickunk5" sodaeq:"clickunk5" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Clickunk6           string         `db:"clickunk6" sodaeq:"clickunk6"`                          // varchar(32) NOT NULL DEFAULT '',
	Clickunk7           int64          `db:"clickunk7" sodaeq:"clickunk7" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Color               int64          `db:"color" sodaeq:"color" length:"4"`                       // int(10) unsigned NOT NULL DEFAULT 0,
	Collectible         int64          `sodaeq:"collectible"`                                       // --- not supported ---
	Collectversion      int64          `sodaeq:"collectversion"`                                    // --- not supported ---
	Combateffects       string         `db:"combateffects" sodaeq:"combateffects"`                  // varchar(10) NOT NULL DEFAULT '',
	Comment             string         `db:"comment" sodaeq:"comment"`                              // varchar(255) NOT NULL DEFAULT '',
	Convertitem         int64          `sodaeq:"convertitem"`                                       // --- not supported ---
	Convertid           int64          `sodaeq:"convertid"`                                         // --- not supported ---
	Convertname         string         `sodaeq:"convertname"`                                       // --- not supported ---
	Cr                  int64          `db:"cr" sodaeq:"cr" length:"1"`                             // int(11) NOT NULL DEFAULT 0,
	Created             string         `db:"created" sodaeq:"created"`                              // varchar(64) NOT NULL DEFAULT '',
	Damage              int64          `db:"damage" sodaeq:"damage" length:"4"`                     // int(11) NOT NULL DEFAULT 0,
	Damageshield        int64          `db:"damageshield" sodaeq:"damageshield"`                    // int(11) NOT NULL DEFAULT 0,
	Deity               int64          `db:"deity" sodaeq:"deity" length:"4"`                       // int(11) NOT NULL DEFAULT 0,
	Delay               int64          `db:"delay" sodaeq:"delay" length:"1"`                       // int(11) NOT NULL DEFAULT 0,
	Dotshielding        int64          `db:"dotshielding" sodaeq:"dotshielding"`                    // int(11) NOT NULL DEFAULT 0,
	Dr                  int64          `db:"dr" sodaeq:"dr"  length:"1"`                            // int(11) NOT NULL DEFAULT 0,
	Dsmitigation        int64          `db:"dsmitigation" sodaeq:"dsmitigation"`                    // smallint(6) NOT NULL DEFAULT 0,
	Elemdmgamt          int64          `db:"elemdmgamt" sodaeq:"elemdmgamt" length:"1"`             // int(11) NOT NULL DEFAULT 0,
	Elemdmgtype         int64          `db:"elemdmgtype" sodaeq:"elemdmgtype" length:"1"`           // int(11) NOT NULL DEFAULT 0,
	Elitematerial       int64          `db:"elitematerial" sodaeq:"elitematerial" length:"4"`       // smallint(6) NOT NULL DEFAULT 0,
	Endur               int64          `db:"endur" sodaeq:"endurance" length:"4"`                   // int(11) NOT NULL DEFAULT 0,
	Enduranceregen      int64          `db:"enduranceregen" sodaeq:"enduranceregen" length:"4"`     // int(11) NOT NULL DEFAULT 0,
	Epicitem            int64          `db:"epicitem" sodaeq:"epicitem"`                            // int(11) NOT NULL DEFAULT 0,
	Evoid               int64          `db:"evoid" sodaeq:"evoid"`                                  // int(11) NOT NULL DEFAULT 0,
	Evoitem             int64          `db:"evoitem" sodaeq:"evoitem"`                              // int(11) NOT NULL DEFAULT 0,
	Evolvinglevel       int64          `db:"evolvinglevel" sodaeq:"evolvinglevel"`                  // int(11) NOT NULL DEFAULT 0,
	Evomax              int64          `db:"evomax" sodaeq:"evomax"`                                // int(11) NOT NULL DEFAULT 0,
	Evolvl              int64          `sodaeq:"evolvl"`                                            // --- not supported ---
	Expendablearrow     int64          `db:"expendablearrow" sodaeq:"expendablearrow" length:"4"`   // smallint(6) NOT NULL DEFAULT 0,
	Extradmgamt         int64          `db:"extradmgamt" sodaeq:"extradmgamt" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Extradmgskill       int64          `db:"extradmgskill" sodaeq:"extradmgskill"  length:"4"`      // int(11) NOT NULL DEFAULT 0,
	Factionamt1         int64          `db:"factionamt1" sodaeq:"factionamt1"`                      // int(11) NOT NULL DEFAULT 0,
	Factionamt2         int64          `db:"factionamt2" sodaeq:"factionamt2"`                      // int(11) NOT NULL DEFAULT 0,
	Factionamt3         int64          `db:"factionamt3" sodaeq:"factionamt3"`                      // int(11) NOT NULL DEFAULT 0,
	Factionamt4         int64          `db:"factionamt4" sodaeq:"factionamt4"`                      // int(11) NOT NULL DEFAULT 0,
	Factionmod1         int64          `db:"factionmod1" sodaeq:"factionmod1"`                      // int(11) NOT NULL DEFAULT 0,
	Factionmod2         int64          `db:"factionmod2" sodaeq:"factionmod2"`                      // int(11) NOT NULL DEFAULT 0,
	Factionmod3         int64          `db:"factionmod3" sodaeq:"factionmod3"`                      // int(11) NOT NULL DEFAULT 0,
	Factionmod4         int64          `db:"factionmod4" sodaeq:"factionmod4"`                      // int(11) NOT NULL DEFAULT 0,
	Favor               int64          `db:"favor" sodaeq:"favor" length:"1"`                       // int(11) NOT NULL DEFAULT 0,
	Filename            string         `db:"filename" sodaeq:"filename" length:"4"`                 // varchar(32) NOT NULL DEFAULT '',
	Focuseffect         int64          `db:"focuseffect" sodaeq:"focuseffect" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Focuslevel          int64          `db:"focuslevel" sodaeq:"focuslevel" length:"4"`             // int(11) NOT NULL DEFAULT 0,
	Focuslevel2         int64          `db:"focuslevel2" sodaeq:"focuslevel2" length:"1"`           // int(11) NOT NULL DEFAULT 0,
	Focusname           string         `db:"focusname" sodaeq:"focusname" length:"4"`               // varchar(64) NOT NULL DEFAULT '',
	Focustype           int64          `db:"focustype" sodaeq:"focustype" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Focusunk1           int64          `db:"focusunk1" sodaeq:"focusunk1" length:"1"`               // int(11) NOT NULL DEFAULT 0,
	Focusunk2           int64          `db:"focusunk2" sodaeq:"focusunk2" length:"1"`               // int(11) NOT NULL DEFAULT 0,
	Focusunk3           int64          `db:"focusunk3" sodaeq:"focusunk3" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Focusunk4           int64          `db:"focusunk4" sodaeq:"focusunk4" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Focusunk5           int64          `db:"focusunk5" sodaeq:"focusunk5" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Focusunk6           string         `db:"focusunk6" sodaeq:"focusunk6"`                          // varchar(32) NOT NULL DEFAULT '',
	Focusunk7           int64          `db:"focusunk7" sodaeq:"focusunk7" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Foodduration        int64          `sodaeq:"foodduration" length:"4"`                           // --- not supported ---
	Fr                  int64          `db:"fr" sodaeq:"fr" length:"1"`                             // int(11) NOT NULL DEFAULT 0,
	Freestorage         int64          `sodaeq:"freestorage"`                                       // int(11) NOT NULL DEFAULT 0,
	Fvnodrop            int64          `db:"fvnodrop" sodaeq:"fvnodrop" length:"1"`                 // int(11) NOT NULL DEFAULT 0,
	Guildfavor          int64          `db:"guildfavor" sodaeq:"guildfavor" length:"1"`             // int(11) NOT NULL DEFAULT 0,
	Haste               int64          `db:"haste" sodaeq:"haste" length:"1"`                       // int(11) NOT NULL DEFAULT 0,
	Healamt             int64          `db:"healamt" sodaeq:"healamt"`                              // smallint(6) NOT NULL DEFAULT 0,
	Heirloom            int64          `db:"heirloom" sodaeq:"heirloom"`                            // int(11) NOT NULL DEFAULT 0,
	Heroicagi           int64          `db:"heroic_agi" sodaeq:"heroic_agi"`                        // smallint(6) NOT NULL DEFAULT 0,
	Heroiccha           int64          `db:"heroic_cha" sodaeq:"heroic_cha"`                        // smallint(6) NOT NULL DEFAULT 0,
	Heroiccr            int64          `db:"heroic_cr" sodaeq:"heroic_cr"`                          // smallint(6) NOT NULL DEFAULT 0,
	Heroicdex           int64          `db:"heroic_dex" sodaeq:"heroic_dex"`                        // smallint(6) NOT NULL DEFAULT 0,
	Heroicdr            int64          `db:"heroic_dr" sodaeq:"heroic_dr"`                          // smallint(6) NOT NULL DEFAULT 0,
	Heroicfr            int64          `db:"heroic_fr" sodaeq:"heroic_fr"`                          // smallint(6) NOT NULL DEFAULT 0,
	Heroicint           int64          `db:"heroic_int" sodaeq:"heroic_int"`                        // smallint(6) NOT NULL DEFAULT 0,
	Heroicmr            int64          `db:"heroic_mr" sodaeq:"heroic_mr"`                          // smallint(6) NOT NULL DEFAULT 0,
	Heroicpr            int64          `db:"heroic_pr" sodaeq:"heroic_pr"`                          // smallint(6) NOT NULL DEFAULT 0,
	Heroicsta           int64          `db:"heroic_sta" sodaeq:"heroic_sta"`                        // smallint(6) NOT NULL DEFAULT 0,
	Heroicstr           int64          `db:"heroic_str" sodaeq:"heroic_str"`                        // smallint(6) NOT NULL DEFAULT 0,
	Heroicsvcorrup      int64          `db:"heroic_svcorrup" sodaeq:"heroic_svcorrup"`              // smallint(6) NOT NULL DEFAULT 0,
	Heroicwis           int64          `db:"heroic_wis" sodaeq:"heroic_wis"`                        // smallint(6) NOT NULL DEFAULT 0,
	Herosforgemodel     int64          `db:"herosforgemodel" sodaeq:"heroforge1"  length:"4"`       // int(11) NOT NULL DEFAULT 0,
	Herosforgemodel2    int64          `sodaeq:"heroforge2" length:"4"`                             // --- not supported ---
	Hp                  int64          `db:"hp" sodaeq:"hp" length:"4"`                             // int(11) NOT NULL DEFAULT 0,
	Icon                int64          `db:"icon" sodaeq:"icon" length:"4"`                         // int(11) NOT NULL DEFAULT 0,
	Idfile              string         `db:"idfile" sodaeq:"idfile" length:"2" behavior:"1"`        // varchar(30) NOT NULL DEFAULT '',
	Itemclass           int64          `db:"itemclass" sodaeq:"itemclass" length:"1"`               // int(11) NOT NULL DEFAULT 0,
	Itemtype            int64          `db:"itemtype" sodaeq:"itemtype" length:"1"`                 // int(11) NOT NULL DEFAULT 0,
	Ldonprice           int64          `db:"ldonprice" sodaeq:"ldonprice" length:"1"`               // int(11) NOT NULL DEFAULT 0,
	Ldonsellbackrate    int64          `db:"ldonsellbackrate" sodaeq:"ldonsellbackrate" length:"1"` // smallint(6) NOT NULL DEFAULT 0,
	Ldonsold            int64          `db:"ldonsold" sodaeq:"ldonsold" length:"4"`                 // int(11) NOT NULL DEFAULT 0,
	Ldontheme           int64          `db:"ldontheme" sodaeq:"ldontheme" length:"1"`               // int(11) NOT NULL DEFAULT 0,
	Light               int64          `db:"light" sodaeq:"light" length:"1"`                       // int(11) NOT NULL DEFAULT 0,
	Lore                string         `db:"lore" sodaeq:"lore" length:"2" behavior:"1"`            // varchar(80) NOT NULL DEFAULT '',
	Lorefile            string         `db:"lorefile" sodaeq:"lorefile" length:"1"`                 // varchar(32) NOT NULL DEFAULT '',
	Loregroup           int64          `db:"loregroup" sodaeq:"loregroup" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Magic               int64          `db:"magic" sodaeq:"magic" length:"1"`                       // int(11) NOT NULL DEFAULT 0,
	Mana                int64          `db:"mana" sodaeq:"mana" length:"4"`                         // int(11) NOT NULL DEFAULT 0,
	Manaregen           int64          `db:"manaregen" sodaeq:"manaregen" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Marketplace         int64          `sodaeq:"marketplace"`                                       // --- not supported ---
	Material            int64          `db:"material" sodaeq:"material" length:"4"`                 // int(11) NOT NULL DEFAULT 0,
	Materialunk1        int64          `sodaeq:"materialunk1"  length:"4"`                          // --- not supported ---
	Maxcharges          int64          `db:"maxcharges" sodaeq:"maxcharges" length:"1"`             // int(11) NOT NULL DEFAULT 0,
	Minstatus           int64          `db:"minstatus" sodaeq:"minstatus"`                          // smallint(5) NOT NULL DEFAULT 0,
	Mounteffect         int64          `sodaeq:"mounteffect"  length:"1"`
	Mounteffecttype     int64          `sodaeq:"mounteffecttype" length:"4"`
	Mountlevel          int64          `sodaeq:"mountlevel" length:"4"`
	Mountlevel2         int64          `sodaeq:"mountlevel2" length:"4"`
	Mountunk1           int64          `sodaeq:"mountunk1" length:"4"`
	Mountunk2           int64          `sodaeq:"mountunk2" length:"1"`
	Mountunk3           int64          `sodaeq:"mountunk3" length:"1"`
	Mountunk4           int64          `sodaeq:"mountunk4" length:"4"`
	Mr                  int64          `db:"mr" sodaeq:"mr" length:"1"`                       // int(11) NOT NULL DEFAULT 0,
	Nodestroy           int64          `sodaeq:"nodestroy"`                                   // --- not supported ---
	Nodrop              int64          `db:"nodrop" sodaeq:"nodrop" length:"1"`               // int(11) NOT NULL DEFAULT 0,
	Noground            int64          `sodaeq:"noground"`                                    // --- not supported ---
	Nonpc               int64          `sodaeq:"nonpc"`                                       // --- not supported ---
	Nopet               int64          `db:"nopet" sodaeq:"nopet" length:"1"`                 // int(11) NOT NULL DEFAULT 0,
	Norent              int64          `db:"norent" sodaeq:"norent" length:"1"`               // int(11) NOT NULL DEFAULT 0,
	Notransfer          int64          `db:"notransfer" sodaeq:"notransfer" length:"4"`       // int(11) NOT NULL DEFAULT 0,
	Nozone              int64          `sodaeq:"nozone"`                                      // --- not supported ---
	Pendingloreflag     int64          `db:"pendingloreflag" sodaeq:"pendingloreflag"`        // tinyint(3) unsigned NOT NULL DEFAULT 0,
	Placeable           int64          `db:"placeable" sodaeq:"placeable"`                    // int(11) NOT NULL DEFAULT 0,
	Placeablebitfield   int64          `sodaeq:"placeablebitfield"`                           // --- not supported ---
	Placeablenpcname    string         `sodaeq:"placeablenpcname"`                            // --- not supported ---
	Pointtype           int64          `db:"pointtype" sodaeq:"pointtype" length:"1"`         // int(11) NOT NULL DEFAULT 0,
	Potionbelt          int64          `db:"potionbelt" sodaeq:"potionbelt"`                  // int(11) NOT NULL DEFAULT 0,
	Potionbeltslots     int64          `db:"potionbeltslots" sodaeq:"potionbeltslots"`        // int(11) NOT NULL DEFAULT 0,
	Powersourcecapacity int64          `db:"powersourcecapacity" sodaeq:"powersourcecap"`     // smallint(6) NOT NULL DEFAULT 0,
	Pr                  int64          `db:"pr" sodaeq:"pr" length:"1"`                       // int(11) NOT NULL DEFAULT 0,
	Prestige            int64          `sodaeq:"prestige" length:"1"`                         // --- not supported ---
	Price               int64          `db:"price" sodaeq:"price" length:"4"`                 // int(11) NOT NULL DEFAULT 0,
	Proceffect          int64          `db:"proceffect" sodaeq:"proceffect" length:"4"`       // int(11) NOT NULL DEFAULT 0,
	Proclevel           int64          `db:"proclevel" sodaeq:"proclevel" length:"4"`         // int(11) NOT NULL DEFAULT 0,
	Proclevel2          int64          `db:"proclevel2" sodaeq:"proclevel2" length:"1"`       // int(11) NOT NULL DEFAULT 0,
	Procname            string         `db:"procname" sodaeq:"procname" length:"4"`           // varchar(64) NOT NULL DEFAULT '',
	Procrate            int64          `db:"procrate" sodaeq:"procrate" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Proctype            int64          `db:"proctype" sodaeq:"proctype" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Prockunk1           int64          `sodaeq:"prockunk1" length:"1"`                        // --- not supported ---
	Procunk1            int64          `db:"procunk1" sodaeq:"procunk1"`                      // int(11) NOT NULL DEFAULT 0,
	Procunk2            int64          `db:"procunk2" sodaeq:"procunk2" length:"1"`           // int(11) NOT NULL DEFAULT 0,
	Procunk3            int64          `db:"procunk3" sodaeq:"procunk3" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Procunk4            int64          `db:"procunk4" sodaeq:"procunk4" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Procunk6            string         `db:"procunk6" sodaeq:"procunk6"`                      // varchar(32) NOT NULL DEFAULT '',
	Procunk7            int64          `db:"procunk7" sodaeq:"procunk7" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Purity              int64          `db:"purity" sodaeq:"purity"`                          // int(11) NOT NULL DEFAULT 0,
	Questitemflag       int64          `db:"questitemflag" sodaeq:"questitemflag"`            // int(11) NOT NULL DEFAULT 0,
	Races               int64          `db:"races" sodaeq:"races" length:"4"`                 // int(11) NOT NULL DEFAULT 0,
	Range               int64          `db:"range" sodaeq:"therange" length:"1"`              // int(11) NOT NULL DEFAULT 0,
	Recastdelay         int64          `db:"recastdelay" sodaeq:"recastdelay" length:"4"`     // int(11) NOT NULL DEFAULT 0,
	Recasttype          int64          `db:"recasttype" sodaeq:"recasttype" length:"4"`       // int(11) NOT NULL DEFAULT 0,
	Reclevel            int64          `db:"reclevel" sodaeq:"reclevel" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Recskill            int64          `db:"recskill" sodaeq:"reqskill" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Regen               int64          `db:"regen" sodaeq:"regen" length:"4"`                 // int(11) NOT NULL DEFAULT 0,
	Reqlevel            int64          `db:"reqlevel" sodaeq:"reqlevel" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Rightclickscriptid  int64          `sodaeq:"rightclickscriptid"`                          // --- not supported ---
	Scriptfileid        int64          `db:"scriptfileid" sodaeq:"scriptfileid"`              // smallint(6) NOT NULL DEFAULT 0,
	Scrolleffect        int64          `db:"scrolleffect" sodaeq:"scrolleffect" length:"4"`   // int(11) NOT NULL DEFAULT 0,
	Scrolllevel         int64          `db:"scrolllevel" sodaeq:"scrolllevel"`                // int(11) NOT NULL DEFAULT 0,
	Scrolllevel2        int64          `db:"scrolllevel2" sodaeq:"scrolllevel2" length:"1"`   // int(11) NOT NULL DEFAULT 0,
	Scrollname          string         `db:"scrollname" sodaeq:"scrollname"`                  // varchar(64) NOT NULL DEFAULT '',
	Scrolltype          int64          `db:"scrolltype" sodaeq:"scrolleffecttype" length:"4"` // int(11) NOT NULL DEFAULT 0,
	Scrollunk1          int64          `db:"scrollunk1" sodaeq:"scrollunk1"`                  // int(11) NOT NULL DEFAULT 0,
	Scrollunk2          int64          `db:"scrollunk2" sodaeq:"scrollunk2"`                  // int(11) NOT NULL DEFAULT 0,
	Scrollunk3          int64          `db:"scrollunk3" sodaeq:"scrollunk3"`                  // int(11) NOT NULL DEFAULT 0,
	Scrollunk4          int64          `db:"scrollunk4" sodaeq:"scrollunk4"`                  // int(11) NOT NULL DEFAULT 0,
	Scrollunk5          int64          `db:"scrollunk5" sodaeq:"scrollunk5"`                  // int(11) NOT NULL DEFAULT 0,
	Scrollunk6          string         `db:"scrollunk6" sodaeq:"scrollunk6"`                  // varchar(32) NOT NULL DEFAULT '',
	Scrollunk7          int64          `db:"scrollunk7" sodaeq:"scrollunk7"`                  // int(11) NOT NULL DEFAULT 0,
	Sellrate            float64        `db:"sellrate" sodaeq:"sellrate" length:"4"`           // float NOT NULL DEFAULT 0,
	Serialization       sql.NullTime   `db:"serialization" sodaeq:"serialization"`            // text DEFAULT NULL,
	Serialized          sql.NullTime   `db:"serialized" sodaeq:"serialized"`                  // datetime DEFAULT NULL,
	Shielding           int64          `db:"shielding" sodaeq:"shielding"`                    // int(11) NOT NULL DEFAULT 0,
	Size                int64          `db:"size" sodaeq:"size" length:"1"`                   // int(11) NOT NULL DEFAULT 0,
	Skillmodmax         int64          `db:"skillmodmax" sodaeq:"skillmodmax" length:"4"`     // int(11) NOT NULL DEFAULT 0,
	Skillmodtype        int64          `db:"skillmodtype" sodaeq:"skillmodtype" length:"4"`   // int(11) NOT NULL DEFAULT 0,
	Skillmodvalue       int64          `db:"skillmodvalue" sodaeq:"skillmodvalue" length:"4"` // int(11) NOT NULL DEFAULT 0,
	Skillmodexta        int64          `sodaeq:"skillmodextra" length:"1"`                    // -- Skillmodextra not supported at this time --
	Slots               int64          `db:"slots" sodaeq:"slots" length:"4"`                 // int(11) NOT NULL DEFAULT 0,
	Source              string         `db:"source" sodaeq:"source"`                          // varchar(20) NOT NULL DEFAULT '',
	Spelldmg            int64          `db:"spelldmg" sodaeq:"spelldmg"`                      // smallint(6) NOT NULL DEFAULT 0,
	Spellshield         int64          `db:"spellshield" sodaeq:"spellshield"`                // int(11) NOT NULL DEFAULT 0,
	Stackable           int64          `db:"stackable" sodaeq:"stackable"`                    // int(11) NOT NULL DEFAULT 0,
	Stacksize           int64          `db:"stacksize" sodaeq:"stacksize" length:"4"`         // int(11) NOT NULL DEFAULT 0,
	Strikethrough       int64          `db:"strikethrough" sodaeq:"strikethrough"`            // int(11) NOT NULL DEFAULT 0,
	Stunresist          int64          `db:"stunresist" sodaeq:"stunresist"`                  // int(11) NOT NULL DEFAULT 0,
	Submitter           string         `sodaeq:"submitter"`                                   // --- not supported ---
	Summonedflag        int64          `db:"summonedflag" sodaeq:"summonedflag" length:"4"`   // tinyint(3) unsigned NOT NULL DEFAULT 0,
	Svcorruption        int64          `db:"svcorruption" sodaeq:"svcorruption" length:"1"`   // int(11) NOT NULL DEFAULT 0,
	Tradeskills         int64          `db:"tradeskills" sodaeq:"tradeskills" length:"1"`     // int(11) NOT NULL DEFAULT 0,
	UNKNOWN02           int64          `sodaeq:"UNKNOWN02" length:"1"`                        // --- not supported ---
	UNKNOWN03           int64          `sodaeq:"UNKNOWN03" length:"1"`                        // --- not supported ---
	UNKNOWN04           int64          `sodaeq:"UNKNOWN04" length:"1"`                        // --- not supported ---
	UNKNOWN06           int64          `sodaeq:"UNKNOWN06" length:"1"`                        // --- not supported ---
	UNKNOWN07           int64          `sodaeq:"UNKNOWN07" length:"1"`                        // --- not supported ---
	UNKNOWN08           int64          `sodaeq:"UNKNOWN08" length:"1"`                        // --- not supported ---
	UNKNOWN09           int64          `sodaeq:"UNKNOWN09" length:"10"`                       // --- not supported ---
	UNKNOWN10           int64          `sodaeq:"UNKNOWN10" length:"4"`                        // --- not supported ---
	UNKNOWN11           int64          `sodaeq:"UNKNOWN11" length:"1"`                        // --- not supported ---
	UNKNOWN12           int64          `sodaeq:"UNKNOWN12" length:"4"`                        // --- not supported ---
	UNKNOWN13           int64          `sodaeq:"UNKNOWN13" length:"4"`                        // --- not supported ---
	UNKNOWN14           string         `sodaeq:"UNKNOWN14"`                                   // --- not supported ---
	UNKNOWN17           string         `sodaeq:"UNKNOWN17"`                                   // --- not supported ---
	UNKNOWN18           string         `sodaeq:"UNKNOWN18"`                                   // --- not supported ---
	UNKNOWN19           string         `sodaeq:"UNKNOWN19"`                                   // --- not supported ---
	UNKNOWN20           string         `sodaeq:"UNKNOWN20"`                                   // --- not supported ---
	UNKNOWN21           string         `sodaeq:"UNKNOWN21"`                                   // --- not supported ---
	UNKNOWN22           string         `sodaeq:"UNKNOWN22"`                                   // --- not supported ---
	UNKNOWN29           string         `sodaeq:"UNKNOWN29"`                                   // --- not supported ---
	UNKNOWN30           string         `sodaeq:"UNKNOWN30"`                                   // --- not supported ---
	UNKNOWN31           string         `sodaeq:"UNKNOWN31"`                                   // --- not supported ---
	UNKNOWN32           string         `sodaeq:"UNKNOWN32"`                                   // --- not supported ---
	UNKNOWN33           string         `sodaeq:"UNKNOWN33"`                                   // --- not supported ---
	UNKNOWN34           string         `sodaeq:"UNKNOWN34"`                                   // --- not supported ---
	UNKNOWN35           string         `sodaeq:"UNKNOWN35"`                                   // --- not supported ---
	UNKNOWN36           string         `sodaeq:"UNKNOWN36"`                                   // --- not supported ---
	UNKNOWN37           string         `sodaeq:"UNKNOWN37"`                                   // --- not supported ---
	UNKNOWN38           string         `sodaeq:"UNKNOWN38"`                                   // --- not supported ---
	UNKNOWN39           string         `sodaeq:"UNKNOWN39"`                                   // --- not supported ---
	UNKNOWN40           string         `sodaeq:"UNKNOWN40"`                                   // --- not supported ---
	UNKNOWN41           string         `sodaeq:"UNKNOWN41"`                                   // --- not supported ---
	UNKNOWN42           string         `sodaeq:"UNKNOWN42"`                                   // --- not supported ---
	UNKNOWN43           string         `sodaeq:"UNKNOWN43"`                                   // --- not supported ---
	UNKNOWN44           string         `sodaeq:"UNKNOWN44"`                                   // --- not supported ---
	UNKNOWN46           string         `sodaeq:"UNKNOWN46"`                                   // --- not supported ---
	UNKNOWN47           string         `sodaeq:"UNKNOWN47"`                                   // --- not supported ---
	UNKNOWN48           string         `sodaeq:"UNKNOWN48"`                                   // --- not supported ---
	UNKNOWN49           string         `sodaeq:"UNKNOWN49"`                                   // --- not supported ---
	UNKNOWN50           string         `sodaeq:"UNKNOWN50"`                                   // --- not supported ---
	UNKNOWN51           string         `sodaeq:"UNKNOWN51"`                                   // --- not supported ---
	UNKNOWN52           string         `sodaeq:"UNKNOWN52"`                                   // --- not supported ---
	UNKNOWN53           string         `sodaeq:"UNKNOWN53"`                                   // --- not supported ---
	UNKNOWN54           string         `sodaeq:"UNKNOWN54"`                                   // --- not supported ---
	UNKNOWN55           string         `sodaeq:"UNKNOWN55"`                                   // --- not supported ---
	UNKNOWN56           string         `sodaeq:"UNKNOWN56"`                                   // --- not supported ---
	UNKNOWN57           string         `sodaeq:"UNKNOWN57"`                                   // --- not supported ---
	UNKNOWN58           string         `sodaeq:"UNKNOWN58"`                                   // --- not supported ---
	UNKNOWN59           string         `sodaeq:"UNKNOWN59"`                                   // --- not supported ---
	UNKNOWN60           string         `sodaeq:"UNKNOWN60"`                                   // --- not supported ---
	UNKNOWN61           string         `sodaeq:"UNKNOWN61"`                                   // --- not supported ---
	UNKNOWN62           string         `sodaeq:"UNKNOWN62"`                                   // --- not supported ---
	UNKNOWN63           string         `sodaeq:"UNKNOWN63"`                                   // --- not supported ---
	UNKNOWN68           string         `sodaeq:"UNKNOWN68"`                                   // --- not supported ---
	UNKNOWN69           string         `sodaeq:"UNKNOWN69"`                                   // --- not supported ---
	UNKNOWN70           string         `sodaeq:"UNKNOWN70"`                                   // --- not supported ---
	UNKNOWN71           string         `sodaeq:"UNKNOWN71"`                                   // --- not supported ---
	UNKNOWN73           string         `sodaeq:"UNKNOWN73"`                                   // --- not supported ---
	UNKNOWN76           string         `sodaeq:"UNKNOWN76"`                                   // --- not supported ---
	UNKNOWN77           string         `sodaeq:"UNKNOWN77"`                                   // --- not supported ---
	UNKNOWN78           string         `sodaeq:"UNKNOWN78"`                                   // --- not supported ---
	UNKNOWN79           string         `sodaeq:"UNKNOWN79"`                                   // --- not supported ---
	UNK012              int64          `db:"UNK012" sodaeq:"UNK012"`                          // int(11) NOT NULL DEFAULT 0,
	UNK013              int64          `db:"UNK013" sodaeq:"UNK013"`                          // int(11) NOT NULL DEFAULT 0,
	UNK014              int64          `db:"UNK014" sodaeq:"UNK014"`                          // int(11) NOT NULL DEFAULT 0,
	UNK033              int64          `db:"UNK033" sodaeq:"UNK033"`                          // int(11) NOT NULL DEFAULT 0,
	UNK054              int64          `db:"UNK054" sodaeq:"UNK054"`                          // int(11) NOT NULL DEFAULT 0,
	UNK059              int64          `db:"UNK059" sodaeq:"UNK059"`                          // int(11) NOT NULL DEFAULT 0,
	UNK060              int64          `db:"UNK060" sodaeq:"UNK060"`                          // int(11) NOT NULL DEFAULT 0,
	UNK120              int64          `db:"UNK120" sodaeq:"UNK120"`                          // int(11) NOT NULL DEFAULT 0,
	UNK121              int64          `db:"UNK121" sodaeq:"UNK121"`                          // int(11) NOT NULL DEFAULT 0,
	UNK123              int64          `db:"UNK123" sodaeq:"UNK123"`                          // int(11) NOT NULL DEFAULT 0,
	UNK124              int64          `db:"UNK124" sodaeq:"UNK124"`                          // int(11) NOT NULL DEFAULT 0,
	UNK127              int64          `db:"UNK127" sodaeq:"UNK127"`                          // int(11) NOT NULL DEFAULT 0,
	UNK132              sql.NullString `db:"UNK132" sodaeq:"UNK132"`                          // text CHARACTER SET utf8 DEFAULT NULL,
	UNK134              string         `db:"UNK134" sodaeq:"UNK134"`                          // varchar(255) NOT NULL DEFAULT '',
	UNK137              int64          `db:"UNK137" sodaeq:"UNK137"`                          // int(11) NOT NULL DEFAULT 0,
	UNK142              int64          `db:"UNK142" sodaeq:"UNK142"`                          // int(11) NOT NULL DEFAULT 0,
	UNK147              int64          `db:"UNK147" sodaeq:"UNK147"`                          // int(11) NOT NULL DEFAULT 0,
	UNK152              int64          `db:"UNK152" sodaeq:"UNK152"`                          // int(11) NOT NULL DEFAULT 0,
	UNK157              int64          `db:"UNK157" sodaeq:"UNK157"`                          // int(11) NOT NULL DEFAULT 0,
	UNK193              int64          `db:"UNK193" sodaeq:"UNK193"`                          // int(11) NOT NULL DEFAULT 0,
	UNK214              int64          `db:"UNK214" sodaeq:"UNK214"`                          // smallint(6) NOT NULL DEFAULT 0,
	UNK219              int64          `db:"UNK219" sodaeq:"UNK219"`                          // int(11) NOT NULL DEFAULT 0,
	UNK220              int64          `db:"UNK220" sodaeq:"UNK220"`                          // int(11) NOT NULL DEFAULT 0,
	UNK221              int64          `db:"UNK221" sodaeq:"UNK221"`                          // int(11) NOT NULL DEFAULT 0,
	UNK223              int64          `db:"UNK223" sodaeq:"UNK223"`                          // int(11) NOT NULL DEFAULT 0,
	UNK224              int64          `db:"UNK224" sodaeq:"UNK224"`                          // int(11) NOT NULL DEFAULT 0,
	UNK225              int64          `db:"UNK225" sodaeq:"UNK225"`                          // int(11) NOT NULL DEFAULT 0,
	UNK226              int64          `db:"UNK226" sodaeq:"UNK226"`                          // int(11) NOT NULL DEFAULT 0,
	UNK227              int64          `db:"UNK227" sodaeq:"UNK227"`                          // int(11) NOT NULL DEFAULT 0,
	UNK228              int64          `db:"UNK228" sodaeq:"UNK228"`                          // int(11) NOT NULL DEFAULT 0,
	UNK229              int64          `db:"UNK229" sodaeq:"UNK229"`                          // int(11) NOT NULL DEFAULT 0,
	UNK230              int64          `db:"UNK230" sodaeq:"UNK230"`                          // int(11) NOT NULL DEFAULT 0,
	UNK231              int64          `db:"UNK231" sodaeq:"UNK231"`                          // int(11) NOT NULL DEFAULT 0,
	UNK232              int64          `db:"UNK232" sodaeq:"UNK232"`                          // int(11) NOT NULL DEFAULT 0,
	UNK233              int64          `db:"UNK233" sodaeq:"UNK233"`                          // int(11) NOT NULL DEFAULT 0,
	UNK234              int64          `db:"UNK234" sodaeq:"UNK234"`                          // int(11) NOT NULL DEFAULT 0,
	UNK236              int64          `db:"UNK236" sodaeq:"UNK236"`                          // int(11) NOT NULL DEFAULT 0,
	UNK237              int64          `db:"UNK237" sodaeq:"UNK237"`                          // int(11) NOT NULL DEFAULT 0,
	UNK238              int64          `db:"UNK238" sodaeq:"UNK238"`                          // int(11) NOT NULL DEFAULT 0,
	UNK239              int64          `db:"UNK239" sodaeq:"UNK239"`                          // int(11) NOT NULL DEFAULT 0,
	UNK240              int64          `db:"UNK240" sodaeq:"UNK240"`                          // int(11) NOT NULL DEFAULT 0,
	UNK241              int64          `db:"UNK241" sodaeq:"UNK241"`                          // int(11) NOT NULL DEFAULT 0,
	Updated             string         `db:"updated" sodaeq:"updated"`                        // datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
	Verified            string         `db:"verified" sodaeq:"verified"`                      // datetime DEFAULT NULL,
	Verifiedby          string         `sodaeq:"verifiedby"`                                  // --- not supported ---
	Weight              int64          `db:"weight" sodaeq:"weight" length:"4"`               // int(11) NOT NULL DEFAULT 0,
	Worneffect          int64          `db:"worneffect" sodaeq:"worneffect" length:"4"`       // int(11) NOT NULL DEFAULT 0,
	Wornlevel           int64          `db:"wornlevel" sodaeq:"wornlevel" length:"4"`         // int(11) NOT NULL DEFAULT 0,
	Wornlevel2          int64          `db:"wornlevel2" sodaeq:"wornlevel2" length:"1"`       // int(11) NOT NULL DEFAULT 0,
	Wornname            string         `db:"wornname" sodaeq:"wornname" length:"4"`           // varchar(64) NOT NULL DEFAULT '',
	Worntype            int64          `db:"worntype" sodaeq:"worntype" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Wornunk1            int64          `db:"wornunk1" sodaeq:"wornunk1" length:"1"`           // int(11) NOT NULL DEFAULT 0,
	Wornunk2            int64          `db:"wornunk2" sodaeq:"wornunk2" length:"1"`           // int(11) NOT NULL DEFAULT 0,
	Wornunk3            int64          `db:"wornunk3" sodaeq:"wornunk3" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Wornunk4            int64          `db:"wornunk4" sodaeq:"wornunk4" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Wornunk5            int64          `db:"wornunk5" sodaeq:"wornunk5" length:"4"`           // int(11) NOT NULL DEFAULT 0,
	Wornunk6            string         `db:"wornunk6" sodaeq:"wornunk6"`                      // varchar(32) NOT NULL DEFAULT '',
	Wornunk7            int64          `db:"wornunk7" sodaeq:"wornunk7" length:"4"`           // int(11) NOT NULL DEFAULT 0,
}
