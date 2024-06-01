package metadata_handler

import (
	adapters "github.com/WildEgor/e-shop-cdn/internal/adapters/storage"
	"github.com/WildEgor/e-shop-cdn/internal/dtos"
	core_dtos "github.com/WildEgor/e-shop-gopack/pkg/core/dtos"
	"github.com/gofiber/fiber/v2"
)

type MetadataHandler struct {
	sa adapters.IFileStorage
}

func NewMetadataHandler(
	sa adapters.IFileStorage,
) *MetadataHandler {
	return &MetadataHandler{
		sa,
	}
}

// Handle 			Show file metadata godoc
// @Summary 		Allow get file metadata
// @Description		show file metadata
// @Tags			Files Controller
// @Accept			json
// @Produce			json
// @Param			x-api-key header	string	true	"123"
// @Param			page	path		string	true	"Filename"
// @Router			/api/v1/cdn/metadata/{filename} [get]
func (h *MetadataHandler) Handle(c *fiber.Ctx) error {
	resp := core_dtos.NewResp(core_dtos.WithOldContext(c))

	query := &dtos.FileQueryDto{
		Filename: c.Params("filename"),
	}

	if err := query.Validate(); err != nil {
		resp.SetStatus(fiber.StatusBadRequest)
		return resp.JSON()
	}

	file, err := h.sa.Metadata(query.Filename)
	if err != nil {
		resp.SetStatus(fiber.StatusNotFound)
		return resp.JSON()
	}

	dto := &dtos.FileMetadataResponseDto{
		Filename: query.Filename,
		FileSize: file.Size,
	}
	dto.SetDownloadUrl(c, query.Filename)

	resp.SetStatus(fiber.StatusOK)
	resp.SetData(dto)

	return resp.JSON()
}
