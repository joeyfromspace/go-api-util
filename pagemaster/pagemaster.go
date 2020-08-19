package pagemaster

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PageMaster is a pagination struct that scrolls through pages
type PageMaster struct {
	collection   string
	ctx          context.Context
	database     *mongo.Database
	from         interface{}
	nextToken    string
	pageSize     int64
	queryTimeout time.Duration
}

// MongoDBQueryFilter is a struct that describes a pagination filter for mongodb
type MongoDBQueryFilter struct {
	ID struct {
		LTE primitive.ObjectID `bson:"$lte"`
	} `bson:"_id"`
}

// Collection returns the collection associated with the PageMaster object
func (p *PageMaster) Collection() *mongo.Collection {
	return p.Database().Collection(p.collection)
}

// Database returns the database associated with the PageMaster object
func (p *PageMaster) Database() *mongo.Database {
	return p.database
}

// FindPaginated executes a mongodb query with the paginated items
func (p *PageMaster) FindPaginated() ([]bson.M, error) {
	results := make([]bson.M, 0)
	filter := p.GetMongoDBQueryFilter()

	coll := p.Collection()

	ctx, cancel := context.WithTimeout(p.ctx, p.queryTimeout)
	defer cancel()

	cursor, err := coll.Find(ctx, filter, &options.FindOptions{
		Limit: p.PageSize(),
	})

	if err != nil {
		return results, err
	}

	ctx, cancel = context.WithTimeout(p.ctx, p.queryTimeout)
	defer cancel()
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &results)

	if err != nil {
		return results, err
	}

	if len(results) > 0 {
		lastDoc := results[len(results)-1]
		oid, ok := lastDoc["_id"].(primitive.ObjectID)
		if ok {
			p.nextToken = oid.Hex()
		}
	}

	return results, nil
}

// GetMongoDBQueryFilter creates pagination filters for mongo queries
func (p *PageMaster) GetMongoDBQueryFilter() MongoDBQueryFilter {
	f := MongoDBQueryFilter{}
	f.ID.LTE = p.from.(primitive.ObjectID)

	return f
}

// NextToken is the string reprsentation of the next object
func (p *PageMaster) NextToken() string {
	return p.nextToken
}

// PageSize returns the page size of the paginated instance
func (p *PageMaster) PageSize() *int64 {
	return &p.pageSize
}

// New instantiates a new PageMaster from passed in options
func New(r *http.Request, d *mongo.Database, collName string, q time.Duration) PageMaster {
	from := r.URL.Query().Get("from")
	pageSize := getPageSize(r)

	if q == 0 {
		q = time.Duration(1 * time.Minute)
	}

	return PageMaster{
		collection:   collName,
		ctx:          r.Context(),
		database:     d,
		from:         from,
		pageSize:     pageSize,
		queryTimeout: q,
	}
}

func getPageSize(r *http.Request) int64 {
	pgStr := r.URL.Query().Get("pageSize")

	if pgStr == "" {
		pgStr = "50"
	}

	i, err := strconv.Atoi(pgStr)

	if err != nil {
		i = 50
	}

	if i > 100 {
		i = 100
	}

	return int64(i)
}
