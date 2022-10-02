package datastore

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/jirevwe/cascade/internal/pkg/config"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionKey string

const CollectionCtx CollectionKey = "collection"

var (
	ErrInvalidCollection = errors.New("Invalid collection type")
	ErrInvalidPtr        = errors.New("out param is not a valid pointer")
)

type Store struct {
	database *mongo.Database
}

func (d *Store) GetDatabase() *mongo.Database {
	return d.database
}

type DB interface {
	GetDatabase() *mongo.Database

	FindAll(ctx context.Context, filter bson.M, sort interface{}, projection, results interface{}) error

	UpdateMany(ctx context.Context, filter, payload bson.M, bulk bool) error

	DeleteMany(ctx context.Context, filter bson.M, hardDelete bool) error

	WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error
}

// mongodb driver -> store (database) -> repo -> service -> handler

var _ DB = &Store{}

/*
 * New initialises a new MongoDB collection pool
 */
func New(cfg config.Configuration) (DB, error) {
	opts := options.Client()
	opts.ApplyURI(cfg.MongoDsn)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	m := &Store{
		database: client.Database(cfg.DbName),
	}

	return m, nil
}

func IsValidPointer(i interface{}) bool {
	v := reflect.ValueOf(i)
	return v.Type().Kind() == reflect.Ptr && !v.IsNil()
}

func (d *Store) FindAll(ctx context.Context, filter bson.M, sort interface{}, projection, results interface{}) error {
	if !IsValidPointer(results) {
		return ErrInvalidPtr
	}

	col, err := d.retrieveCollection(ctx)
	if err != nil {
		return err
	}

	collection := d.database.Collection(col)

	ops := options.Find()

	if projection != nil {
		ops.Projection = projection
	}

	if sort != nil {
		ops.Sort = sort
	}

	if filter == nil {
		filter = bson.M{}
	}

	cursor, err := collection.Find(ctx, filter, ops)
	if err != nil {
		return err
	}

	return cursor.All(ctx, results)
}

func (d *Store) UpdateMany(ctx context.Context, filter, payload bson.M, bulk bool) error {
	col, err := d.retrieveCollection(ctx)
	if err != nil {
		return err
	}

	collection := d.database.Collection(col)

	if !bulk {
		_, err = collection.UpdateMany(ctx, filter, payload)
		return err
	}

	var msgOperations []mongo.WriteModel
	updateMessagesOperation := mongo.NewUpdateManyModel()
	updateMessagesOperation.SetFilter(filter)
	updateMessagesOperation.SetUpdate(payload)

	msgOperations = append(msgOperations, updateMessagesOperation)
	res, err := collection.BulkWrite(ctx, msgOperations)
	if err != nil {
		return err
	}

	log.Infof("store: results of update %s op: %+v", collection.Name(), res)

	return nil
}

func (d *Store) DeleteMany(ctx context.Context, filter bson.M, hardDelete bool) error {
	col, err := d.retrieveCollection(ctx)
	if err != nil {
		return err
	}
	collection := d.database.Collection(col)

	payload := bson.M{
		"deleted_at": primitive.NewDateTimeFromTime(time.Now()),
	}

	if hardDelete {
		_, err := collection.DeleteMany(ctx, filter)
		return err
	} else {
		_, err := collection.UpdateMany(ctx, filter, bson.M{"$set": payload})
		return err
	}
}

func (d *Store) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := d.database.Client().StartSession()
	if err != nil {
		return err
	}

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := fn(sessCtx)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (d *Store) retrieveCollection(ctx context.Context) (string, error) {
	c, err := d.database.ListCollections(ctx, bson.M{})
	if err != nil {
		return "", err
	}
	defer c.Close(ctx)

	if c.Next(ctx) {
		key := c.Current.Lookup("name").StringValue()
		if ctx.Value(CollectionCtx) == key {
			return key, nil
		}
	}

	return "", ErrInvalidCollection
}
