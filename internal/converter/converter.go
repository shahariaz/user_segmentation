package converter

import (
	"github.com/shahariaz/user_segmentation/internal/config"
	models "github.com/shahariaz/user_segmentation/internal/model"
)

type Converter struct {
	schema            *models.SchemaInfo
	operators         map[string]string
	versionFields     map[string]string
	reversePredicates map[string]string
}

func NewConverter() *Converter {
	schema := config.GetSchemaConfig()
	return &Converter{
		schema:            schema,
		operators:         config.GetOperatorMappings(),
		versionFields:     config.GetVersionFields(),
		reversePredicates: config.GetReversePredicates(),
	}
}

func (c *Converter) ConvertToDQL(jsonQuery *models.JSONQuery) (*models.JSONQuery, error) {

	return jsonQuery, nil
}
