package scan

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/image/webp"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"gen-library/backend/db"
	"gen-library/backend/util"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

// ScanFolder walks the root directory, importing any new images it finds.
// It returns the number of files added or updated.
func ScanFolder(gdb *gorm.DB, root string) (int, error) {
	var count int
	// Ensure root is absolute for filepath.Rel to behave predictably
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return 0, err
	}

	err = gdb.Transaction(func(tx *gorm.DB) error {
		walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(d.Name()))
			switch ext {
			case ".png", ".jpg", ".jpeg", ".webp":
			default:
				return nil
			}

			added, err := processFile(tx, absRoot, path, ext)
			if err != nil {
				// Log and continue scanning
				log.Printf("scan: %v", err)
				return nil
			}
			if added {
				count++
			}
			return nil
		})
		if walkErr != nil {
			return walkErr
		}
		return nil
	})

	return count, err
}

// ScanFile imports or updates a single image file without walking directories.
// It returns true if a row was inserted or updated.
func ScanFile(gdb *gorm.DB, root, path string) (bool, error) {
	// Ensure root and path are absolute for consistent behavior
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
	default:
		return false, nil
	}

	var added bool
	err = gdb.Transaction(func(tx *gorm.DB) error {
		var err error
		added, err = processFile(tx, absRoot, absPath, ext)
		return err
	})
	if err != nil {
		return false, err
	}
	return added, nil
}

// processFile handles a single image file. It returns true if a DB row was
// inserted or updated.
func processFile(tx *gorm.DB, root, path, ext string) (bool, error) {
	// Compute hash first to detect existing files regardless of path
	sha, err := util.HashFileSHA256(path)
	if err != nil {
		return false, err
	}

	// Stat for size and mtime
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	size := fi.Size()
	mtime := fi.ModTime()

	// Determine dimensions
	width, height := getImageDimensions(path, ext)

	// Extract metadata
	metaMap, err := extractMetadata(path, ext)
	if err != nil {
		// non-fatal, continue with what we have
		log.Printf("metadata error %s: %v", path, err)
		metaMap = map[string]string{}
	}

	// Merge any JSON blobs into the metadata map
	mergeJSONMeta(metaMap)

	// Parse generation parameters if present
	normalizeParameters(metaMap)

	// Extract model hash, loras, and embeddings from sui_models
	modelHash, loras, embeds := extractModels(metaMap)
	assocLoras := []*db.Lora{}
	for _, lr := range loras {
		name := lr.Name
		hash := ""
		if lr.Hash != nil {
			hash = *lr.Hash
		}
		var l db.Lora
		if err := tx.Where("name = ?", name).First(&l).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if hash != "" {
					var existing db.Lora
					if err := tx.Where("hash = ?", hash).First(&existing).Error; err == nil {
						log.Printf("hash conflict for lora %s; using existing %s", name, existing.Name)
						l = existing
					} else {
						l = db.Lora{Name: name}
						if err := tx.Create(&l).Error; err != nil {
							return false, err
						}
					}
				} else {
					l = db.Lora{Name: name}
					if err := tx.Create(&l).Error; err != nil {
						return false, err
					}
				}
			} else {
				return false, err
			}
		}
		if hash != "" {
			if l.Hash == nil {
				l.Hash = &hash
				if err := tx.Save(&l).Error; err != nil {
					return false, err
				}
			} else if *l.Hash != hash {
				var existing db.Lora
				if err := tx.Where("hash = ?", hash).First(&existing).Error; err == nil && existing.ID != l.ID {
					log.Printf("hash conflict for lora %s; using existing %s", name, existing.Name)
					l = existing
				} else if errors.Is(err, gorm.ErrRecordNotFound) {
					l.Hash = &hash
					if err := tx.Save(&l).Error; err != nil {
						return false, err
					}
				} else if err != nil {
					return false, err
				}
			}
		}
		assocLoras = append(assocLoras, &l)
	}
	if modelHash != "" && metaMap["model hash"] == "" {
		metaMap["model hash"] = modelHash
	}

	// Prepare Image model
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	rel = filepath.ToSlash(rel)

	img := db.Image{
		Path:       rel,
		FileName:   dName(path),
		Ext:        strings.TrimPrefix(ext, "."),
		SizeBytes:  size,
		SHA256:     sha,
		NSFW:       checkNSFW(metaMap, dName(path)),
		Embeddings: embeds,
	}
	if width > 0 {
		img.Width = &width
	}
	if height > 0 {
		img.Height = &height
	}
	if !mtime.IsZero() {
		ct := mtime
		img.CreatedTime = &ct
	}

	// Normalized fields from metaMap
	if v, ok := metaMap["sourceapp"]; ok {
		img.SourceApp = &v
	}
	if v, ok := metaMap["model"]; ok {
		img.ModelName = &v
	}
	if v, ok := metaMap["model hash"]; ok {
		img.ModelHash = &v
	}
	if v, ok := metaMap["prompt"]; ok {
		img.Prompt = &v
	}
	if v, ok := metaMap["negative prompt"]; ok {
		img.NegativePrompt = &v
	}
	if v, ok := metaMap["sampler"]; ok {
		img.Sampler = &v
	}
	if v, ok := metaMap["steps"]; ok {
		if iv, err := strconv.Atoi(v); err == nil {
			img.Steps = &iv
		}
	}
	if v, ok := metaMap["cfg scale"]; ok {
		if fv, err := strconv.ParseFloat(v, 64); err == nil {
			img.CFGScale = &fv
		}
	}
	if v, ok := metaMap["seed"]; ok {
		img.Seed = &v
	}
	if v, ok := metaMap["scheduler"]; ok {
		img.Scheduler = &v
	}
	if v, ok := metaMap["clip skip"]; ok {
		if iv, err := strconv.Atoi(v); err == nil {
			img.ClipSkip = &iv
		}
	}
	if v, ok := metaMap["variationseed"]; ok {
		if iv, err := strconv.Atoi(v); err == nil {
			img.VariationSeed = &iv
		}
	}
	if v, ok := metaMap["variationseedstrength"]; ok {
		if fv, err := strconv.ParseFloat(v, 64); err == nil {
			img.VariationSeedStrength = &fv
		}
	}
	if v, ok := metaMap["aspectratio"]; ok {
		img.AspectRatio = &v
	}
	if v, ok := metaMap["refinercontrolpercentage"]; ok {
		if fv, err := strconv.ParseFloat(v, 64); err == nil {
			img.RefinerControlPercentage = &fv
		}
	}
	if v, ok := metaMap["refinerupscale"]; ok {
		if fv, err := strconv.ParseFloat(v, 64); err == nil {
			img.RefinerUpscale = &fv
		}
	}
	if v, ok := metaMap["refinerupscalemethod"]; ok {
		img.RefinerUpscaleMethod = &v
	}

	// Store raw metadata JSON
	if len(metaMap) > 0 {
		if raw, err := json.Marshal(metaMap); err == nil {
			img.RawMetadata = datatypes.JSON(raw)
		}
	}

	// Check if exists by SHA without triggering a "record not found" log
	var existing db.Image
	res := tx.Where("sha256 = ?", sha).Limit(1).Find(&existing)
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected > 0 {
		// Already exists - maybe moved
		if existing.Path != rel {
			upd := map[string]any{"path": rel, "file_name": dName(path)}
			if err := tx.Model(&db.Image{}).Where("id = ?", existing.ID).Updates(upd).Error; err != nil {
				return false, err
			}
			return true, nil
		}
		return false, nil
	}

	if err := tx.Create(&img).Error; err != nil {
		return false, err
	}
	if len(assocLoras) > 0 {
		if err := tx.Model(&img).Association("Loras").Append(assocLoras); err != nil {
			return false, err
		}
	}
	return true, nil
}

