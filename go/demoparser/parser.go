package demoparser

import (
  "bytes"
  "fmt"
  "github.com/golang/geo/r2"
  "github.com/golang/geo/r3"
  "github.com/markus-wa/demoinfocs-golang"
  "github.com/markus-wa/demoinfocs-golang/common"
  "github.com/markus-wa/demoinfocs-golang/events"
  st "github.com/markus-wa/demoinfocs-golang/sendtables"
  "reflect"
  "strings"
  "syscall/js"
  "time"
)

type EquipmentInfo struct {
  UniqueID       int64                   `bson:"UniqueID"`
  Weapon         common.EquipmentElement `bson:"Weapon"`
  OwnerID        int64                   `bson:"OwnerID"`
  AmmoInMagazine int                     `bson:"AmmoInMagazine"`
  AmmoReserve    int                     `bson:"AmmoReserve"`
  ZoomLevel      int                     `bson:"ZoomLevel"`
  // AmmoType		int 					`bson:"AmmoType"`
}

func NewEquipmentInfo(Equip common.Equipment) EquipmentInfo {
  EI := EquipmentInfo{
    Equip.UniqueID(),
    Equip.Weapon,
    -1,
    // Equip.AmmoType,
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
  ID       int    `bson:"ID"`
  Score    int    `bson:"Score"`
  ClanName string `bson:"ClanName"`
  Flag     string `bson:"Flag"`
}

type InfernoInfo struct {
  UniqueID     int64      `bson:"UniqueID"`
  ConvexHull2D []r2.Point `bson:"ConvexHull2D"`
}

type EquipmentElementStaticInfo struct {
  UniqueID int64                   `bson:"UniqueID"`
  EntityID int                     `bson:"EntityID"`
  OwnerID  int64                   `bson:"OwnerID"`
  Weapon   common.EquipmentElement `bson:"Weapon"`
}

var EquipmentElements = make(map[int64]EquipmentElementStaticInfo)

func NewEquipmentElementStaticInfo(GP common.GrenadeProjectile) EquipmentElementStaticInfo {
  GPI := EquipmentElementStaticInfo{
    GP.UniqueID(),
    GP.EntityID,
    // -1,
    -1,
    GP.Weapon,
  }
  // if GP.Owner != nil {
  //	GPI.OwnerID = GP.Owner.SteamID
  // }
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
          // default:
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
            // EI := getMap(Equip)
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
    // if field := reflectedEvent.Field(i); field.Kind() != reflect.Ptr && field.Kind() != reflect.Struct {
    //	if field.CanInterface() {
    //		resultMap[reflectedEvent.Type().Field(i).Name] = field.Interface()
    //	}
    // } else if field.CanInterface() {
    //
    // }
  }

  return resultMap
}

type DemoParser struct {
  inBuf            []uint8
  outBuf           bytes.Buffer
  onDemoLoadCb     js.Func
  shutdownCb       js.Func
  initMemCb        js.Func
  getHeader        js.Func
  nextFrame        js.Func
  getPos           js.Func
  getChickenPos    js.Func
  getWeaponPos     js.Func
  inverseTranslate js.Func
  translate        js.Func
  getBomb          js.Func
  decodePositions  js.Func
  parser           *demoinfocs.Parser
  header           common.DemoHeader
  hasNextFrame     bool
  firing           map[int64]time.Duration
  bombDefused      bool

  chickensByEntityID map[int]*Chicken
  weaponsByEntityID  map[int]*Weapon
  doorsByEntityID    map[int]*Door
  equipmentMapping   map[*st.ServerClass]common.EquipmentElement

  console js.Value
  done    chan struct{}
}

func New() *DemoParser {
  return &DemoParser{
    console:            js.Global().Get("console"),
    done:               make(chan struct{}),
    hasNextFrame:       true,
    firing:             make(map[int64]time.Duration),
    chickensByEntityID: make(map[int]*Chicken),
    weaponsByEntityID:  make(map[int]*Weapon),
    doorsByEntityID:    make(map[int]*Door),
  }
}

