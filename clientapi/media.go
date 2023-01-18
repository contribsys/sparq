package clientapi

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/util/blurhash"
	"github.com/contribsys/sparq/web"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// TODO: focus
func postMediaHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			httpError(w, errors.New("POST only"), http.StatusBadRequest)
			return
		}
		err := r.ParseMultipartForm(32 << 20) // 32 MB
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}

		ctx := web.Ctx(r)
		aid := ctx.CurrentUserID
		if aid == web.Anonymous {
			httpError(w, errors.New("Unauthorized"), http.StatusUnauthorized)
			return
		}

		ffile, _, err := r.FormFile("file")
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		media := &model.TootMedia{
			AccountId:   aid,
			Description: r.Form.Get("description"),
		}

		// 0. Save original media to disk
		origfile, err := os.CreateTemp("", "orig-*")
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		_, err = io.Copy(origfile, ffile)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		defer os.Remove(origfile.Name())

		// Media normalization
		// 1. Convert original to optimized JPG
		newfile, err := os.CreateTemp("", "full-*.jpg")
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		defer os.Remove(newfile.Name())
		_, err = compact(origfile.Name(), newfile)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		media.MimeType = "image/jpeg"
		fimg, _, err := image.Decode(newfile)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		// 2. Generate thumbnail
		newthumb, err := os.CreateTemp("", "thumb-*.jpg")
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		defer os.Remove(newthumb.Name())
		_, err = thumb(newfile.Name(), newthumb)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		media.ThumbMimeType = "image/jpeg"

		// 3. Grab metadata
		timg, _, err := image.Decode(newthumb)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		hash, err := blurhash.Encode(4, 3, timg)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		media.Blurhash = hash
		media.Meta = fmt.Sprintf(`{"original":{"width":%d,"height":%d},"small":{"width":%d,"height":%d}}`,
			fimg.Bounds().Dx(), fimg.Bounds().Dy(),
			timg.Bounds().Dx(), timg.Bounds().Dy())

		now := time.Now().UTC()
		dir := fmt.Sprintf("%s/%d/%d/%d", s.MediaRoot(), now.Year(), now.Month(), now.Day())
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		// 4. Save to DB
		result, err := s.DB().ExecContext(r.Context(), `
			insert into toot_medias (accountid, mimetype, thumbmimetype, description, blurhash, createdat, meta)
			 values (?, ?, ?, ?, ?, ?, ?)`,
			media.AccountId, media.MimeType, media.ThumbMimeType, media.Description, media.Blurhash, now, media.Meta)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		mid, err := result.LastInsertId()
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		full, err := os.Create(fmt.Sprintf("%s/full-%d.jpg", dir, mid))
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(full, newfile)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		thumb, err := os.Create(fmt.Sprintf("%s/thumb-%d.jpg", dir, mid))
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(thumb, newthumb)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		media.Id = uint64(mid)
		media.CreatedAt = now
		media.Path = media.DiskPath("full")
		media.ThumbPath = media.DiskPath("thumb")

		// 5. Push tmp files to URLs
		_, err = s.DB().ExecContext(r.Context(), `
			update toot_medias set path = ?, thumbpath = ? where id = ?`, media.Path, media.ThumbPath, mid)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(toAttachmentMap(media))
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
	}
}

// -format '{"height": %h, "width": %w}'
func compact(filename string, newfile *os.File) (string, error) {
	return run("convert", "-quality", "60", "-strip", filename, newfile.Name())
}

func thumb(filename string, newfile *os.File) (string, error) {
	return run("convert", "-thumbnail", "100", filename, newfile.Name())
}

// func file(file *os.File) (string, error) {
// return run("file", file.Name())
// }

func run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to process media: '%s' %v", out, args))
	}
	return string(out), nil
}

func getMediaAttachmentHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := web.Ctx(r)
		aid := ctx.CurrentUserID
		if aid == web.Anonymous {
			httpError(w, errors.New("Unauthorized"), http.StatusUnauthorized)
			return
		}

		mid := mux.Vars(r)["id"]
		var media model.TootMedia

		err := s.DB().Get(&media, "select * from toot_medias where id = ?", mid)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(toAttachmentMap(&media))
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
	}
}

func toAttachmentMap(media *model.TootMedia) map[string]any {
	attach := map[string]any{}
	meta := map[string]any{}
	err := json.Unmarshal([]byte(media.Meta), &meta)
	if err == nil {
		attach["meta"] = meta
	}
	attach["id"] = strconv.FormatUint(media.Id, 10)
	attach["url"] = media.PublicUri("full")
	attach["path"] = media.DiskPath("full")
	attach["preview_url"] = media.PublicUri("thumb")
	attach["preview_path"] = media.DiskPath("thumb")
	attach["type"] = media.MimeType
	attach["preview_type"] = media.ThumbMimeType
	attach["description"] = media.Description
	attach["blurhash"] = media.Blurhash
	return attach
}
