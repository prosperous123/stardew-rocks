package view

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"time"

	"github.com/disintegration/imaging"
	"github.com/nictuku/stardew-rocks/parser"
	"github.com/salviati/go-tmx/tmx"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func xnaRect(x0, y0, width, height int) image.Rectangle {
	return image.Rect(x0, y0, x0+width, y0+height)
}

func tileCoordinates(id int, tileWidth, tileHeight, tilemapWidth int) (x0, y0 int) {
	numColumns := tilemapWidth / tileWidth
	y0 = id / numColumns * tileHeight
	x0 = id % numColumns * tileWidth
	return x0, y0
}

var mask = image.NewUniform(color.Alpha{255})

var treeRects = map[int]image.Rectangle{
	0: xnaRect(32, 128, 16, 16),
	1: xnaRect(0, 128, 16, 16),
	2: xnaRect(16, 128, 16, 16),
	3: xnaRect(0, 96, 16, 32),
	4: xnaRect(0, 96, 16, 32),
}

func treeAsset(treeType int) string {
	return fmt.Sprintf("../TerrainFeatures/tree%d_spring.png", treeType) // TODO: other seasons
}

func grassOffset(grassType int) int {
	// TODO: other seasons
	switch grassType {
	case 1:
		return 0
	case 2:
		return 60
	case 3:
		return 80
	case 4:
		return 100
	}
	return 0
}

func grassRect(idx, grassType int) image.Rectangle {
	return xnaRect(idx, grassOffset(grassType), 15, 20)
}

func maybeFlip(flip bool, img image.Image, r image.Rectangle) image.Image {

	if flip {
		return imaging.FlipH(imaging.Crop(img, r))
	}
	return img
}

func drawTree(pm *parser.Map, item *parser.TerrainItem, img draw.Image) {
	m := pm.TMX

	if item.Value.TerrainFeature.Type != "Tree" {
		return
	}
	p := treeAsset(item.Value.TerrainFeature.TreeType)
	src, err := pm.FetchSource(p)
	if err != nil {
		log.Printf("Error fetching terrain asset %v: %v", p, err)
		return
	}
	stage := item.Value.TerrainFeature.GrowthStage
	if stage < 5 {
		sr, ok := treeRects[stage]
		if !ok {
			log.Printf("Unknown tree rect for %v", item.Value.TerrainFeature.GrowthStage)
			return
		}
		r := sr.Sub(sr.Min).Add(image.Point{item.Key.Vector2.X * m.TileWidth, item.Key.Vector2.Y * m.TileHeight})
		draw.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over)
	} else {
		{
			// shadow
			src, err := pm.FetchSource("../LooseSprites/Cursors.png")
			if err != nil {
				log.Printf("Error fetching terrain asset %v", err)
				return
			}
			sr := xnaRect(663, 1011, 41, 30)
			r := sr.Sub(sr.Min).Add(image.Point{item.Key.Vector2.X*m.TileWidth - m.TileWidth, // centralize
				item.Key.Vector2.Y*m.TileHeight - 0, // vertical centralize
			})
			draw.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over)
		}
		{
			// stump
			sr := xnaRect(32, 96, 16, 32)
			r := sr.Sub(sr.Min).Add(image.Point{item.Key.Vector2.X * m.TileWidth,
				item.Key.Vector2.Y*m.TileHeight - m.TileHeight, // stump offset
			})
			draw.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over)
		}
		{
			// tree
			sr := image.Rect(0, 0, 48, 96)
			r := sr.Sub(sr.Min).Add(image.Point{
				item.Key.Vector2.X*m.TileWidth - m.TileWidth, // centralize
				item.Key.Vector2.Y*m.TileHeight - 80,         // stump offset
			})
			draw.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over)
		}
	}
}

func drawGrass(pm *parser.Map, item *parser.TerrainItem, img draw.Image) {
	if item.Value.TerrainFeature.Type != "Grass" {
		return
	}
	whichWeed := rand.Perm(5)
	if item.Value.TerrainFeature.NumberOfWeeds >= len(whichWeed) {
		return
	}
	offsetWeeds := [4][]int{
		rand.Perm(5),
		rand.Perm(5),
		rand.Perm(5),
		rand.Perm(5),
	}
	flipWeed := [4]bool{
		rand.Float32() < 0.5,
		rand.Float32() < 0.5,
		rand.Float32() < 0.5,
		rand.Float32() < 0.5,
	}
	m := pm.TMX
	src, err := pm.FetchSource("../TerrainFeatures/grass.png")
	if err != nil {
		log.Fatalf("Error fetching image asset %v", err)
	}
	for i, weed := range whichWeed[0:item.Value.TerrainFeature.NumberOfWeeds] {
		fmt.Println("c", i, flipWeed[i])
		idx := 0
		if weed < len(whichWeed) {
			idx = whichWeed[weed] * 15
		}
		sr := grassRect(idx, item.Value.TerrainFeature.GrassType)
		fmt.Printf("%#v\n", item.Value.TerrainFeature)
		r := sr.Sub(sr.Min).Add(image.Point{item.Key.Vector2.X*m.TileWidth + (i%2*m.TileWidth/2 + offsetWeeds[2][i]) + 7,
			item.Key.Vector2.Y*m.TileHeight + i%2*m.TileHeight/2 + offsetWeeds[3][i] + 10,
		})
		if flipWeed[i] {
			sr = xnaRect(0, 0, m.TileWidth, m.TileHeight)
		}

		draw.DrawMask(img, r, maybeFlip(flipWeed[i], src, sr), sr.Min, mask, sr.Min, draw.Over)
	}

}

