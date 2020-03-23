package demoparser

import (
  "github.com/golang/geo/r3"
  st "github.com/markus-wa/demoinfocs-golang/sendtables"
)

type Door struct {
  EntityID    int
  Position    r3.Vector
  Entity      *st.Entity
  AngRotation float64
  ModelIndex int
}

func (dp *DemoParser) bindNewDoor(doorEntity *st.Entity) {
  entityID := doorEntity.ID()
  door := dp.doorsByEntityID[entityID]

  if door == nil {
    door = &Door{
      EntityID: entityID,
      Entity:   doorEntity,
    }
    dp.doorsByEntityID[entityID] = door
  }

  doorEntity.OnDestroy(func() {
    delete(dp.doorsByEntityID, entityID)
    door.Entity = nil
  })

  doorEntity.BindProperty("m_nModelIndex", &door.ModelIndex, st.ValTypeInt)
  doorEntity.BindPosition(&door.Position)
  doorEntity.FindPropertyI("m_angRotation").OnUpdate(func(val st.PropertyValue) {
    door.AngRotation = val.VectorVal.Y
  })
}