// Start sets up all the callbacks and waits for the close signal
// to be sent from the browser.
func (dp *DemoParser) Start() {
  dp.console.Call("log", "Started")
  // Setup callbacks
  dp.setupInitMemCb()
  js.Global().Set("initMem", dp.initMemCb)

  dp.setupOnDemoLoadCb()
  js.Global().Set("loadDemo", dp.onDemoLoadCb)

  // dp.setupShutdownCb()
  // js.Global().Get("document").
  //   Call("getElementById", "close").
  //   Call("addEventListener", "click", dp.shutdownCb)

  dp.setupShutdownCb()
  js.Global().Set("closeGo", dp.shutdownCb)

  dp.setupGetHeader()
  js.Global().Set("getHeader", dp.getHeader)

  dp.setupDecodePositions()
  js.Global().Set("decodePositions", dp.decodePositions)

  dp.setupNextFrame()
  js.Global().Set("nextFrame", dp.nextFrame)

  dp.setupGetPositions()
  js.Global().Set("getPositions", dp.getPos)

  dp.setupGetChickenPositions()
  js.Global().Set("getChickenPositions", dp.getChickenPos)

  dp.setupGetWeapons()
  js.Global().Set("getWeaponPositions", dp.getWeaponPos)

  dp.setupInverseTranslate()
  js.Global().Set("inverseTranslate", dp.inverseTranslate)

  dp.setupTranslate()
  js.Global().Set("translate", dp.translate)

  dp.setupGetBombState()
  js.Global().Set("getBomb", dp.getBomb)

  <-dp.done
  dp.log("Shutting down app")
  dp.onDemoLoadCb.Release()
  dp.shutdownCb.Release()
  dp.getHeader.Release()
  dp.nextFrame.Release()
  dp.getPos.Release()
  dp.getChickenPos.Release()
  dp.getWeaponPos.Release()
}

// utility function to log a msg to the UI from inside a callback
func (dp *DemoParser) log(msg string) {
  dp.console.Call("log", msg)
  // js.Global().Get("document").
  //   Call("getElementById", "status").
  //   Set("innerText", msg)
}

func (dp *DemoParser) parseNextFrame() bool {
  var err error
  dp.hasNextFrame, err = dp.parser.ParseNextFrame()
  if !dp.hasNextFrame {
    dp.log("There are no more frames in the demo")
    return false
  }
  dp.checkError(err)
  return true
}

type ChickenMovementInfo struct {
  Position r3.Vector `bson:"Position" json:"Position"`
  ViewX    float32   `bson:"ViewX" json:"ViewX"`
  ViewY    float32   `bson:"ViewY" json:"ViewY"`
}

func NewChickenMovementInfo(entity *st.Entity) ChickenMovementInfo {
  return ChickenMovementInfo{
    Position: entity.Position(),
    ViewX:    float32(entity.FindPropertyI("m_angRotation").Value().VectorVal.X),
    ViewY:    float32(entity.FindPropertyI("m_angRotation").Value().VectorVal.Y),
  }
}

type WeaponMovementInfo struct {
  Position      r3.Vector `bson:"Position" json:"Position"`
  Name          string    `json:"WeaponName" bson:"WeaponName"`
  ServerClass   string    `json:"ServerClass" bson:"ServerClass"`
  ServerClassID int       `json:"ServerClassID" bson:"ServerClassID"`
  AngRotation   float64   `json:"AngRotation" bson:"AngRotation"`
}

func NewWeaponMovementInfo(pos r3.Vector, name string, serverClass string, id int) WeaponMovementInfo {
  return WeaponMovementInfo{
    Position:      pos,
    Name:          name,
    ServerClass:   serverClass,
    ServerClassID: id,
  }
}

