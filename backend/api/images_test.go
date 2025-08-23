package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	api "gen-library/backend/api"
	"gen-library/backend/db"
)

// setupRouter initializes an in-memory database, seeds test data and returns a gin.Engine.
func setupRouter(t *testing.T) (*gin.Engine, bool) {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.ApplyMigrations(gdb))

	// Seed tags
	tagAnimal := db.Tag{Name: "animal"}
	tagCat := db.Tag{Name: "cat"}
	tagDog := db.Tag{Name: "dog"}
	tagFlower := db.Tag{Name: "flower"}
	require.NoError(t, gdb.Create(&tagAnimal).Error)
	require.NoError(t, gdb.Create(&tagCat).Error)
	require.NoError(t, gdb.Create(&tagDog).Error)
	require.NoError(t, gdb.Create(&tagFlower).Error)

	// Seed images
	imgCat := db.Image{Path: "cat.jpg", FileName: "cat", Ext: "jpg", SizeBytes: 1, SHA256: "sha1", NSFW: false, Tags: []*db.Tag{&tagAnimal, &tagCat}}
	imgDog := db.Image{Path: "dog.jpg", FileName: "dog", Ext: "jpg", SizeBytes: 1, SHA256: "sha2", NSFW: true, Tags: []*db.Tag{&tagAnimal, &tagDog}}
	imgSun := db.Image{Path: "sunflower.jpg", FileName: "sunflower", Ext: "jpg", SizeBytes: 1, SHA256: "sha3", NSFW: false, Tags: []*db.Tag{&tagFlower}}
	require.NoError(t, gdb.Create(&imgCat).Error)
	require.NoError(t, gdb.Create(&imgDog).Error)
	require.NoError(t, gdb.Create(&imgSun).Error)

	// Determine if FTS is available
	hasFTS := gdb.Exec("SELECT 1 FROM images_fts LIMIT 1").Error == nil

	gin.SetMode(gin.TestMode)
	r := gin.New()
	api.RegisterRoutes(r, gdb)
	return r, hasFTS
}

func getFileNames(t *testing.T, r *gin.Engine, url string) []string {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []struct {
			FileName string `json:"fileName"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	names := make([]string, 0, len(resp.Items))
	for _, it := range resp.Items {
		names = append(names, it.FileName)
	}
	return names
}

func TestListImagesFilters(t *testing.T) {
	r, hasFTS := setupRouter(t)

	t.Run("nsfw modes", func(t *testing.T) {
		names := getFileNames(t, r, "/api/images")
		require.ElementsMatch(t, []string{"cat", "sunflower"}, names)

		names = getFileNames(t, r, "/api/images?nsfw=show")
		require.ElementsMatch(t, []string{"cat", "dog", "sunflower"}, names)

		names = getFileNames(t, r, "/api/images?nsfw=only")
		require.ElementsMatch(t, []string{"dog"}, names)
	})

	t.Run("fts wildcard search", func(t *testing.T) {
		if !hasFTS {
			t.Skip("fts5 not available")
		}
		names := getFileNames(t, r, "/api/images?q=sunf")
		require.ElementsMatch(t, []string{"sunflower"}, names)
	})

	t.Run("tag requirements", func(t *testing.T) {
		names := getFileNames(t, r, "/api/images?tags=animal&nsfw=show")
		require.ElementsMatch(t, []string{"cat", "dog"}, names)

		names = getFileNames(t, r, "/api/images?tags=animal,cat&nsfw=show")
		require.ElementsMatch(t, []string{"cat"}, names)

		names = getFileNames(t, r, "/api/images?tags=animal,flower&nsfw=show")
		require.Empty(t, names)
	})
}
