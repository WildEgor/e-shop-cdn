package upload_handler

import (
	"crypto/md5"
	"github.com/WildEgor/e-shop-cdn/internal/adapters/pubsub"
	adapters "github.com/WildEgor/e-shop-cdn/internal/adapters/storage"
	domains "github.com/WildEgor/e-shop-cdn/internal/domain"
	"github.com/WildEgor/e-shop-cdn/internal/dtos"
	"github.com/WildEgor/e-shop-cdn/internal/repositories"
	"github.com/WildEgor/e-shop-cdn/internal/utils"
	core_dtos "github.com/WildEgor/e-shop-gopack/pkg/core/dtos"
	"github.com/gofiber/fiber/v2"
	"sync"
)

type UploadHandler struct {
	fr     repositories.IFilesRepository
	sa     adapters.IFileStorage
	pubsub pubsub.IPubSub
}

func NewUploadHandler(
	fr repositories.IFilesRepository,
	sa adapters.IFileStorage,
	pubsub pubsub.IPubSub,
) *UploadHandler {
	return &UploadHandler{
		fr,
		sa,
		pubsub,
	}
}

//		Upload files godoc
//		@Summary Allow upload multiple files
//		@Description	upload files
//		@Tags			Upload Controller
//		@Accept			multipart/form-data
//		@Produce		json
//		@Param			x-api-key header	string	true	"123"
//	 	@Param			files	formData	file true	"Files"
//		@Router			/api/v1/cdn/upload [post]
func (h *UploadHandler) Handle(c *fiber.Ctx) error {
	resp := core_dtos.NewResp(core_dtos.WithOldContext(c))

	resp.SetStatus(fiber.StatusOK)

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	catalog := c.FormValue("catalog")

	files := form.File["files"]
	vf := make([]domains.FileWrapper, 0)

	for _, file := range files {
		f := domains.WrapFile(catalog, file)

		if f.IsValidFormat() {
			vf = append(vf, f)
			continue
		}

		domains.SetFileExtErr(resp, f.Name)
	}

	if len(vf) != len(files) {
		resp.SetStatus(fiber.StatusBadRequest)
		return resp.JSON()
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(vf))

	rf := make([]dtos.FileResponseDto, 0)

	for _, file := range vf {
		go func(r *core_dtos.ResponseDto, fileWrapper domains.FileWrapper) {
			defer wg.Done()

			fr, err := fileWrapper.Data().Open()
			defer fr.Close() // TODO: handle error!
			if err != nil {
				mu.Lock()
				defer mu.Unlock()
				r.SetStatus(fiber.StatusBadRequest)
				domains.SetFileServeErr(r)
				return
			}

			if err := h.sa.Upload(c.Context(), fileWrapper.FullPath(), fr); err != nil {
				mu.Lock()
				defer mu.Unlock()
				r.SetStatus(fiber.StatusInternalServerError)
				domains.SetStorageErr(r)
				return
			}

			fbuf, err := utils.ReadFileToBuffer(fileWrapper.Data())
			if err != nil {
				mu.Lock()
				defer mu.Unlock()
				resp.SetStatus(fiber.StatusInternalServerError)
				domains.SetFileServeErr(resp)
				return
			}

			checksum := md5.Sum(fbuf)
			ofn, err := h.fr.AddFile(fileWrapper.FullPath(), checksum[:])
			if err != nil {
				mu.Lock()
				defer mu.Unlock()
				r.SetStatus(fiber.StatusBadRequest)
				domains.SetSaveFileErr(r)
				return
			}

			fp := dtos.FileResponseDto{
				Filename:    ofn,
				DownloadUrl: h.sa.DownloadURL(ofn),
			}

			mu.Lock()
			defer mu.Unlock()
			rf = append(rf, fp)

			h.pubsub.Publish([]string{fileWrapper.DirPrefix()}, "ADDED")
		}(resp, file)
	}

	wg.Wait()

	resp.SetData(rf)

	return resp.JSON()
}
