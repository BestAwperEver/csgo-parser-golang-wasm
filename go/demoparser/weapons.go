package demoparser

import (
  "github.com/golang/geo/r3"
  "github.com/markus-wa/demoinfocs-golang/common"
  st "github.com/markus-wa/demoinfocs-golang/sendtables"
)

type Weapon struct {
  EntityID int
  Position r3.Vector
  Entity   *st.Entity
  Owner    *common.Player
  Type     common.EquipmentElement
}

func (dp *DemoParser) bindWeapon(weaponEntity *st.Entity, weaponType common.EquipmentElement) {
  entityID := weaponEntity.ID()
  wp := dp.weaponsByEntityID[entityID]

  if wp == nil {
    wp = &Weapon{
      EntityID: entityID,
      Entity:   weaponEntity,
      Type:     weaponType,
    }
    dp.weaponsByEntityID[entityID] = wp

    // if wp.Type == common.EqUnknown {
    //   dp.dbgLog(weaponEntity.ServerClass().Name() + " has unknown type",
    //     dbgPrintUnknownWeapons)
    // }
  }

  weaponEntity.BindPosition(&wp.Position)
  weaponEntity.FindPropertyI("m_hOwnerEntity").OnUpdate(func(val st.PropertyValue) {
    wp.Owner = dp.parser.GameState().Participants().FindByHandle(val.IntVal)
  })

  weaponEntity.OnDestroy(func() {
    delete(dp.weaponsByEntityID, entityID)
    wp.Entity = nil
  })
}
