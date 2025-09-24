package converter

import models "github.com/shahariaz/user_segmentation/internal/model"

type Converter struct {
	model             *models.SchemaInfo
	operators         map[string]string
	versionFields     map[string]string
	reversePredicates map[string]string
}

func (c *Converter) ConvertToDQL() {

}
