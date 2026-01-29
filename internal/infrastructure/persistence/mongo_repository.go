// internal/infrastructure/persistence/mongo_repository.go
package persistence

import (
	"context"
	"fmt"
	"time"

	"myapp/internal/application/interfaces"
	"myapp/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoUnitOfWork struct {
	client  *mongo.Client
	db      *mongo.Database
	session mongo.Session
}

func NewMongoUnitOfWork(client *mongo.Client, dbName string) interfaces.UnitOfWork {
	return &mongoUnitOfWork{
		client: client,
		db:     client.Database(dbName),
	}
}

func (uow *mongoUnitOfWork) Begin(ctx context.Context) error {
	session, err := uow.client.StartSession()
	if err != nil {
		return err
	}

	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return err
	}

	uow.session = session
	return nil
}

func (uow *mongoUnitOfWork) Commit(ctx context.Context) error {
	if uow.session == nil {
		return nil
	}

	if err := uow.session.CommitTransaction(ctx); err != nil {
		return err
	}

	uow.session.EndSession(ctx)
	uow.session = nil
	return nil
}

func (uow *mongoUnitOfWork) Rollback(ctx context.Context) error {
	if uow.session == nil {
		return nil
	}

	if err := uow.session.AbortTransaction(ctx); err != nil {
		return err
	}

	uow.session.EndSession(ctx)
	uow.session = nil
	return nil
}

func (uow *mongoUnitOfWork) Products() interfaces.ProductRepository {
	return &mongoProductRepository{
		collection: uow.db.Collection("products"),
		session:    uow.session,
	}
}

func (uow *mongoUnitOfWork) Tenants() interfaces.TenantRepository {
	return &mongoTenantRepository{
		collection: uow.db.Collection("tenants"),
		session:    uow.session,
	}
}

func (uow *mongoUnitOfWork) StockHistory() interfaces.StockHistoryRepository {
	return &mongoStockHistoryRepository{
		collection: uow.db.Collection("stock_history"),
		session:    uow.session,
	}
}

// Product Repository Implementation
type mongoProductRepository struct {
	collection *mongo.Collection
	session    mongo.Session
}

func (r *mongoProductRepository) FindByID(ctx context.Context, productID string) (*domain.Product, error) {
	if r.session != nil {
		ctx = mongo.NewSessionContext(ctx, r.session)
	}

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, domain.ErrInvalidProductID
	}

	var result struct {
		ID           primitive.ObjectID `bson:"_id"`
		Name         string             `bson:"name"`
		CurrentStock int                `bson:"current_stock"`
		LastUpdated  time.Time          `bson:"last_updated"`
		TenantID     string             `bson:"tenant_id"`
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}, options.FindOne()).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	stock, _ := domain.NewStockQuantity(result.CurrentStock)
	return &domain.Product{
		ID:           result.ID.Hex(),
		Name:         result.Name,
		CurrentStock: stock,
		LastUpdated:  result.LastUpdated,
		TenantID:     result.TenantID,
	}, nil
}

func (r *mongoProductRepository) Save(ctx context.Context, product *domain.Product) error {
	if r.session != nil {
		ctx = mongo.NewSessionContext(ctx, r.session)
	}

	objID, _ := primitive.ObjectIDFromHex(product.ID)

	update := bson.M{
		"$set": bson.M{
			"current_stock": product.CurrentStock.Value(),
			"last_updated":  product.LastUpdated,
		},
		"$inc": bson.M{
			"total_added": product.CurrentStock.Value(), // Simplified
		},
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		update,
		options.Update(),
	)

	return err
}

func (r *mongoProductRepository) UpdateStock(ctx context.Context, productID string, newStock domain.StockQuantity) error {
	if r.session != nil {
		ctx = mongo.NewSessionContext(ctx, r.session)
	}

	// Alternative implementation
	objID, _ := primitive.ObjectIDFromHex(productID)

	update := bson.M{
		"$set": bson.M{
			"current_stock": newStock.Value(),
			"last_updated":  time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		update,
	)

	return err
}

// Tenant Repository Implementation
type mongoTenantRepository struct {
	collection *mongo.Collection
	session    mongo.Session
}

func (r *mongoTenantRepository) FindByID(ctx context.Context, tenantID string) (*domain.Tenant, error) {
	if r.session != nil {
		ctx = mongo.NewSessionContext(ctx, r.session)
	}

	var result struct {
		ID       string `bson:"_id"`
		Name     string `bson:"name"`
		MaxStock int    `bson:"max_stock"`
		IsActive bool   `bson:"is_active"`
	}

	err := r.collection.FindOne(ctx, bson.M{"_id": tenantID}, options.FindOne()).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	maxStock, _ := domain.NewStockQuantity(result.MaxStock)
	return &domain.Tenant{
		ID:       result.ID,
		Name:     result.Name,
		MaxStock: maxStock,
		IsActive: result.IsActive,
	}, nil
}

// Stock History Repository Implementation
type mongoStockHistoryRepository struct {
	collection *mongo.Collection
	session    mongo.Session
}

func (r *mongoStockHistoryRepository) Create(ctx context.Context, event domain.StockAddedEvent) error {
	if r.session != nil {
		ctx = mongo.NewSessionContext(ctx, r.session)
	}

	productID, _ := primitive.ObjectIDFromHex(event.ProductID)

	document := bson.M{
		"product_id":     productID,
		"tenant_id":      event.TenantID,
		"quantity":       event.Quantity.Value(),
		"previous_stock": event.Previous.Value(),
		"new_stock":      event.Current.Value(),
		"added_by":       event.AddedBy,
		"notes":          event.Notes,
		"created_at":     event.Timestamp,
		"operation":      "stock_add",
	}

	_, err := r.collection.InsertOne(ctx, document, options.InsertOne())
	return err
}
