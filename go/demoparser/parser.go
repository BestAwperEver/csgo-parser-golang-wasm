package demoparser

import (
  "bytes"
  "github.com/golang/geo/r2"
  "github.com/golang/geo/r3"
  "github.com/markus-wa/demoinfocs-golang"
  "github.com/markus-wa/demoinfocs-golang/common"
  "github.com/markus-wa/demoinfocs-golang/events"
  "reflect"
  "syscall/js"
)

type EquipmentInfo struct {
  UniqueID		int64					`bson:"UniqueID"`
  Weapon			common.EquipmentElement `bson:"Weapon"`
  OwnerID			int64					`bson:"OwnerID"`
  //AmmoType		int 					`bson:"AmmoType"`
  AmmoInMagazine	int 					`bson:"AmmoInMagazine"`
  AmmoReserve		int 					`bson:"AmmoReserve"`
  ZoomLevel		int 					`bson:"ZoomLevel"`
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

type TeamStateInfo struct {
  ID			int		`bson:"ID"`
  Score		int		`bson:"Score"`
  ClanName	string	`bson:"ClanName"`
  Flag		string	`bson:"Flag"`
}

type InfernoInfo struct {
  UniqueID		int64		`bson:"UniqueID"`
  ConvexHull2D	[]r2.Point	`bson:"ConvexHull2D"`
}

type EquipmentElementStaticInfo struct {
  UniqueID	int64					`bson:"UniqueID"`
  EntityID	int						`bson:"EntityID"`
  OwnerID		int64					`bson:"OwnerID"`
  Weapon		common.EquipmentElement	`bson:"Weapon"`
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

type DemoParser struct {
  inBuf         []uint8
  outBuf        bytes.Buffer
  onDemoLoadCb  js.Func
  shutdownCb    js.Func
  initMemCb     js.Func
  getHeader     js.Func
  parser        *demoinfocs.Parser
  header        common.DemoHeader
  hasNextFrame  bool

  console js.Value
  done    chan struct{}
}

// New returns a new instance of parser
func New() *DemoParser {
  return &DemoParser {
    console: js.Global().Get("console"),
    done:    make(chan struct{}),
    hasNextFrame: false,
  }
}

// Start sets up all the callbacks and waits for the close signal
// to be sent from the browser.
func (dp *DemoParser) Start() {
  // Setup callbacks
  dp.setupInitMemCb()
  js.Global().Set("initMem", dp.initMemCb)

  dp.setupOnDemoLoadCb()
  js.Global().Set("loadDemo", dp.onDemoLoadCb)

  dp.setupShutdownCb()
  js.Global().Get("document").
    Call("getElementById", "close").
    Call("addEventListener", "click", dp.shutdownCb)

  dp.setupGetHeader()
  js.Global().Set("getHeader", dp.getHeader)

  <-dp.done
  dp.log("Shutting down app")
  dp.onDemoLoadCb.Release()
  dp.shutdownCb.Release()
}

// utility function to log a msg to the UI from inside a callback
func (dp *DemoParser) log(msg string) {
  js.Global().Get("document").
    Call("getElementById", "status").
    Set("innerText", msg)
}

func (dp *DemoParser) parseNextFrame() {
  if !dp.hasNextFrame {
    dp.log("There are no more frames in the demo")
    return
  }
  var err error
  dp.hasNextFrame, err = dp.parser.ParseNextFrame()
  dp.checkError(err)
}

type PlayerMovementInfo struct {
  SteamID		int64		`bson:"SteamID"`
  Team  common.Team `bson:"Team"`
  Position	r3.Vector	`bson:"Position"`
  ViewX		float32		`bson:"ViewX"`
  ViewY		float32		`bson:"ViewY"`
}

func NewPlayerMovementInfo(player *common.Player) PlayerMovementInfo {
  return PlayerMovementInfo{
    player.SteamID,
    player.Team,
    player.Position,
    player.ViewDirectionX,
    player.ViewDirectionY,
  }
}

func (dp *DemoParser) getPlayersPositions() []PlayerMovementInfo {
  PMIS := make([]PlayerMovementInfo, len(dp.parser.GameState().Participants().Playing()))
  for i, p := range dp.parser.GameState().Participants().Playing() {
    PMIS[i] = NewPlayerMovementInfo(p)
  }
  return PMIS
}

func (dp *DemoParser) checkError(err error) {
  if err != nil {
    dp.log(err.Error())
  }
}
