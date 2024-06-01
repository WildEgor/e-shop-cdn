package replace_handler

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/WildEgor/e-shop-cdn/internal/adapters/pubsub"
	adapters "github.com/WildEgor/e-shop-cdn/internal/adapters/storage"
	domains "github.com/WildEgor/e-shop-cdn/internal/domain"
	"github.com/WildEgor/e-shop-cdn/internal/dtos"
	"github.com/WildEgor/e-shop-cdn/internal/repositories"
	"github.com/WildEgor/e-shop-cdn/internal/utils"
	core_dtos "github.com/WildEgor/e-shop-gopack/pkg/core/dtos"
	"github.com/gofiber/fiber/v2"
)

type ReplaceHandler struct {
	fr     repositories.IFilesRepository
	sa     adapters.IFileStorage
	pubsub pubsub.IPubSub
}

func NewReplaceHandler(
	fr repositories.IFilesRepository,
	sa adapters.IFileStorage,
	pubsub pubsub.IPubSub,
) *ReplaceHandler {
	return &ReplaceHandler{
		fr,
		sa,
		pubsub,
	}
}

// Handle 			Replace file godoc
// @Summary 		Allow to replace file
// @Description		replace file
// @Tags			Upload Controller
// @Accept			multipart/form-data
// @Produce			json
// @Param			x-api-key header	string	true	"123"
// @Param			file	formData	file true	"File"
// @Router			/api/v1/cdn/file/{id}/replace [post]
func (h *ReplaceHandler) Handle(c *fiber.Ctx) error {
	resp := core_dtos.NewResp(core_dtos.WithOldContext(c))
	resp.SetStatus(fiber.StatusOK)

	form, err := c.MultipartForm()
	if err != nil {
		resp.SetStatus(fiber.StatusBadRequest)
		resp.SetMessage(err.Error())
		return resp.JSON()
	}

	files := form.File["file"]
	if len(files) == 0 || len(files) > 1 {
		resp.SetStatus(fiber.StatusBadRequest)
		return resp.JSON()
	}

	query := &dtos.FileIdDto{
		FileId: c.Params("id"),
	}
	if err := query.Validate(); err != nil {
		resp.SetStatus(fiber.StatusBadRequest)
		resp.SetMessage(err.Error())
		return resp.JSON()
	}

	exist, err := h.fr.FindById(query.FileId)
	if err != nil {
		resp.SetStatus(fiber.StatusNotFound)
		resp.SetMessage(err.Error())
		return resp.JSON()
	}

	fileWrapper := domains.WrapFileModel(exist, files[0])
	if !fileWrapper.IsValidFormat() {
		domains.SetFileExtErr(resp, fileWrapper.Name)
		return resp.JSON()
	}

	fbuf, err := utils.ReadFileToBuffer(fileWrapper.Data())
	if err != nil {
		domains.SetFileServeErr(resp)
		return resp.JSON()
	}

	fr, err := fileWrapper.Data().Open()
	if err != nil {
		domains.SetFileServeErr(resp)
		return resp.JSON()
	}

	if err := h.sa.Upload(c.Context(), fileWrapper.FullPath(), fr); err != nil {
		domains.SetStorageErr(resp)
		return resp.JSON()
	}

	checksum := md5.Sum(fbuf)
	exist.CheckSum = hex.EncodeToString(checksum[:])
	if err := h.fr.UpdateFile(exist); err != nil {
		resp.SetStatus(fiber.StatusInternalServerError)
		return resp.JSON()
	}

	h.pubsub.Publish([]string{fileWrapper.DirPrefix()}, "UPDATED")

	return resp.JSON()
}
