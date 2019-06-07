package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/markus-wa/demoinfocs-golang/events"
	"io"
	"log"
	"reflect"
	"syscall/js"

	dem "github.com/markus-wa/demoinfocs-golang"
	"github.com/markus-wa/demoinfocs-golang/common"
)

const (
	// WASM doesn't enjoy the big buffer sizes allocated by default
	msgQueueBufferSize = 2048 * 1024
)

type TeamStateInfo struct {
	ID			int		`bson:"ID" json:"ID"`
	Score		int		`bson:"Score" json:"Score"`
	ClanName	string	`bson:"ClanName" json:"ClanName"`
	Flag		string	`bson:"Flag" json:"Flag"`
}

type EquipmentElementStaticInfo struct {
	UniqueID	int64					`bson:"UniqueID" json:"UniqueID"`
	EntityID	int						`bson:"EntityID" json:"EntityID"`
	OwnerID		int64					`bson:"OwnerID" json:"OwnerID"`
	Weapon		common.EquipmentElement	`bson:"Weapon" json:"Weapon"`
}

type EquipmentInfo struct {
	UniqueID		int64					`bson:"UniqueID" json:"UniqueID"`
	Weapon			common.EquipmentElement `bson:"Weapon" json:"Weapon"`
	OwnerID			int64					`bson:"OwnerID" json:"OwnerID"`
	//AmmoType		int 					`bson:"AmmoType" json:"AmmoType"`
	AmmoInMagazine	int 					`bson:"AmmoInMagazine" json:"AmmoInMagazine"`
	AmmoReserve		int 					`bson:"AmmoReserve" json:"AmmoReserve"`
	ZoomLevel		int 					`bson:"ZoomLevel" json:"ZoomLevel"`
}

func NewEquipmentInfo(Equip common.Equipment) EquipmentInfo {
	EI := EquipmentInfo{
		Equip.UniqueID(),
		Equip.Weapon,
		-1,
		//Equip.AmmoType,
		Equip.AmmoInMagazine,
		0,
		Equip.ZoomLevel,
	}
	if Equip.Owner != nil {
		EI.OwnerID = Equip.Owner.SteamID
	}
	if Equip.AmmoReserve != 0 {
		EI.AmmoReserve = Equip.AmmoReserve
	}

	return EI
}

var EquipmentElements = make(map[int64]EquipmentElementStaticInfo)

func NewEquipmentElementStaticInfo(GP common.GrenadeProjectile) EquipmentElementStaticInfo {
	GPI := EquipmentElementStaticInfo{
		GP.UniqueID(),
		GP.EntityID,
		//-1,
		-1,
		GP.Weapon,
	}
	//if GP.Owner != nil {
	//	GPI.OwnerID = GP.Owner.SteamID
	//}
	if GP.Thrower != nil {
		GPI.OwnerID = GP.Thrower.SteamID
	}
	return GPI
}

func getMap(event interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})
	reflectedEvent := reflect.ValueOf(event)
	if reflectedEvent.Kind() != reflect.Struct {
		panic("getMap received a non-struct object")
	}
	if reflectedEvent.NumField() == 0 {
		return nil
	}
	for i := 0; i < reflectedEvent.NumField(); i++ {
		if field := reflectedEvent.Field(i); field.CanInterface() {
			switch field.Kind() {
			case reflect.Ptr:
				if field.IsNil() == false {
					switch field := reflect.Indirect(field); field.Type().Name() {
					//default:
					//	data[reflectedEvent.Type().Field(i).Name] = reflectedEvent.Field(i).Interface()
					case "Player":
						P := field.Interface().(common.Player)
						resultMap[reflectedEvent.Type().Field(i).Name] = P.SteamID
					case "GrenadeProjectile":
						GP := field.Interface().(common.GrenadeProjectile)
						if _, ok := EquipmentElements[GP.UniqueID()]; !ok {
							EquipmentElements[GP.UniqueID()] = NewEquipmentElementStaticInfo(GP)
						}
						resultMap[reflectedEvent.Type().Field(i).Name] = GP.UniqueID()
					case "Equipment":
						Equip := field.Interface().(common.Equipment)
						EI := NewEquipmentInfo(Equip)
						//EI := getMap(Equip)
						resultMap[reflectedEvent.Type().Field(i).Name] = EI
					case "TeamState":
						TS := field.Interface().(common.TeamState)
						resultMap[reflectedEvent.Type().Field(i).Name] = TeamStateInfo{
							TS.ID,
							TS.Score,
							TS.ClanName,
							TS.Flag,
						}
					case "Inferno":
						INF := field.Interface().(common.Inferno)
						resultMap[reflectedEvent.Type().Field(i).Name] = INF.UniqueID()
					case "BombEvent":
						BE := field.Interface().(events.BombEvent)
						resultMap["Player"] = BE.Player.SteamID
						resultMap["Site"] = BE.Site
					}
				} else {
					resultMap[reflectedEvent.Type().Field(i).Name] = -1
				}
			case reflect.Struct:
				resultMap[reflectedEvent.Type().Field(i).Name] = getMap(field.Interface())
			default:
				resultMap[reflectedEvent.Type().Field(i).Name] = field.Interface()
			}
		}
		//if field := reflectedEvent.Field(i); field.Kind() != reflect.Ptr && field.Kind() != reflect.Struct {
		//	if field.CanInterface() {
		//		resultMap[reflectedEvent.Type().Field(i).Name] = field.Interface()
		//	}
		//} else if field.CanInterface() {
		//
		//}
	}

	return resultMap
}

