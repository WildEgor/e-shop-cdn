package domains

import (
	"fmt"
	"github.com/WildEgor/e-shop-cdn/internal/models"
	"github.com/WildEgor/e-shop-cdn/internal/utils"
	"mime/multipart"
	"slices"
)

// FileWrapper wrap multipart files
type FileWrapper struct {
	Name    string
	Catalog string
	data    *multipart.FileHeader
}

// WrapFile wrap file (sanitize filename)
func WrapFile(catalog string, data *multipart.FileHeader) FileWrapper {
	name, _ := utils.SanitizeFilename(data.Filename)
	return FileWrapper{
		name,
		catalog,
		data,
	}
}

func WrapFileModel(model *models.FileModel, data *multipart.FileHeader) FileWrapper {
	return FileWrapper{
		model.Name,
		"",
		data,
	}
}

func (f FileWrapper) FullPath() string {
	if len(f.Catalog) == 0 {
		return f.Name
	}

	return fmt.Sprintf("%s/%s", f.Catalog, f.Name)
}

func (f FileWrapper) DirPrefix() string {
	return fmt.Sprintf("%s/*", f.Catalog)
}

// Data return raw data
func (f FileWrapper) Data() *multipart.FileHeader {
	return f.data
}

// IsValidFormat check if file allowed
func (f FileWrapper) IsValidFormat() bool {
	return slices.Contains([]string{"image/jpeg", "image/png"}, f.data.Header["Content-Type"][0])
}

// IsEqualName check name equality (naive)
func (f FileWrapper) IsEqualName(name string) bool {
	return f.Name == name
}
