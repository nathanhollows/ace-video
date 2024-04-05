package models

import (
	"context"
)

type Tag struct {
	baseModel
	belongsToUser

	ID      string `bun:",pk,type:varchar(36)" json:"id"`
	MediaID string `bun:",pk,type:varchar(36)" json:"file_id"`
	Media   *Media `bun:"rel:belongs-to,join:media_id=id" json:"media"`
	Name    string `bun:",pk,type:varchar(255)" json:"name"`
}

type Tags []*Tag

// Save the tag to the database
func (t *Tag) Save(ctx context.Context) error {
	_, err := db.NewInsert().Model(t).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// FindAllTags finds all Tag
func FindAllTags(ctx context.Context) (Tags, error) {
	tags := Tags{}
	err := db.NewSelect().
		Model(&tags).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return tags, nil
}
