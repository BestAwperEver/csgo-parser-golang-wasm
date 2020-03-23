package demoparser

import (
  "github.com/markus-wa/demoinfocs-golang/events"
)

// func (dp *DemoParser) chickenEntityHandler(events.DataTablesParsed) {
//   chickens := dp.parser.ServerClasses().FindByName("CChicken")
//   if chickens != nil {
//     chickens.OnEntityCreated(dp.bindNewChicken)
//   }
// }
//
// func (dp *DemoParser) doorEntityHandler(events.DataTablesParsed) {
//   doors := dp.parser.ServerClasses().FindByName("CPropDoorRotating")
//   if doors != nil {
//     doors.OnEntityCreated(dp.bindNewDoor)
//   }
// }

func (dp *DemoParser) entityHandler(events.DataTablesParsed) {

  // weaponClasses := []string{
  //   "CWeaponCSBase", "CBaseGrenade", "CBaseCSGrenade",
  // }

  for _, sc := range dp.parser.ServerClasses() {
    if sc.Name() == "CEconEntity" {
      sc.OnEntityCreated(dp.bindNewWeapon)
    } else if sc.Name() == "CChicken" {
      sc.OnEntityCreated(dp.bindNewChicken)
    } else if sc.Name() == "CPropDoorRotating" {
      sc.OnEntityCreated(dp.bindNewDoor)
    } else {
      for _, bc := range sc.BaseClasses() {
        switch bc.Name() {
        case "CWeaponCSBase", "CBaseGrenade", "CBaseCSGrenade":
          // sc2 := sc // Local copy for loop
          sc.OnEntityCreated(dp.bindNewWeapon)
          // case "CBaseGrenade": // Grenade that has been thrown by player.
          //   sc.OnEntityCreated(p.bindGrenadeProjectiles)
          // case "CBaseCSGrenade":
          //   // @micvbang TODO: handle grenades dropped by dead player.
          //   // Grenades that were dropped by a dead player (and can be picked up by other players).
        }
      }
    }
  }

  // for _, weaponClass := range weaponClasses {
  //   weapons := dp.parser.ServerClasses().FindByName(weaponClass)
  //   if weapons != nil {
  //     weapons.OnEntityCreated(func(weaponEntity *st.Entity) {
  //       dp.bindNewWeapon(weaponEntity)
  //     })
  //   }
  // }
}
