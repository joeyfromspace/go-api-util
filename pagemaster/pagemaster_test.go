package pagemaster

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mongoTestInit() (*mongo.Database, *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		panic("cannot connect to mongo")
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	db := client.Database("test")
	coll := db.Collection("testcollection")

	return db, coll
}

func insertTestDocuments(c *mongo.Collection, n int) ([]interface{}, error) {
	docs := make([]interface{}, 0)

	for i := 0; i < n; i++ {
		d := bson.D{
			0: bson.E{Key: "_id", Value: primitive.NewObjectIDFromTimestamp(time.Now())},
			1: bson.E{Key: "createdAt", Value: time.Now().Unix()},
			2: bson.E{Key: "rev", Value: i},
		}
		docs = append(docs, d)
	}

	_, err := c.InsertMany(context.TODO(), docs)
	return docs, err
}

func TestPageMaster_Collection(t *testing.T) {
	type fields struct {
		collection   string
		ctx          context.Context
		database     *mongo.Database
		from         interface{}
		nextToken    string
		pageSize     int64
		queryTimeout time.Duration
	}

	db, coll := mongoTestInit()

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		0: {
			name: "it should return the passed in collection",
			fields: fields{
				collection: "testcollection",
				database:   db,
			},
			want: coll.Name(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PageMaster{
				collection:   tt.fields.collection,
				ctx:          tt.fields.ctx,
				database:     tt.fields.database,
				from:         tt.fields.from,
				nextToken:    tt.fields.nextToken,
				pageSize:     tt.fields.pageSize,
				queryTimeout: tt.fields.queryTimeout,
			}
			if got := p.Collection(); !reflect.DeepEqual(got.Name(), tt.want) {
				t.Errorf("PageMaster.Collection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageMaster_Database(t *testing.T) {
	type fields struct {
		collection   string
		ctx          context.Context
		database     *mongo.Database
		from         interface{}
		nextToken    string
		pageSize     int64
		queryTimeout time.Duration
	}

	db, _ := mongoTestInit()

	tests := []struct {
		name   string
		fields fields
		want   *mongo.Database
	}{
		0: {
			name: "get database back",
			fields: fields{
				collection: "testcollection",
				database:   db,
			},
			want: db,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PageMaster{
				collection:   tt.fields.collection,
				ctx:          tt.fields.ctx,
				database:     tt.fields.database,
				from:         tt.fields.from,
				nextToken:    tt.fields.nextToken,
				pageSize:     tt.fields.pageSize,
				queryTimeout: tt.fields.queryTimeout,
			}
			if got := p.Database(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PageMaster.Database() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageMaster_FindPaginated(t *testing.T) {
	type fields struct {
		collection   string
		ctx          context.Context
		database     *mongo.Database
		from         interface{}
		nextToken    string
		pageSize     int64
		queryTimeout time.Duration
	}

	db, coll := mongoTestInit()
	testDocCount := 500

	_, err := coll.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		t.Errorf("error deleting documents: %s", err)
		return
	}

	testDocs, err := insertTestDocuments(coll, testDocCount)
	if err != nil {
		t.Errorf("error inserting documents: %s", err)
	}

	tests := []struct {
		name    string
		fields  fields
		test    func(t *testing.T, r []interface{}, p *PageMaster)
		wantErr bool
	}{
		0: {
			name: "test defaults are returned",
			fields: fields{
				ctx:          context.TODO(),
				collection:   "testcollection",
				database:     db,
				pageSize:     50,
				queryTimeout: time.Duration(10 * time.Second),
			},
			test: func(t *testing.T, r []interface{}, p *PageMaster) {
				if len(r) != 50 {
					t.Errorf("PageMaster.FindPaginated() unexpected result length = got %v, want %v", len(r), 50)
				}

				for i, d := range r {
					if i >= 50 {
						break
					}
					resDoc := d.(bson.D)
					doc := testDocs[(len(testDocs)-1)-i].(bson.D)
					x := resDoc[0].Value
					y := doc[0].Value
					if !reflect.DeepEqual(x, y) {
						t.Errorf("PageMaster.FindPaginated() unexpected value = got %v, want %v", x, y)
						break
					}
				}

				x := p.NextToken()
				y := r[len(r)-1].(bson.D)[0].Value.(primitive.ObjectID).Hex()
				if !reflect.DeepEqual(x, y) {
					t.Errorf("PageMaster.FindPaginated() unexpected value = got %v, want %v", x, y)
				}
			},
			wantErr: false,
		},
		1: {
			name: "test from token is used",
			fields: fields{
				ctx:          context.TODO(),
				collection:   "testcollection",
				database:     db,
				pageSize:     50,
				from:         testDocs[(len(testDocs)-1)-49].(bson.D)[0].Value.(primitive.ObjectID),
				queryTimeout: time.Duration(1 * time.Minute),
			},
			test: func(t *testing.T, r []interface{}, p *PageMaster) {
				firstExpectedDoc := testDocs[(len(testDocs)-1)-50].(bson.D)
				firstExpectedID := firstExpectedDoc[0].Value
				firstResultDoc := r[0].(bson.D)
				firstResultID := firstResultDoc[0].Value

				if !reflect.DeepEqual(firstExpectedID, firstResultID) {
					t.Errorf("PageMaster.FindPaginated() wrong skip detected in from = got %v, want %v", firstResultID, firstExpectedID)
					return
				}

				for i, d := range r {
					if i >= 50 {
						break
					}
					resDoc := d.(bson.D)
					doc := testDocs[(len(testDocs)-1)-(i+50)].(bson.D)
					x := resDoc[0].Value
					y := doc[0].Value
					if !reflect.DeepEqual(x, y) {
						t.Errorf("PageMaster.FindPaginated() unexpected value = got %v, want %v", x, y)
						break
					}
				}

				x := p.NextToken()
				y := r[len(r)-1].(bson.D)[0].Value.(primitive.ObjectID).Hex()
				if !reflect.DeepEqual(x, y) {
					t.Errorf("PageMaster.FindPaginated() unexpected value = got %v, want %v", x, y)
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PageMaster{
				collection:   tt.fields.collection,
				ctx:          tt.fields.ctx,
				database:     tt.fields.database,
				from:         tt.fields.from,
				nextToken:    tt.fields.nextToken,
				pageSize:     tt.fields.pageSize,
				queryTimeout: tt.fields.queryTimeout,
			}
			got, err := p.FindPaginated()
			if (err != nil) != tt.wantErr {
				t.Errorf("PageMaster.FindPaginated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.test(t, got, p)
		})
	}
}

func TestPageMaster_GetMongoDBQueryFilter(t *testing.T) {
	type fields struct {
		collection   string
		ctx          context.Context
		database     *mongo.Database
		from         interface{}
		nextToken    string
		pageSize     int64
		queryTimeout time.Duration
	}

	db, _ := mongoTestInit()
	testID := primitive.NewObjectIDFromTimestamp(time.Now())

	tests := []struct {
		name   string
		fields fields
		want   bson.M
	}{
		0: {
			name: "ensure the query has the anticipated shape",
			fields: fields{
				collection: "testcollection",
				database:   db,
				from:       testID,
			},
			want: bson.M{
				"_id": bson.M{
					"$lt": testID,
				},
			},
		},
		1: {
			name: "ensure the query is empty if from is zero value",
			fields: fields{
				collection: "testcollection",
				database:   db,
			},
			want: bson.M{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PageMaster{
				collection:   tt.fields.collection,
				ctx:          tt.fields.ctx,
				database:     tt.fields.database,
				from:         tt.fields.from,
				nextToken:    tt.fields.nextToken,
				pageSize:     tt.fields.pageSize,
				queryTimeout: tt.fields.queryTimeout,
			}
			if got := p.GetMongoDBQueryFilter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PageMaster.GetMongoDBQueryFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageMaster_NextToken(t *testing.T) {
	type fields struct {
		collection   string
		ctx          context.Context
		database     *mongo.Database
		from         interface{}
		nextToken    string
		pageSize     int64
		queryTimeout time.Duration
	}

	db, _ := mongoTestInit()
	testID := primitive.NewObjectIDFromTimestamp(time.Now())

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		0: {
			name: "it should return the expected next token",
			fields: fields{
				collection: "testcollection",
				ctx:        context.TODO(),
				database:   db,
				nextToken:  testID.Hex(),
			},
			want: testID.Hex(),
		},
		1: {
			name: "it should be empty string if no next token",
			fields: fields{
				collection: "testcollection",
				ctx:        context.TODO(),
				database:   db,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PageMaster{
				collection:   tt.fields.collection,
				ctx:          tt.fields.ctx,
				database:     tt.fields.database,
				from:         tt.fields.from,
				nextToken:    tt.fields.nextToken,
				pageSize:     tt.fields.pageSize,
				queryTimeout: tt.fields.queryTimeout,
			}
			if got := p.NextToken(); got != tt.want {
				t.Errorf("PageMaster.NextToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageMaster_PageSize(t *testing.T) {
	type fields struct {
		collection   string
		ctx          context.Context
		database     *mongo.Database
		from         interface{}
		nextToken    string
		pageSize     int64
		queryTimeout time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   *int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PageMaster{
				collection:   tt.fields.collection,
				ctx:          tt.fields.ctx,
				database:     tt.fields.database,
				from:         tt.fields.from,
				nextToken:    tt.fields.nextToken,
				pageSize:     tt.fields.pageSize,
				queryTimeout: tt.fields.queryTimeout,
			}
			if got := p.PageSize(); got != tt.want {
				t.Errorf("PageMaster.PageSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		o *NewOptions
	}
	tests := []struct {
		name string
		args args
		want PageMaster
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFromTokenFromRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFromTokenFromRequest(tt.args.r); got != tt.want {
				t.Errorf("getFromTokenFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPageSize(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPageSize(tt.args.r); got != tt.want {
				t.Errorf("getPageSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
