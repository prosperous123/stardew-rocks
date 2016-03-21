package view

import (
	"fmt"
	"image"
	"image/draw"
	"log"

	"github.com/nictuku/stardew-rocks/parser"
)

func drawBuilding(pm *parser.Map, building *parser.Building, img draw.Image) {
	if building.Type == "" {
		return
	}
	m := pm.TMX
	sourcePath := fmt.Sprintf("../Buildings/%v.png", building.BuildingType)
	src, err := pm.FetchSource(sourcePath)
	if err != nil {
		log.Printf("Error fetching asset %v: %v", sourcePath, err)
		return
	}
	switch building.Type { // Also works for building.BuildingType = Deluxe Coop
	case "Coop":
		sr := xnaRect(16, 112, 16, 16)
		r := sr.Sub(sr.Min).Add(image.Point{
			(building.TileX + building.AnimalDoor.X) * m.TileWidth,
			(building.TileY + building.AnimalDoor.Y) * m.TileHeight,
		})
		sb.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over, objectLayer)

		// Animal door
		sr = xnaRect(0, 112, 16, 16)
		r = sr.Sub(sr.Min).Add(image.Point{
			((building.TileX + building.AnimalDoor.X) * m.TileWidth) + 16, // TODO: Open door
			(building.TileY + building.AnimalDoor.Y) * m.TileHeight,
		})
		sb.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over, objectLayer)

		// Coop
		sr = xnaRect(0, 0, 96, 112)
		dp := image.Point{
			(building.TileX * m.TileWidth),
			building.TileY*m.TileHeight + building.TilesHigh*m.TileHeight,
		}
		sb.DrawMask(img, topLeftAlign(sr, dp), src, sr.Min, mask, sr.Min, draw.Over, houseLayer)

	default:
		return
	}
}
