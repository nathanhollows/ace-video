package models

import (
	"context"
	"os"
	"regexp"

	"github.com/google/uuid"
)

type Media struct {
	baseModel
	belongsToUser

	ID          string `bun:",pk,type:varchar(36)" json:"id"`
	Title       string `bun:",type:varchar(255)" json:"title"`
	FileName    string `bun:",type:varchar(255)" json:"-"`
	MimeType    string `bun:",type:varchar(255)" json:"mime_type"`
	FilePath    string `bun:",type:varchar(255)" json:"file_path"`
	Description string `bun:",type:text" json:"description"`
	Caption     string `bun:",type:text" json:"caption"`
	Tags        Tags   `bun:"rel:has-many,join:id=media_id" json:"tags"`
}

type Library []*Media

// Save the media to the database
func (m *Media) Save(ctx context.Context) error {
	var err error
	if m.ID == "" {
		m.ID = uuid.New().String()
		_, err = db.NewInsert().Model(m).Exec(ctx)
	} else {
		_, err = db.NewUpdate().Model(m).
			Where("id = ?", m.ID).
			Exec(ctx)
	}

	if err != nil {
		return err
	}
	return nil
}

// FindMediaByID finds media by ID
func FindMediaByID(ctx context.Context, id string) (*Media, error) {
	media := &Media{}
	err := db.NewSelect().
		Model(media).
		Relation("Tags").
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return media, nil
}

// FindAllMedia finds all media
func FindAllMedia(ctx context.Context) (Library, error) {
	media := Library{}
	err := db.NewSelect().
		Model(&media).
		Relation("Tags").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return media, nil
}

// FindMatchingMedia finds media that match the options
func FindMatchingMedia(ctx context.Context, options JSONOptions) (Library, error) {
	media := Library{}
	query := db.NewSelect().
		Model(&media).
		Relation("Tags")
	if options.Limit != 0 {
		query = query.Limit(options.Limit)
	}
	if options.Offset != 0 {
		query = query.Offset(options.Offset)
	}
	if options.Sort != "" {
		query = query.OrderExpr(options.Sort + " " + options.Order)
	}
	if options.Search != "" {
		query = query.Where("title LIKE ?", "%"+options.Search+"%")
	}
	if options.Tags != "" {
		// Tags are comma separated
		// The media must match any of the tags
		query = query.Where("tags.name IN (?)", options.Tags)
	}
	if options.ID != "" {
		query = query.Where("id = ?", options.ID)
	}
	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return media, nil
}

// SaveMedia saves the media to the database and the file to the filesystem
func SaveMedia(ctx context.Context, file []byte, fileName string) error {
	media := &Media{
		ID:       uuid.New().String(),
		Title:    fileName,
		FileName: fileName,
		FilePath: "/media/" + fileName,
	}
	err := media.Save(ctx)
	if err != nil {
		return err
	}
	err = os.WriteFile("static/media/"+fileName, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Type returns the type of the model
func (m *Media) Type() string {
	// Can be "image", "video", "audio", "document", "other"
	// Split the mime type by "/" and get the first part
	re := regexp.MustCompile(`[^/]+`)
	return re.FindString(m.MimeType)
}

// GetFileURL returns the URL of the file
func (m *Media) GetFileURL() string {
	site := os.Getenv("SITE_URL")
	return site + "assets" + m.FilePath
}