type PlayerMovementInfo struct {
  Name      string      `bson:"SteamID" json:"SteamID"`
  Team      common.Team `bson:"Team" json:"Team"`
  Position  r3.Vector   `bson:"Position" json:"Position"`
  ViewX     float32     `bson:"ViewX" json:"ViewX"`
  ViewY     float32     `bson:"ViewY" json:"ViewY"`
  HP        int         `bson:"HP" json:"HP"`
  Armor     int         `bson:"Armor" json:"Armor"`
  Flash     int         `json:"Flash"`
  IsBlinded bool        `json:"IsBlinded"`
  IsFiring  bool        `json:"IsFiring"`
  IsAlive   bool        `json:"IsAlive"`
  HasBomb   bool        `json:"HasBomb"`
}

func (dp *DemoParser) NewPlayerMovementInfo(player *common.Player) PlayerMovementInfo {
  res := PlayerMovementInfo{
    player.Name,
    player.Team,
    player.Position,
    player.ViewDirectionX,
    player.ViewDirectionY,
    player.Hp,
    player.Armor,
    int(player.FlashDurationTimeRemaining().Nanoseconds() / 1000000000 / 50),
    player.IsBlinded(),
    false,
    player.IsAlive(),
    false,
  }
  if dp.parser.CurrentTime()-dp.firing[player.SteamID] < time.Millisecond*100 {
    res.IsFiring = true
  }
  if dp.parser.GameState().Bomb().Carrier == player {
    res.HasBomb = true
  }
  return res
}

func (dp *DemoParser) getPlayersPositions() []PlayerMovementInfo {
  PMIS := make([]PlayerMovementInfo, len(dp.parser.GameState().Participants().Playing()))
  for i, p := range dp.parser.GameState().Participants().Playing() {
    PMIS[i] = dp.NewPlayerMovementInfo(p)
  }
  return PMIS
}

func (dp *DemoParser) getChickensPositions() []ChickenMovementInfo {
  ChickenMovementInfos := make([]ChickenMovementInfo, 0, 8)
  for _, chicken := range dp.chickensByEntityID {
    if chicken.Entity != nil {
      ChickenMovementInfos = append(ChickenMovementInfos, NewChickenMovementInfo(chicken.Entity))
    }
  }
  // for _, entity := range dp.parser.GameState().Entities() {
  //   if entity.ServerClass().Name() == "CChicken" {
  //     ChickenMovementInfos = append(ChickenMovementInfos, NewChickenMovementInfo(entity))
  //   }
  // }
  return ChickenMovementInfos
}

func (dp *DemoParser) EquipmentMapping(modelIndex int, class *st.ServerClass) common.EquipmentElement {

  originalString := dp.parser.GetModelPreCache()[modelIndex]

  wepType := dp.parser.GetEquipmentMapping()[class]

  wepFix := func(defaultName, altName string, alt common.EquipmentElement) {
    // Check 'altName' first because otherwise the m4a1_s is recognized as m4a4
    if strings.Contains(originalString, altName) {
      wepType = alt
    } else if !strings.Contains(originalString, defaultName) {
      panic(fmt.Sprintf("Unknown weapon model %q", originalString))
    }
  }

  switch wepType {
  case common.EqP2000:
    wepFix("_pist_hkp2000", "_pist_223", common.EqUSP)
  case common.EqM4A4:
    wepFix("_rif_m4a1", "_rif_m4a1_s", common.EqM4A1)
  case common.EqP250:
    wepFix("_pist_p250", "_pist_cz_75", common.EqCZ)
  case common.EqDeagle:
    wepFix("_pist_deagle", "_pist_revolver", common.EqRevolver)
  case common.EqMP7:
    wepFix("_smg_mp7", "_smg_mp5sd", common.EqMP5)
  }

  return wepType
}

