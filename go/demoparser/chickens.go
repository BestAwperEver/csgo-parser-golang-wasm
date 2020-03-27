package demoparser

import (
  "github.com/golang/geo/r3"
  st "github.com/markus-wa/demoinfocs-golang/sendtables"
)

type Chicken struct {
  EntityID  int
  Entity    *st.Entity
  Position  r3.Vector
  ViewAngle float32
}

func (dp *DemoParser) bindChicken(chickenEntity *st.Entity) {
  entityID := chickenEntity.ID()
  ch := dp.chickensByEntityID[entityID]

  if ch == nil {
    ch = &Chicken{
      EntityID: entityID,
      Entity:   chickenEntity,
    }
    dp.chickensByEntityID[entityID] = ch
  }

  chickenEntity.OnDestroy(func() {
    delete(dp.chickensByEntityID, entityID)
    ch.Entity = nil
  })

  chickenEntity.BindPosition(&ch.Position)
  chickenEntity.FindPropertyI("m_angRotation").OnUpdate(func(val st.PropertyValue) {
    ch.ViewAngle = float32(val.VectorVal.Y)
  })
}