// getImageDimensions returns width and height for supported formats.
func getImageDimensions(path, ext string) (int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	switch ext {
	case ".webp":
		cfg, err := webp.DecodeConfig(f)
		if err != nil {
			return 0, 0
		}
		return cfg.Width, cfg.Height
	default:
		cfg, _, err := image.DecodeConfig(f)
		if err != nil {
			return 0, 0
		}
		return cfg.Width, cfg.Height
	}
}

// extractMetadata gathers textual metadata from the image depending on format.
// Keys are stored in lowercase.
func extractMetadata(path, ext string) (map[string]string, error) {
	switch ext {
	case ".png":
		return parsePNGChunks(path)
	case ".jpg", ".jpeg":
		return parseJPEG(path)
	case ".webp":
		return parseWebP(path)
	default:
		return map[string]string{}, nil
	}
}

// parsePNGChunks extracts tEXt and iTXt chunks.
func parsePNGChunks(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	meta := map[string]string{}
	// Skip signature
	if _, err := f.Seek(8, io.SeekStart); err != nil {
		return nil, err
	}
	for {
		var lengthBuf [4]byte
		if _, err := io.ReadFull(f, lengthBuf[:]); err != nil {
			if err == io.EOF {
				break
			}
			return meta, err
		}
		length := int64(uint32(lengthBuf[0])<<24 | uint32(lengthBuf[1])<<16 | uint32(lengthBuf[2])<<8 | uint32(lengthBuf[3]))
		var typ [4]byte
		if _, err := io.ReadFull(f, typ[:]); err != nil {
			return meta, err
		}
		data := make([]byte, length)
		if _, err := io.ReadFull(f, data); err != nil {
			return meta, err
		}
		// skip CRC
		if _, err := f.Seek(4, io.SeekCurrent); err != nil {
			return meta, err
		}

		key := string(typ[:])
		switch key {
		case "tEXt":
			parts := bytes.SplitN(data, []byte{0}, 2)
			if len(parts) == 2 {
				meta[strings.ToLower(string(parts[0]))] = string(parts[1])
			}
		case "iTXt":
			// iTXt: keyword\0 compressionFlag compressionMethod languageTag\0 translatedKeyword\0 text
			parts := bytes.SplitN(data, []byte{0}, 5)
			if len(parts) >= 5 {
				keyword := string(parts[0])
				compFlag := parts[1]
				text := parts[4]
				if len(compFlag) > 0 && compFlag[0] == 1 {
					zr, err := zlib.NewReader(bytes.NewReader(text))
					if err == nil {
						decompressed, _ := io.ReadAll(zr)
						zr.Close()
						text = decompressed
					}
				}
				meta[strings.ToLower(keyword)] = string(text)
			}
		}
		if key == "IEND" {
			break
		}
	}
	return meta, nil
}

