package infra

import (
	"context"
	"errors"
	"time"

	"auth-service/internal/auth/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Implementasi UserRepository untuk MongoDB
type MongoUserRepository struct {
	Collection *mongo.Collection
}

// Helper: konversi domain.User <-> dokumen Mongo
func toMongoUser(u *domain.User) bson.M {
	id, _ := primitive.ObjectIDFromHex(u.ID)
	return bson.M{
		"_id":        id,
		"name":       u.Name,
		"email":      u.Email,
		"password":   u.Password,
		"saldo":      u.Saldo,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
	}
}

// NEW: Helper safe convert primitive.DateTime / time.Time
func convertToTime(v interface{}) time.Time {
	switch val := v.(type) {
	case time.Time:
		return val
	case primitive.DateTime:
		return val.Time()
	default:
		return time.Time{}
	}
}

func fromMongoUser(doc bson.M) *domain.User {
	id, _ := doc["_id"].(primitive.ObjectID)
	return &domain.User{
		ID:        id.Hex(),
		Name:      doc["name"].(string),
		Email:     doc["email"].(string),
		Password:  doc["password"].(string),
		Saldo:     int(doc["saldo"].(int32)),
		CreatedAt: convertToTime(doc["created_at"]),
		UpdatedAt: convertToTime(doc["updated_at"]),
	}
}

func (r *MongoUserRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		users = append(users, fromMongoUser(doc))
	}
	return users, nil
}

func (r *MongoUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objectID}
	var doc bson.M
	err = r.Collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return fromMongoUser(doc), nil
}

func (r *MongoUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	filter := bson.M{"email": email}
	var doc bson.M
	err := r.Collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return fromMongoUser(doc), nil
}

func (r *MongoUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	id := primitive.NewObjectID()
	user.ID = id.Hex()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := r.Collection.InsertOne(ctx, toMongoUser(user))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, id string, user *domain.User) (*domain.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	user.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"email":      user.Email,
			"password":   user.Password,
			"saldo":      user.Saldo,
			"updated_at": user.UpdatedAt,
		},
	}
	_, err = r.Collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return nil, err
	}
	user.ID = id
	return user, nil
}

func (r *MongoUserRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.Collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
