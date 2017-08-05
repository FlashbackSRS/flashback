package model

import (
	fb "github.com/FlashbackSRS/flashback-model"
)

func extractTemplateFiles(v *fb.FileCollectionView) map[string]string {
	templates := make(map[string]string)
	for _, filename := range v.FileList() {
		att, _ := v.GetFile(filename)
		if _, ok := templateTypes[att.ContentType]; ok {
			templates[filename] = string(att.Content)
		}
	}
	return templates
}
