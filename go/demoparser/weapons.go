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

func (dp *DemoParser) bindNewWeapon(weaponEntity *st.Entity) {
  entityID := weaponEntity.ID()
  wp := dp.weaponsByEntityID[entityID]

  if wp == nil {
    wp = &Weapon{
      EntityID: entityID,
      Entity:   weaponEntity,
      Type:     dp.parser.Weapons()[entityID].Weapon,
    }
    dp.weaponsByEntityID[entityID] = wp
  }

  weaponEntity.OnDestroy(func() {
    delete(dp.weaponsByEntityID, entityID)
    wp.Entity = nil
  })

  weaponEntity.BindPosition(&wp.Position)
  weaponEntity.FindPropertyI("m_hOwnerEntity").OnUpdate(func(val st.PropertyValue) {
    wp.Owner = dp.parser.GameState().Participants().FindByHandle(val.IntVal)
  })
}
