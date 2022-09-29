package datastore

import (
	"context"
	"errors"
	"reflect"
	"time"

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

var (
	// TODO: load this from a config file or env vars
	collectionKeys = []CollectionKey{CollectionKey("users")}
)

type MongoStore struct {
	IsConnected bool
	Database    *mongo.Database
}

type Store interface {
	FindAll(ctx context.Context, filter bson.M, sort interface{}, projection, results interface{}) error

	UpdateMany(ctx context.Context, filter, payload bson.M, bulk bool) error

	DeleteMany(ctx context.Context, filter bson.M, hardDelete bool) error

	WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error
}

// mongodb driver -> store (database) -> repo -> service -> handler

var _ Store = &MongoStore{}

/*
 * New
 * This initialises a new MongoDB repo for the collection
 */
func New(database *mongo.Database) Store {
	MongoStore := &MongoStore{
		IsConnected: true,
		Database:    database,
	}

	return MongoStore
}

func IsValidPointer(i interface{}) bool {
	v := reflect.ValueOf(i)
	return v.Type().Kind() == reflect.Ptr && !v.IsNil()
}

func (d *MongoStore) FindAll(ctx context.Context, filter bson.M, sort interface{}, projection, results interface{}) error {
	if !IsValidPointer(results) {
		return ErrInvalidPtr
	}

	col, err := d.retrieveCollection(ctx)
	if err != nil {
		return err
	}
	collection := d.Database.Collection(col)

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

func (d *MongoStore) UpdateMany(ctx context.Context, filter, payload bson.M, bulk bool) error {
	col, err := d.retrieveCollection(ctx)
	if err != nil {
		return err
	}

	collection := d.Database.Collection(col)

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

	log.Infof("cascade: results of update %s op: %+v\n", collection.Name(), res)

	return nil
}

func (d *MongoStore) DeleteMany(ctx context.Context, filter bson.M, hardDelete bool) error {
	col, err := d.retrieveCollection(ctx)
	if err != nil {
		return err
	}
	collection := d.Database.Collection(col)

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

func (d *MongoStore) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := d.Database.Client().StartSession()
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

func (d *MongoStore) retrieveCollection(ctx context.Context) (string, error) {
	for _, key := range collectionKeys {
		if ctx.Value(CollectionCtx) == key {
			return string(key), nil
		}
	}

	return "", ErrInvalidCollection
}
