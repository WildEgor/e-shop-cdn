package repositories

import (
	"context"
	"encoding/hex"
	mongo2 "github.com/WildEgor/e-shop-cdn/internal/db/mongo"
	"github.com/WildEgor/e-shop-cdn/internal/dtos"
	"github.com/WildEgor/e-shop-cdn/internal/models"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type IFilesRepository interface {
	RenameFile(oldname, newname string) error
	UpdateFile(model *models.FileModel) error
	FindById(id string) (*models.FileModel, error)
	RemoveFileById(id string) (*models.FileModel, error)
	DeleteById(id string) error
	AddFile(filename string, checksum []byte) (string, error)
	PaginateFiles(opts *dtos.PaginationOpts) (*models.PaginatedFiles, error)
	StreamDeletedFiles() <-chan *models.FileModel
}

var _ IFilesRepository = (*FileRepository)(nil)

type FileRepository struct {
	coll *mongo.Collection
}

func NewFileRepository(
	db *mongo2.Connection,
) *FileRepository {

	coll := db.Db().Collection(models.CollectionFiles)

	return &FileRepository{
		coll,
	}
}

func (r *FileRepository) StreamDeletedFiles() <-chan *models.FileModel {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	out := make(chan *models.FileModel)
	filter := bson.D{{
		Key: "status", Value: models.DeletedStatus,
	}}

	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
	}

	go func() {
		defer cancel()
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var file models.FileModel
			if err := cursor.Decode(&file); err != nil {
				continue
			}

			out <- &file
		}

		close(out)
	}()

	return out
}

func (r *FileRepository) PaginateFiles(opts *dtos.PaginationOpts) (*models.PaginatedFiles, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := &models.PaginatedFiles{
		Data: make([]models.FileModel, 0),
	}

	l := int64(opts.Limit)
	skip := int64(opts.Page*opts.Limit - opts.Limit)
	fOpt := options.FindOptions{Limit: &l, Skip: &skip}
	filter := bson.D{{
		Key: "status", Value: models.ActiveStatus,
	}}

	curr, err := r.coll.Find(ctx, filter, &fOpt)
	if err != nil {
		return result, err
	}
	defer curr.Close(ctx)

	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return result, err
	}

	result.Total = count

	for curr.Next(ctx) {
		var el models.FileModel
		curr.Decode(&el)

		result.Data = append(result.Data, el)
	}

	return result, nil
}

func (r *FileRepository) AddFile(filename string, checksum []byte) (string, error) {
	model := &models.FileModel{
		Name:      filename,
		CheckSum:  hex.EncodeToString(checksum),
		Status:    models.ActiveStatus,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	filter := bson.D{{
		Key: "$and",
		Value: bson.D{
			{"check_sum", bson.D{
				{"$eq", model.CheckSum},
			}}},
	}}

	existed := r.coll.FindOne(context.TODO(), filter) // TODO: ctx

	if existed != nil && existed.Decode(&model) == nil {
		return model.Name, nil
	}

	_, err := r.coll.InsertOne(context.TODO(), model) // TODO: ctx
	if err != nil {
		return "", errors.New(`err insert`) // TODO
	}

	return model.Name, nil
}

func (r *FileRepository) RemoveFileById(id string) (*models.FileModel, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var model *models.FileModel
	filter := bson.D{{Key: "_id", Value: _id}}

	if err := r.coll.FindOne(context.TODO(), filter).Decode(&model); err != nil {
		if errors.As(err, &mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	if model != nil {
		update := bson.D{
			{"$set",
				bson.D{
					{"status", models.DeletedStatus},
					{"updated_at", time.Now().UTC()},
				},
			},
		}

		_, err := r.coll.UpdateOne(context.Background(), filter, update) // TODO: ctx
		if err != nil {
			return nil, err
		}
	}

	return model, nil
}

func (r *FileRepository) RenameFile(oldname, newname string) error {
	update := bson.D{
		{"$set",
			bson.D{
				{"file_name", newname},
				{"updated_at", time.Now().UTC()},
			},
		},
	}
	filter := bson.D{{Key: "file_name", Value: oldname}, {Key: "is_deleted", Value: false}}

	_, err := r.coll.UpdateOne(context.Background(), filter, update) // TODO: ctx
	if err != nil {
		return errors.New(`Mongo error`) // TODO
	}

	return nil
}

func (r *FileRepository) UpdateFile(model *models.FileModel) error {
	filter := bson.D{{Key: "_id", Value: model.Id}, {"status", models.ActiveStatus}}

	update := bson.D{
		{"$set",
			bson.D{
				{"file_name", model.Name},
				{"checksum", model.CheckSum},
				{"updated_at", time.Now().UTC()},
			},
		},
	}

	_, err := r.coll.UpdateOne(context.Background(), filter, update) // TODO: ctx
	if err != nil {
		return errors.New(`Mongo error`) // TODO
	}

	return nil
}

func (r *FileRepository) FindById(id string) (*models.FileModel, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var model *models.FileModel
	filter := bson.D{{Key: "_id", Value: _id}, {"status", models.ActiveStatus}}

	if err := r.coll.FindOne(context.TODO(), filter).Decode(&model); err != nil {
		if errors.As(err, &mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return model, err
}

func (r *FileRepository) DeleteById(id string) error {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.D{
		{"_id",
			_id,
		},
	}

	_, err = r.coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}