func main() {
	c := make(chan struct{}, 0)

	dem.DefaultParserConfig = dem.ParserConfig{
		MsgQueueBufferSize: msgQueueBufferSize,
	}

	registerCallbacks()

	fmt.Println("WASM Go Initialized")

	<-c

	dbgPrint("Returned from main()")
}

func registerCallbacks() {
	js.Global().Set("newParser", js.FuncOf(newParser))
}

// TODO: buffer reader/writer?

type parser struct {
	reader	io.ReadCloser
	writer	io.WriteCloser
	parser	*dem.Parser
	ready	bool
	header	common.DemoHeader
}

func md5hex(b []byte) string {
	x := md5.Sum(b)
	return hex.EncodeToString(x[:])
}

func (p *parser) write(b64 string) {
	b, err := base64.StdEncoding.DecodeString(b64)
	checkError(err)

	n, err := p.writer.Write(b)
	// It's fine if there's no reader and we can't write
	if n < len(b) && err != io.ErrClosedPipe {
		checkError(err)
	}
}

func dbgPrint(s interface{}) {
	fmt.Println(s)
}

func (p *parser) init() {
	dbgPrint("init")
	p.parser = dem.NewParser(p.reader)
	p.ready = true
}

func (p *parser) readHeader() {
	dbgPrint("readHeader")
	if !p.ready {
		p.init()
	}
	var err error
	p.header, err = p.parser.ParseHeader()
	checkError(err)
}

func (p *parser) getHeader(callback js.Value) common.DemoHeader {
	var err error
	defer func() {
		err = p.reader.Close()
		checkError(err)
	}()

	dbgPrint("getHeader")
	if !p.ready {
		dbgPrint("getHeader: !ready")
		p.init()
		p.readHeader()
	}

	b, err := json.Marshal(getMap(p.header))
	checkError(err)

	dbgPrint("Parsed")

	// Return result to JS
	callback.Invoke(string(b))
	return p.header
}

type playerStats struct {
	Name    string `json:"playerName"`
	Kills   int    `json:"Kills"`
	Deaths  int    `json:"Deaths"`
	Assists int    `json:"Assists"`
}

func statsFor(p *common.Player) playerStats {
	return playerStats{
		Name:    p.Name,
		Kills:   p.AdditionalPlayerInformation.Kills,
		Deaths:  p.AdditionalPlayerInformation.Deaths,
		Assists: p.AdditionalPlayerInformation.Assists,
	}
}

func newParser(_ js.Value, args []js.Value) interface{} {
	r, w := io.Pipe()
	p := &parser {
		reader:	r,
		writer:	w,
		parser:	nil,
		ready:	false,
	}

	m := map[string]interface{}{
		"write": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			p.write(args[0].String())
			return nil
		}),
		"close": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			var err error
			//err = r.Close()
			//checkError(err)
			err = w.Close()
			checkError(err)
			return nil
		}),
		"parseFinalStats": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			go p.parseFinalStats(args[0])
			return nil
		}),
		"init": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			p.init()
			return nil
		}),
		"readHeader": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			p.readHeader()
			return nil
		}),
		"getHeader": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			//dbgPrint(getMap(p.getHeader()))
			//h, err := json.Marshal(getMap(p.getHeader()))
			//			//checkError(err)
			//			//args[0].Invoke(string(h))
			//			//return h
			go p.getHeader(args[0])
			return nil
		}),
	}

	// Callback to signal that creation finished, ready to receive data
	args[0].Invoke(m)
	return m
}

//func (p *parser) drawReplay(callback js.Value) {
//	cvs.Start(30, Render)
//}

func (p *parser) parseFinalStats(callback js.Value) {
	var err error

	defer func() {
		err = p.reader.Close()
		checkError(err)
	}()

	p.init()
	//p.parser = dem.NewParser(p.reader)
	////


	//header := p.getHeader()
	header, err := p.parser.ParseHeader()
	checkError(err)
	// TODO: report headerpointer error
	//fmt.Println("Header:", header)
	dbgPrint("Map: " + header.MapName)

	//var players []*common.Player
	var stats []playerStats

	//parser.RegisterEventHandler(func(e events.MatchStartedChanged) {
	//	if e.NewIsStarted {
	//		players = make([]*common.Player, len(parser.GameState().Participants().Playing()))
	//		copy(players, parser.GameState().Participants().Playing())
	//	}
	//})

	p.parser.RegisterEventHandler(func(e events.RoundEnd) {
		if p.parser.GameState().TeamTerrorists().Score == 15 && e.Winner == common.TeamTerrorists ||
			p.parser.GameState().TeamCounterTerrorists().Score == 15 && e.Winner == common.TeamCounterTerrorists ||
			p.parser.GameState().TotalRoundsPlayed() == 29 {
			//game is over (works for mm)
			for _, p := range p.parser.GameState().Participants().Playing() {
				stats = append(stats, statsFor(p))
			}
		}
	})

	err = p.parser.ParseToEnd()
	checkError(err)

	dbgPrint("Parsed")


	////
	b, err := json.Marshal(stats)
	checkError(err)

	// Return result to JS
	callback.Invoke(string(b))
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
