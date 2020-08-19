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

const maxPageSize int64 = 100

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

// NewOptions used to specify initialization options for a new PageMaster instance
type NewOptions struct {
	Collection         string
	Context            context.Context
	Database           *mongo.Database
	FromToken          string
	PageSize           int64
	UnmarshalInterface interface{}
	QueryTimeout       time.Duration
	Request            *http.Request
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
func (p *PageMaster) FindPaginated() ([]interface{}, error) {
	var lastID primitive.ObjectID
	results := make([]interface{}, 0)
	filter := p.GetMongoDBQueryFilter()

	coll := p.Collection()

	ctx, cancel := context.WithTimeout(p.ctx, p.queryTimeout)
	defer cancel()

	cursor, err := coll.Find(ctx, filter, &options.FindOptions{
		Limit: p.PageSize(),
		Sort:  bson.M{"_id": -1},
	})

	if err != nil {
		return results, err
	}

	ctx, cancel = context.WithTimeout(p.ctx, p.queryTimeout)
	defer cancel()
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var d bson.M
		var v bson.D

		err = cursor.Decode(&v)
		if err != nil {
			return results, err
		}

		err = cursor.Decode(&d)
		if err != nil {
			return results, err
		}

		lid, ok := d["_id"].(primitive.ObjectID)
		if ok {
			lastID = lid
		}
		results = append(results, v)
	}

	if len(results) > 0 {
		p.nextToken = lastID.Hex()
	}

	return results, nil
}

// GetMongoDBQueryFilter creates pagination filters for mongo queries
func (p *PageMaster) GetMongoDBQueryFilter() bson.M {
	f := bson.M{}
	if p.from != nil {
		f["_id"] = bson.M{"$lt": p.from.(primitive.ObjectID)}
	}

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
func New(o *NewOptions) PageMaster {
	c := o.Collection
	r := o.Request
	d := o.Database
	from := o.FromToken
	pageSize := o.PageSize
	qt := o.QueryTimeout

	if r == nil {
		panic("instantiated with no request")
	}

	if d == nil {
		panic("instantiated with no database")
	}

	if c == "" {
		panic("instantiated with no collection name")
	}

	if from == "" {
		from = getFromTokenFromRequest(o.Request)
	}

	if pageSize == 0 {
		pageSize = getPageSize(o.Request)
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	if qt == 0 {
		qt = time.Duration(1 * time.Minute)
	}

	return PageMaster{
		collection:   c,
		ctx:          r.Context(),
		database:     d,
		from:         from,
		pageSize:     pageSize,
		queryTimeout: qt,
	}
}

func getFromTokenFromRequest(r *http.Request) string {
	return r.URL.Query().Get("from")
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