func (dp *DemoParser) getWeaponPositions() []WeaponMovementInfo {
  WeaponMovementInfos := make([]WeaponMovementInfo, 0)
  // weapons := dp.parser.Weapons()

  for _, weapon := range dp.weaponsByEntityID {
    if weapon.Entity != nil && weapon.Position.X != 0 {
      WeaponMovementInfos = append(WeaponMovementInfos,
        NewWeaponMovementInfo(
          weapon.Position,
          weapon.Type.String(),
          weapon.Entity.ServerClass().Name(),
          weapon.Entity.ServerClass().ID(),
        ))
    }
  }

  for _, door := range dp.doorsByEntityID {
    if door.Entity != nil {
      // sc := door.Entity.ServerClass()
      weaponMovementInfo := NewWeaponMovementInfo(
        door.Position,
        "Door",
        "CPropDoorRotating",
        0,
      )
      weaponMovementInfo.AngRotation = door.AngRotation
      WeaponMovementInfos = append(WeaponMovementInfos, weaponMovementInfo)
    }
  }

  // for _, entity := range dp.parser.GameState().Entities() {
  //   prop := entity.FindPropertyI("m_nModelIndex")
  //   if prop == nil {
  //     continue
  //   }
  //   sc := entity.ServerClass()
  //   modelIndex := prop.Value().IntVal
  //   weapon := dp.EquipmentMapping(modelIndex, sc) // weapons[entityID].Weapon.String()
  //   pos := entity.Position()
  //   if sc.Name() == "CPropDoorRotating" {
  //     if pos != (r3.Vector{}) {
  //       weaponMovementInfo := NewWeaponMovementInfo(pos, "Door", sc.Name(), sc.ID())
  //       weaponMovementInfo.AngRotation = entity.FindPropertyI("m_angRotation").Value().VectorVal.Y
  //       WeaponMovementInfos = append(WeaponMovementInfos, weaponMovementInfo)
  //     }
  //     continue
  //   }
    // else if sc.ID() == 53 {
    //   // } else if sc.Name() == "CEconEntity" {
    //   if pos != (r3.Vector{}) {
    //     weaponMovementInfo := NewWeaponMovementInfo(pos, weapon.String(), sc.Name(), sc.ID())
    //     WeaponMovementInfos = append(WeaponMovementInfos, weaponMovementInfo)
    //   }
    //   continue
    // }
    // else if sc.Name() == "CEconWearable" {
    //   if pos != (r3.Vector{}) {
    //     weaponMovementInfo := NewWeaponMovementInfo(pos, weapon.String(), sc.Name(), sc.ID())
    //     WeaponMovementInfos = append(WeaponMovementInfos, weaponMovementInfo)
    //   }
    //   continue
    // }
    // L:
    //   for _, bc := range sc.BaseClasses() {
    //     switch bc.Name() {
    //     case "CWeaponCSBase", "CBaseGrenade", "CBaseCSGrenade":
    //       if pos != (r3.Vector{}) {
    //         WeaponMovementInfos = append(WeaponMovementInfos,
    //           NewWeaponMovementInfo(pos, weapon.String(), sc.Name(), sc.ID()))
    //       }
    //       break L
    //       // default:
    //       //   if pos != (r3.Vector{}) {
    //       //     WeaponMovementInfos = append(WeaponMovementInfos,
    //       //       NewWeaponMovementInfo(pos, "UNKNOWN", sc.Name()))
    //       //   }
    //       //   break L
    //     }
    //   }
  // }

  // for entityID, weapon := range dp.parser.Weapons() {
  //   entity := dp.parser.GameState().Entities()[entityID]
  //   if entity == nil || entity.HasPosition() == false {
  //     continue
  //   }
  //   pos := dp.parser.GameState().Entities()[entityID].Position()
  //   if pos != (r3.Vector{}) {
  //     WeaponMovementInfos = append(WeaponMovementInfos, NewWeaponMovementInfo(pos, weapon.String()))
  //   }
  // }

  return WeaponMovementInfos
}

func (dp *DemoParser) getBombState() BombState {
  res := BombState{Defused: dp.bombDefused}
  if dp.parser.GameState().Bomb().Carrier == nil {
    res.X = dp.parser.GameState().Bomb().Position().X
    res.Y = dp.parser.GameState().Bomb().Position().Y
  }
  return res
}

func (dp *DemoParser) checkError(err error) {
  if err != nil {
    dp.log("err: " + err.Error())
    panic(err)
    // dp.console.Call("log", "err:", err.Error())
  }
}