// parseJPEG extracts EXIF and XMP data from JPEG files.
func parseJPEG(path string) (map[string]string, error) {
	meta := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return meta, err
	}
	defer f.Close()

	if x, err := exif.Decode(f); err == nil {
		x.Walk(exifWalker(meta))
	}

	// Read entire file to look for XMP packet
	if data, err := os.ReadFile(path); err == nil {
		extractXMP(string(data), meta)
	}
	return meta, nil
}

// parseWebP extracts metadata from WebP files. WebP stores EXIF/XMP in RIFF
// chunks; for simplicity we scan the file for an embedded XMP packet and a
// plaintext "parameters" string.
func parseWebP(path string) (map[string]string, error) {
	meta := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return meta, err
	}
	extractXMP(string(data), meta)
	// Many generators also embed the parameters string verbatim
	if bytes.Contains(data, []byte("parameters")) {
		s := string(data)
		idx := strings.Index(s, "parameters")
		meta["parameters"] = s[idx:]
	}
	return meta, nil
}

// exifWalker collects all tags into the provided map
type exifWalker map[string]string

func (w exifWalker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	if s, err := tag.StringVal(); err == nil {
		w[strings.ToLower(string(name))] = s
	}
	return nil
}

// extractXMP finds an XMP packet in the provided data and parses known fields.
func extractXMP(data string, meta map[string]string) {
	start := strings.Index(strings.ToLower(data), "<x:xmpmeta")
	if start == -1 {
		return
	}
	end := strings.Index(strings.ToLower(data[start:]), "</x:xmpmeta>")
	if end == -1 {
		return
	}
	xmp := data[start : start+end+12]

	// Capture tags or attributes
	fetch := func(tag string) string {
		re := regexp.MustCompile("(?is)<[^>]*" + tag + "[^>]*>(.*?)</[^>]*" + tag + "[^>]*>")
		if m := re.FindStringSubmatch(xmp); len(m) >= 2 {
			return strings.TrimSpace(m[1])
		}
		reAttr := regexp.MustCompile("(?is)" + tag + "=\"([^\"]*)\"")
		if m := reAttr.FindStringSubmatch(xmp); len(m) >= 2 {
			return strings.TrimSpace(m[1])
		}
		return ""
	}
	candidates := map[string]string{
		"prompt":          "prompt",
		"negativeprompt":  "negative prompt",
		"negative_prompt": "negative prompt",
		"sampler":         "sampler",
		"steps":           "steps",
		"cfgscale":        "cfg scale",
		"cfg_scale":       "cfg scale",
		"seed":            "seed",
		"model":           "model",
		"modelhash":       "model hash",
		"model_hash":      "model hash",
		"scheduler":       "scheduler",
		"clipskip":        "clip skip",
		"clip_skip":       "clip skip",
		"parameters":      "parameters",
		"sourceapp":       "sourceapp",
	}
	for tag, key := range candidates {
		if v := fetch(tag); v != "" {
			meta[key] = v
		}
	}
}