func drawTile(pm *parser.Map, tile *tmx.DecodedTile, img draw.Image, x, y int) {
	if tile.IsNil() {
		return
	}
	m := pm.TMX

	// Fetch tile from tileset.
	src, err := pm.FetchSource(tile.Tileset.Image.Source)
	if err != nil {
		log.Printf("Error fetching image asset %v: %v", tile.Tileset.Image.Source, err)
		return
	}
	srcBounds := src.Bounds()
	x0, y0 := tileCoordinates(int(tile.ID), m.TileWidth, m.TileHeight, srcBounds.Dx())
	sr := image.Rect(x0, y0, x0+m.TileHeight, y0+m.TileWidth)
	r := sr.Sub(sr.Min).Add(image.Point{x * m.TileWidth, y * m.TileHeight})
	// DrawMask with draw.Over and an alpha channel ensure the background is transparent.
	// Anyway, it works.
	draw.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over)
}

func WriteImage(pm *parser.Map, sg *parser.SaveGame, w io.Writer) {
	//fmt.Println("sg game locations", sg.Locations.GameLocations)
	var farm *parser.GameLocation
	for _, gameloc := range sg.Locations.GameLocations {
		if gameloc.Name == "Farm" {
			farm = &gameloc
			break
		}
	}

	if farm == nil {
		log.Printf("farm not found for %v", sg.Player.Name)
		return
	}

	// TODO: do not trust the user input. Don't hit files or slice indexes based on data.
	m := pm.TMX

	dirt := &image.Uniform{color.RGBA{0xEE, 0xAC, 0x24, 0xFF}}
	img := image.NewRGBA(image.Rect(0, 0, m.Width*m.TileWidth, m.Height*m.TileHeight))

	draw.Draw(img, img.Bounds(), dirt, image.ZP, draw.Src)

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			for _, layer := range m.Layers { // Layers are apparently ordered correctly.
				if layer.Name == "Paths" {
					continue // Looks ugly. Need some work to look pretty.
				}
				drawTile(pm, layer.DecodedTiles[y*m.Width+x], img, x, y)
			}
		}
	}

	{
		p := "../Buildings/houses.png"
		src, err := pm.FetchSource(p)
		if err != nil {
			log.Printf("Error fetching terrain asset %v: %v", p, err)
			panic(err)
		}
		{
			// house
			sr := image.Rect(0, 0, 160, 144)
			r := sr.Sub(sr.Min).Add(image.Point{930, 130})
			draw.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over)
		}
		{
			// greenHouse
			sr := image.Rect(160, 0, 272, 160)
			r := sr.Sub(sr.Min).Add(image.Point{400, 96})
			draw.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over)
		}
	}

	// objects are in Maps/springobjects.png
	{
		p := "../Maps/springobjects.png"
		src, err := pm.FetchSource(p)
		if err != nil {
			log.Printf("Error fetching terrain asset %v: %v", p, err)
			panic(err)
		}
		srcBounds := src.Bounds()

		for _, item := range farm.Objects.Items {
			x0, y0 := tileCoordinates(item.Value.Object.ParentSheetIndex, 16, 16, srcBounds.Dx())
			sr := image.Rect(x0, y0, x0+16, y0+16)
			r := sr.Sub(sr.Min).Add(image.Point{item.Key.Vector2.X * 16, item.Key.Vector2.Y * 16})
			draw.DrawMask(img, r, src, sr.Min, mask, sr.Min, draw.Over)
		}
	}

	{
		// Order items to be displayed based on their Y order. Items with a higher Y should be drawn last.
		items := make([][]*parser.TerrainItem, m.Height)
		for i := range farm.TerrainFeatures.Items {
			item := farm.TerrainFeatures.Items[i] // separate pointer for each item
			items[item.Y()] = append(items[item.Y()], &item)
		}

		{

			for _, row := range items {
				for _, item := range row {
					drawTree(pm, item, img)
					drawGrass(pm, item, img) //
				}
			}
		}

	}
	if err := png.Encode(w, img); err != nil {
		panic(err)
	}
}
