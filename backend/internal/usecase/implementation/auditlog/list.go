package auditlog

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	"github.com/skryfon/employee360/backend/internal/domain/repository"
	altypes "github.com/skryfon/employee360/backend/internal/types/auditlog"
	aluc "github.com/skryfon/employee360/backend/internal/usecase/interface/auditlog"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

const (
	maxActionLen     = 100
	maxEntityTypeLen = 100
)

type listAuditLogsUseCase struct {
	repo repository.AuditLogQueryRepository
}

// NewListAuditLogsUseCase creates a new ListAuditLogsUseCase.
func NewListAuditLogsUseCase(repo repository.AuditLogQueryRepository) aluc.ListAuditLogsUseCase {
	return &listAuditLogsUseCase{repo: repo}
}

func (uc *listAuditLogsUseCase) Execute(c context.Context, tenantID uuid.UUID, input altypes.ListAuditLogsQuery) (*altypes.ListAuditLogsResult, error) {
	if tenantID == uuid.Nil {
		return nil, domainerrors.ErrUnauthorized
	}

	if input.From != nil && input.To != nil && input.From.After(*input.To) {
		return nil, fmt.Errorf("%w: from must not be after to", domainerrors.ErrInvalidAuditLogFilter)
	}
	action := strings.TrimSpace(input.Action)
	entityType := strings.TrimSpace(input.EntityType)
	if utf8.RuneCountInString(action) > maxActionLen || utf8.RuneCountInString(entityType) > maxEntityTypeLen {
		return nil, fmt.Errorf("%w: filter value too long", domainerrors.ErrInvalidAuditLogFilter)
	}

	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	filter := repository.AuditLogFilter{
		Action:       action,
		ActionPrefix: strings.HasSuffix(action, "."),
		ActorUserID:  input.ActorUserID,
		EntityType:   entityType,
		From:         input.From,
		To:           input.To,
	}

	items, total, err := uc.repo.List(c, tenantID, filter, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	for _, e := range items {
		e.Metadata = sanitizeMetadata(e.Metadata)
	}
	return &altypes.ListAuditLogsResult{Entries: items, Total: total, Page: page, PageSize: pageSize}, nil
}
