package demoparser

import (
  "github.com/markus-wa/demoinfocs-golang/common"
  "github.com/markus-wa/demoinfocs-golang/events"
  st "github.com/markus-wa/demoinfocs-golang/sendtables"
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

func (dp *DemoParser) handleDataTablesParsed(events.DataTablesParsed) {
  for _, sc := range dp.parser.ServerClasses() {
    switch sc.Name() {
    case "CEconEntity":
        sc.OnEntityCreated(func(weaponEntity *st.Entity) {
          if wep, ok := dp.parser.GameState().Weapons()[weaponEntity.ID()]; ok {
            dp.bindWeapon(weaponEntity, wep.Weapon)
          } else {
            dp.bindWeapon(weaponEntity, common.EqDefuseKit)
          }
        })
    case "CChicken":
      sc.OnEntityCreated(dp.bindChicken)
    case "CPropDoorRotating":
      sc.OnEntityCreated(dp.bindDoor)
    // case "CCSPlayer":
    //   sc.OnEntityCreated(dp.bindPlayer)
    // case "CCSGameRulesProxy":
    //   sc.OnEntityCreated(dp.bindGameRules)
    default:
      for _, bc := range sc.BaseClasses() {
        switch bc.Name() {
        // case "CBaseGrenade": // Grenade that has been thrown by player.
        case "CWeaponCSBase", "CBaseCSGrenade":
          sc.OnEntityCreated(func(weaponEntity *st.Entity) {
            if wep, ok := dp.parser.GameState().Weapons()[weaponEntity.ID()]; ok {
              dp.bindWeapon(weaponEntity, wep.Weapon)
            } else {
              dp.bindWeapon(weaponEntity, common.EqUnknown)
            }
          })
        }
      }
    }
  }
}