// mergeJSONMeta parses any metadata values that contain JSON objects
// and merges their key/value pairs back into the main metadata map.
func mergeJSONMeta(meta map[string]string) {
	// Alias certain JSON keys to match the normalized keys expected elsewhere.
	aliases := map[string]string{
		"negativeprompt":  "negative prompt",
		"negative_prompt": "negative prompt",
		"cfgscale":        "cfg scale",
		"cfg_scale":       "cfg scale",
		"modelhash":       "model hash",
		"model_hash":      "model hash",
		"clipskip":        "clip skip",
		"clip_skip":       "clip skip",
	}

	// Keys to exclude entirely from the metadata map.
	exclude := map[string]struct{}{
		"sui_extra_data": {},
		"images":         {},
	}

	// Keep parsing as long as we keep discovering new JSON blobs.
	for {
		changed := false
		for k, v := range meta {
			if _, skip := exclude[k]; skip {
				delete(meta, k)
				changed = true
				continue
			}
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			var obj map[string]any
			if json.Unmarshal([]byte(v), &obj) == nil {
				for kk, val := range obj {
					key := strings.ToLower(kk)
					if _, skip := exclude[key]; skip {
						continue
					}
					if alias, ok := aliases[key]; ok {
						key = alias
					}
					var sval string
					switch t := val.(type) {
					case string:
						sval = t
					case float64:
						sval = strconv.FormatFloat(t, 'f', -1, 64)
					case bool:
						sval = strconv.FormatBool(t)
					default:
						if b, err := json.Marshal(t); err == nil {
							sval = string(b)
						}
					}
					if sval == "" {
						continue
					}
					if cur, ok := meta[key]; !ok || cur != sval {
						meta[key] = sval
						changed = true
					}
				}
			}
		}
		if !changed {
			break
		}
	}

	// Set Source App for SwarmUI metadata
	if _, ok := meta["swarm_version"]; ok {
		if cur, ok := meta["sourceapp"]; !ok || cur == "" {
			meta["sourceapp"] = "SwarmUI"
		}
	}
}

// normalizeParameters parses the Stable Diffusion parameter string if present
// and merges extracted key/value pairs back into the meta map.
func normalizeParameters(meta map[string]string) {
	param, ok := meta["parameters"]
	if !ok {
		return
	}

	// If the parameters string is actually JSON, skip parsing to avoid
	// clobbering values like the prompt.
	if json.Valid([]byte(strings.TrimSpace(param))) {
		return
	}

	lines := strings.Split(param, "\n")
	if len(lines) == 0 {
		return
	}
	if _, ok := meta["prompt"]; !ok {
		meta["prompt"] = strings.TrimSpace(lines[0])
	}
	var idx int
	for i, l := range lines[1:] {
		if strings.HasPrefix(strings.ToLower(l), "negative prompt:") {
			meta["negative prompt"] = strings.TrimSpace(l[len("Negative prompt:"):])
			idx = i + 2
			break
		}
	}
	if idx < len(lines) {
		paramsLine := lines[idx]
		parts := strings.Split(paramsLine, ",")
		for _, p := range parts {
			kv := strings.SplitN(p, ":", 2)
			if len(kv) != 2 {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(kv[0]))
			val := strings.TrimSpace(kv[1])
			meta[key] = val
		}
	}
}

// extractModels parses the sui_models JSON array and returns the model hash,
// any loras, and any embeddings used by the generation.
func extractModels(meta map[string]string) (string, []db.Lora, []db.Embedding) {
	s, ok := meta["sui_models"]
	if !ok {
		return "", nil, nil
	}
	var entries []struct {
		Name  string `json:"name"`
		Param string `json:"param"`
		Hash  string `json:"hash"`
	}
	if err := json.Unmarshal([]byte(s), &entries); err != nil {
		return "", nil, nil
	}
	var modelHash string
	loras := []db.Lora{}
	embeds := []db.Embedding{}
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name, ".safetensors")
		if strings.HasPrefix(name, "LyCORIS/") {
			name = strings.TrimPrefix(name, "LyCORIS/")
		}
		switch e.Param {
		case "model":
			if e.Hash != "" {
				modelHash = e.Hash
			}
			if name != "" && meta["model"] == "" {
				meta["model"] = name
			}
		case "loras":
			if name != "" || e.Hash != "" {
				var hptr *string
				if e.Hash != "" {
					h := e.Hash
					hptr = &h
				}
				loras = append(loras, db.Lora{Name: name, Hash: hptr})
			}
		case "used_embeddings":
			if name != "" || e.Hash != "" {
				embeds = append(embeds, db.Embedding{Name: name, Hash: e.Hash})
			}
		}
	}
	return modelHash, loras, embeds
}

// checkNSFW applies a simple keyword heuristic on prompts and file name.
func checkNSFW(meta map[string]string, filename string) bool {
	keywords := []string{"nsfw", "nude", "naked", "sex", "fuck", "topless", "bottomless", "pubic", "cum", "porn", "erotic", "pussy", "cock", "penis", "vagina"}
	text := strings.ToLower(filename)
	if p, ok := meta["prompt"]; ok {
		text += " " + strings.ToLower(p)
	}
	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

// dName returns base name of path
func dName(path string) string { return filepath.Base(path) }
