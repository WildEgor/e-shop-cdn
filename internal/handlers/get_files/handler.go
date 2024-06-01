package get_files_handler

import (
	"github.com/WildEgor/e-shop-cdn/internal/dtos"
	"github.com/WildEgor/e-shop-cdn/internal/repositories"
	core_dtos "github.com/WildEgor/e-shop-gopack/pkg/core/dtos"
	"github.com/gofiber/fiber/v2"
)

type GetFilesHandler struct {
	fr repositories.IFilesRepository
}

func NewGetFilesHandler(
	fr repositories.IFilesRepository,
) *GetFilesHandler {
	return &GetFilesHandler{
		fr,
	}
}

// Handle 			Show files godoc
// @Summary 		Allow get paginated files
// @Description		show paginated files
// @Tags			Files Controller
// @Accept			json
// @Produce			json
// @Param			x-api-key header	string	true	"123"
// @Param			page	query		int	false	"Page"
// @Param			limit	query		int	false	"Page"
// @Router			/api/v1/cdn/files [post]
func (h *GetFilesHandler) Handle(c *fiber.Ctx) error {
	resp := core_dtos.NewResp(core_dtos.WithOldContext(c))

	query := &dtos.PaginationQueryDto{
		Page:  c.Query("page", "1"),
		Limit: c.Query("limit", "10"),
	}

	if err := query.Validate(); err != nil {
		resp.SetStatus(fiber.StatusBadRequest)
		return resp.JSON()
	}

	opts := dtos.FromPaginationQueryDtoToPaginationOpts(query)

	pf, err := h.fr.PaginateFiles(opts)
	if err != nil {
		resp.SetStatus(fiber.StatusNotFound)
		return resp.JSON()
	}

	dto := dtos.NewPaginatedResponse(pf.Total, pf.Data)

	resp.SetStatus(fiber.StatusOK)
	resp.SetData(dto)

	return resp.JSON()
}
