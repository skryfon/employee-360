package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

// gormRoleRepository implements repository.RoleRepository. The port has no
// tenant parameter, so every query is scoped by the tenant in context
// (ctx.TenantIDFromContext); a context without a tenant matches nothing.
type gormRoleRepository struct {
	db *gorm.DB
}

// NewGormRoleRepository constructs a GORM-backed RoleRepository.
func NewGormRoleRepository(db *gorm.DB) repository.RoleRepository {
	return &gormRoleRepository{db: db}
}

var _ repository.RoleRepository = (*gormRoleRepository)(nil)

// scoped returns a DB handle restricted to the context tenant.
func (r *gormRoleRepository) scoped(c context.Context) *gorm.DB {
	tenantID, err := uuid.Parse(tenantString(c))
	if err != nil {
		tenantID = uuid.Nil // matches no row: tenants never use the nil UUID
	}
	return database.DBFromContext(c, r.db).Where("tenant_id = ?", tenantID)
}

func tenantString(c context.Context) string {
	s, _ := ctx.TenantIDFromContext(c)
	return s
}

func (r *gormRoleRepository) Create(c context.Context, role *entity.Role) error {
	return database.DBFromContext(c, r.db).Create(role).Error
}

func (r *gormRoleRepository) first(c context.Context, where string, arg any) (*entity.Role, error) {
	var role entity.Role
	if err := r.scoped(c).Where(where, arg).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepository) GetByID(c context.Context, id uuid.UUID) (*entity.Role, error) {
	return r.first(c, "id = ?", id)
}

func (r *gormRoleRepository) GetByName(c context.Context, name string) (*entity.Role, error) {
	return r.first(c, "name = ?", name)
}

func (r *gormRoleRepository) List(c context.Context) ([]*entity.Role, error) {
	var roles []*entity.Role
	if err := r.scoped(c).Order("name").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *gormRoleRepository) Update(c context.Context, role *entity.Role) error {
	res := r.scoped(c).Model(&entity.Role{}).Where("id = ?", role.ID).
		Updates(map[string]any{"name": role.Name, "updated_at": role.UpdatedAt})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainerrors.ErrRoleNotFound
	}
	return nil
}

func (r *gormRoleRepository) Delete(c context.Context, id uuid.UUID) error {
	res := r.scoped(c).Where("id = ?", id).Delete(&entity.Role{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainerrors.ErrRoleNotFound
	}
	return nil
}
