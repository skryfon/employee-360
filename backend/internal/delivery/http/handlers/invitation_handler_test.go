package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/delivery/http/response"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
	invtypes "github.com/skryfon/employee360/backend/internal/types/invitation"
)

// fakeInvUC implements all five invitation usecases with configurable results.
type fakeInvUC struct {
	inv       *entity.UserInvitation
	items     []*entity.UserInvitation
	total     int64
	err       error
	gotInvite invtypes.InviteUserRequest
	gotAccept invtypes.AcceptInvitationRequest
	gotToken  string
	gotID     uuid.UUID
	gotLimit  int
	gotOffset int
	gotTenant uuid.UUID
	gotActor  uuid.UUID
}

func (f *fakeInvUC) invite(_ context.Context, tenantID, actorID uuid.UUID, req invtypes.InviteUserRequest) (*entity.UserInvitation, error) {
	f.gotInvite = req
	f.gotTenant, f.gotActor = tenantID, actorID
	return f.inv, f.err
}

type fakeInviteUC struct{ *fakeInvUC }

func (f fakeInviteUC) Execute(c context.Context, tenantID, actorID uuid.UUID, req invtypes.InviteUserRequest) (*entity.UserInvitation, error) {
	return f.invite(c, tenantID, actorID, req)
}

type fakeAcceptUC struct{ *fakeInvUC }

func (f fakeAcceptUC) Execute(_ context.Context, req invtypes.AcceptInvitationRequest) error {
	f.gotAccept = req
	return f.err
}

type fakeResendUC struct{ *fakeInvUC }

func (f fakeResendUC) Execute(_ context.Context, tenantID, actorID, id uuid.UUID) (*entity.UserInvitation, error) {
	f.gotID = id
	f.gotTenant, f.gotActor = tenantID, actorID
	return f.inv, f.err
}

type fakeRevokeUC struct{ *fakeInvUC }

func (f fakeRevokeUC) Execute(_ context.Context, tenantID, actorID, id uuid.UUID) error {
	f.gotID = id
	f.gotTenant, f.gotActor = tenantID, actorID
	return f.err
}

type fakeListUC struct{ *fakeInvUC }

func (f fakeListUC) Execute(_ context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserInvitation, int64, error) {
	f.gotLimit, f.gotOffset = limit, offset
	f.gotTenant = tenantID
	return f.items, f.total, f.err
}

type fakeValidateUC struct{ *fakeInvUC }

func (f fakeValidateUC) Execute(_ context.Context, token string) (*invtypes.ValidateInvitationResponse, error) {
	f.gotToken = token
	if f.err != nil {
		return nil, f.err
	}
	return &invtypes.ValidateInvitationResponse{Email: "new@acme.com", Role: "employee"}, nil
}

// Identity the fake auth middleware injects for authenticated invitation tests.
var (
	testTenantID = uuid.New()
	testActorID  = uuid.New()
)

// withIdentity stands in for the Auth/Tenant middleware: it stores the caller's
// identity on the request context via the ctx package.
func withIdentity(tenantID, userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := ctx.WithTenantID(c.Request.Context(), tenantID.String())
		rc = ctx.WithUserID(rc, userID.String())
		c.Request = c.Request.WithContext(rc)
		c.Next()
	}
}

func setupInvitationHandlerTest() (*gin.Engine, *fakeInvUC) {
	return setupInvitationHandlerTestWith(withIdentity(testTenantID, testActorID))
}

// setupInvitationHandlerTestWith builds the engine with an optional identity middleware.
func setupInvitationHandlerTestWith(mw ...gin.HandlerFunc) (*gin.Engine, *fakeInvUC) {
	gin.SetMode(gin.TestMode)
	f := &fakeInvUC{inv: pendingInvitation()}
	h := NewInvitationHandler(fakeInviteUC{f}, fakeAcceptUC{f}, fakeResendUC{f}, fakeRevokeUC{f}, fakeListUC{f}, fakeValidateUC{f})
	engine := gin.New()
	engine.Use(mw...)
	engine.POST("/invitations", h.Invite)
	engine.GET("/invitations", h.List)
	engine.POST("/invitations/:id/resend", h.Resend)
	engine.DELETE("/invitations/:id", h.Revoke)
	engine.POST("/accept", h.Accept)
	engine.GET("/validate", h.Validate)
	return engine, f
}

func pendingInvitation() *entity.UserInvitation {
	return &entity.UserInvitation{
		ID: uuid.New(), Email: "new@acme.com", RoleID: uuid.New(), InvitedBy: uuid.New(),
		TokenHash: "secret-hash", ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}
}

func doInv(engine *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return env
}

func TestInvitationHandler_Invite(t *testing.T) {
	roleID := uuid.New()
	body := `{"email":"new@acme.com","role_id":"` + roleID.String() + `","first_name":"Ann"}`

	t.Run("success returns 201 and never exposes tokens", func(t *testing.T) {
		engine, f := setupInvitationHandlerTest()
		rec := doInv(engine, http.MethodPost, "/invitations", body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", rec.Code, rec.Body.String())
		}
		if f.gotInvite.Email != "new@acme.com" || f.gotInvite.RoleID != roleID || f.gotInvite.FirstName != "Ann" {
			t.Errorf("unexpected request passed to usecase: %+v", f.gotInvite)
		}
		raw := rec.Body.String()
		for _, leak := range []string{"token", "secret-hash", "tenant_id"} {
			if strings.Contains(raw, leak) {
				t.Errorf("response must not contain %q: %s", leak, raw)
			}
		}
		data := decodeEnvelope(t, rec).Data.(map[string]any)
		if data["status"] != "pending" {
			t.Errorf("want status pending, got %v", data["status"])
		}
	})

	t.Run("bad json", func(t *testing.T) {
		engine, _ := setupInvitationHandlerTest()
		if rec := doInv(engine, http.MethodPost, "/invitations", "not json"); rec.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d", rec.Code)
		}
	})

	t.Run("invalid role uuid", func(t *testing.T) {
		engine, _ := setupInvitationHandlerTest()
		rec := doInv(engine, http.MethodPost, "/invitations", `{"email":"a@b.com","role_id":"nope"}`)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d", rec.Code)
		}
	})
}

func TestInvitationHandler_Resend(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		engine, f := setupInvitationHandlerTest()
		rec := doInv(engine, http.MethodPost, "/invitations/"+f.inv.ID.String()+"/resend", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "token") {
			t.Errorf("response must not expose token: %s", rec.Body.String())
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		engine, _ := setupInvitationHandlerTest()
		if rec := doInv(engine, http.MethodPost, "/invitations/nope/resend", ""); rec.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d", rec.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		engine, f := setupInvitationHandlerTest()
		f.err = domainerrors.ErrInvitationNotFound
		if rec := doInv(engine, http.MethodPost, "/invitations/"+uuid.NewString()+"/resend", ""); rec.Code != http.StatusNotFound {
			t.Errorf("want 404, got %d", rec.Code)
		}
	})
}

func TestInvitationHandler_Revoke(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		engine, _ := setupInvitationHandlerTest()
		rec := doInv(engine, http.MethodDelete, "/invitations/"+uuid.NewString(), "")
		if rec.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if !decodeEnvelope(t, rec).Success {
			t.Error("want success=true")
		}
	})

	t.Run("invalid uuid", func(t *testing.T) {
		engine, _ := setupInvitationHandlerTest()
		if rec := doInv(engine, http.MethodDelete, "/invitations/nope", ""); rec.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d", rec.Code)
		}
	})

	t.Run("not pending", func(t *testing.T) {
		engine, f := setupInvitationHandlerTest()
		f.err = domainerrors.ErrInvitationNotPending
		if rec := doInv(engine, http.MethodDelete, "/invitations/"+uuid.NewString(), ""); rec.Code != http.StatusConflict {
			t.Errorf("want 409, got %d", rec.Code)
		}
	})
}

func TestInvitationHandler_List(t *testing.T) {
	now := time.Now()
	past, future := now.Add(-time.Hour), now.Add(time.Hour)
	statusOf := func(i *entity.UserInvitation) string { return string(i.Status(now)) }

	t.Run("derived status and no token leak", func(t *testing.T) {
		engine, f := setupInvitationHandlerTest()
		f.items = []*entity.UserInvitation{
			{ID: uuid.New(), TokenHash: "h1", ExpiresAt: future},
			{ID: uuid.New(), TokenHash: "h2", ExpiresAt: future, AcceptedAt: &past},
			{ID: uuid.New(), TokenHash: "h3", ExpiresAt: future, RevokedAt: &past},
			{ID: uuid.New(), TokenHash: "h4", ExpiresAt: past},
		}
		f.total = 4
		rec := doInv(engine, http.MethodGet, "/invitations", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "token") || strings.Contains(rec.Body.String(), "h1") {
			t.Errorf("response must not expose token data: %s", rec.Body.String())
		}
		env := decodeEnvelope(t, rec)
		items := env.Data.([]any)
		want := []string{"pending", "accepted", "revoked", "expired"}
		for i, it := range items {
			if got := it.(map[string]any)["status"]; got != want[i] {
				t.Errorf("item %d: want status %s, got %v (entity says %s)", i, want[i], got, statusOf(f.items[i]))
			}
		}
		if env.Meta == nil || env.Meta.TotalItems != 4 || env.Meta.TotalPages != 1 {
			t.Errorf("unexpected meta: %+v", env.Meta)
		}
	})

	t.Run("empty list is an array", func(t *testing.T) {
		engine, _ := setupInvitationHandlerTest()
		rec := doInv(engine, http.MethodGet, "/invitations", "")
		if !strings.Contains(rec.Body.String(), `"data":[]`) {
			t.Errorf("want empty array, got %s", rec.Body.String())
		}
	})

	paging := []struct {
		name, query         string
		limit, offset, page int
	}{
		{"defaults", "", 20, 0, 1},
		{"explicit page", "?page=3&page_size=10", 10, 20, 3},
		{"max page size", "?page_size=100", 100, 0, 1},
		{"over max falls back to default", "?page_size=101", 20, 0, 1},
		{"zero page clamps to 1", "?page=0", 20, 0, 1},
		{"garbage falls back", "?page=x&page_size=y", 20, 0, 1},
	}
	for _, tt := range paging {
		t.Run(tt.name, func(t *testing.T) {
			engine, f := setupInvitationHandlerTest()
			f.total = 250
			rec := doInv(engine, http.MethodGet, "/invitations"+tt.query, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("want 200, got %d", rec.Code)
			}
			if f.gotLimit != tt.limit || f.gotOffset != tt.offset {
				t.Errorf("usecase got limit=%d offset=%d, want %d/%d", f.gotLimit, f.gotOffset, tt.limit, tt.offset)
			}
			if meta := decodeEnvelope(t, rec).Meta; meta.Page != tt.page || meta.PageSize != tt.limit {
				t.Errorf("unexpected meta: %+v", meta)
			}
		})
	}
}

func TestInvitationHandler_Accept(t *testing.T) {
	good := `{"token":"abc","password":"password123"}`

	t.Run("success", func(t *testing.T) {
		engine, f := setupInvitationHandlerTest()
		rec := doInv(engine, http.MethodPost, "/accept", good)
		if rec.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if f.gotAccept.Token != "abc" || f.gotAccept.Password != "password123" {
			t.Errorf("unexpected request passed to usecase: %+v", f.gotAccept)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		engine, _ := setupInvitationHandlerTest()
		if rec := doInv(engine, http.MethodPost, "/accept", "{"); rec.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d", rec.Code)
		}
	})

	for name, tt := range map[string]struct {
		err  error
		code int
	}{
		"invalid token":    {domainerrors.ErrInvalidToken, http.StatusBadRequest},
		"weak password":    {domainerrors.ErrInvalidPassword, http.StatusBadRequest},
		"unexpected error": {errors.New("boom"), http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			engine, f := setupInvitationHandlerTest()
			f.err = tt.err
			rec := doInv(engine, http.MethodPost, "/accept", good)
			if rec.Code != tt.code {
				t.Errorf("want %d, got %d", tt.code, rec.Code)
			}
			if strings.Contains(rec.Body.String(), "boom") {
				t.Errorf("internal error details leaked: %s", rec.Body.String())
			}
		})
	}
}

// TestWriteInvitationError_Mapping covers every branch via the Resend endpoint.
func TestWriteInvitationError_Mapping(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{domainerrors.ErrUnauthorized, http.StatusUnauthorized},
		{domainerrors.ErrForbidden, http.StatusForbidden},
		{domainerrors.ErrInvitationNotFound, http.StatusNotFound},
		{domainerrors.ErrInvitationNotPending, http.StatusConflict},
		{domainerrors.ErrEmailAlreadyExists, http.StatusConflict},
		{domainerrors.ErrRoleNotFound, http.StatusBadRequest},
		{domainerrors.ErrDepartmentNotFound, http.StatusBadRequest},
		{domainerrors.ErrPositionNotFound, http.StatusBadRequest},
		{domainerrors.ErrInvalidRole, http.StatusBadRequest},
		{domainerrors.ErrInvalidEmail, http.StatusBadRequest},
		{errors.New("db exploded"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			engine, f := setupInvitationHandlerTest()
			f.err = tt.err
			rec := doInv(engine, http.MethodPost, "/invitations", `{"email":"a@b.com"}`)
			if rec.Code != tt.code {
				t.Errorf("want %d, got %d: %s", tt.code, rec.Code, rec.Body.String())
			}
			if rec.Code == http.StatusInternalServerError && strings.Contains(rec.Body.String(), "db exploded") {
				t.Error("internal error message leaked")
			}
		})
	}
}

func TestInvitationHandler_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		engine, f := setupInvitationHandlerTest()
		rec := doInv(engine, http.MethodGet, "/validate?token=abc", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if f.gotToken != "abc" {
			t.Errorf("want token abc, got %q", f.gotToken)
		}
		data := decodeEnvelope(t, rec).Data.(map[string]any)
		if data["email"] != "new@acme.com" || data["role"] != "employee" {
			t.Errorf("unexpected payload: %+v", data)
		}
	})

	t.Run("missing token query param", func(t *testing.T) {
		engine, _ := setupInvitationHandlerTest()
		rec := doInv(engine, http.MethodGet, "/validate", "")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d", rec.Code)
		}
	})

	t.Run("invalid or expired token error", func(t *testing.T) {
		engine, f := setupInvitationHandlerTest()
		f.err = domainerrors.ErrInvalidToken
		rec := doInv(engine, http.MethodGet, "/validate?token=expired", "")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d", rec.Code)
		}
	})
}

func TestInvitationHandler_FailureStateCodes(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{domainerrors.ErrInvitationExpired, http.StatusGone, "INVITATION_EXPIRED"},
		{domainerrors.ErrInvitationRevoked, http.StatusForbidden, "INVITATION_REVOKED"},
		{domainerrors.ErrInvitationAccepted, http.StatusConflict, "INVITATION_ACCEPTED"},
		{domainerrors.ErrInvalidToken, http.StatusBadRequest, "INVALID_TOKEN"},
	}
	seen := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			for _, req := range []struct{ method, path, body string }{
				{http.MethodGet, "/validate?token=abc", ""},
				{http.MethodPost, "/accept", `{"token":"abc","password":"password123"}`},
			} {
				engine, f := setupInvitationHandlerTest()
				f.err = tc.err
				rec := doInv(engine, req.method, req.path, req.body)
				if rec.Code != tc.status {
					t.Fatalf("%s: want %d, got %d", req.path, tc.status, rec.Code)
				}
				env := decodeEnvelope(t, rec)
				if env.Error == nil || env.Error.Code != tc.code || env.Error.Message == "" {
					t.Fatalf("%s: want code %s, got %+v", req.path, tc.code, env.Error)
				}
			}
			engine, f := setupInvitationHandlerTest()
			f.err = tc.err
			msg := decodeEnvelope(t, doInv(engine, http.MethodGet, "/validate?token=abc", "")).Error.Message
			if seen[msg] {
				t.Errorf("message not distinct: %q", msg)
			}
			seen[msg] = true
		})
	}
}

func TestInvitationHandler_ValidateMissingTokenCode(t *testing.T) {
	engine, _ := setupInvitationHandlerTest()
	env := decodeEnvelope(t, doInv(engine, http.MethodGet, "/validate", ""))
	if env.Error == nil || env.Error.Code != "INVALID_TOKEN" {
		t.Errorf("want INVALID_TOKEN, got %+v", env.Error)
	}
}

func TestInvitationHandler_PassesIdentityToUsecase(t *testing.T) {
	engine, f := setupInvitationHandlerTest()
	id := uuid.New().String()

	rec := doInv(engine, http.MethodPost, "/invitations/"+id+"/resend", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("resend: want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if f.gotTenant != testTenantID || f.gotActor != testActorID {
		t.Errorf("resend: usecase got tenant %s actor %s", f.gotTenant, f.gotActor)
	}
	f.gotTenant, f.gotActor = uuid.Nil, uuid.Nil
	doInv(engine, http.MethodDelete, "/invitations/"+id, "")
	if f.gotTenant != testTenantID || f.gotActor != testActorID {
		t.Errorf("revoke: usecase got tenant %s actor %s", f.gotTenant, f.gotActor)
	}
	f.gotTenant = uuid.Nil
	doInv(engine, http.MethodGet, "/invitations", "")
	if f.gotTenant != testTenantID {
		t.Errorf("list: usecase got tenant %s", f.gotTenant)
	}
}

func TestInvitationHandler_MissingIdentityIs401(t *testing.T) {
	engine, f := setupInvitationHandlerTestWith() // no identity middleware
	id := uuid.New().String()
	body := `{"email":"new@acme.com","role_id":"` + uuid.New().String() + `"}`
	for name, req := range map[string][3]string{
		"invite": {http.MethodPost, "/invitations", body},
		"list":   {http.MethodGet, "/invitations", ""},
		"resend": {http.MethodPost, "/invitations/" + id + "/resend", ""},
		"revoke": {http.MethodDelete, "/invitations/" + id, ""},
	} {
		t.Run(name, func(t *testing.T) {
			rec := doInv(engine, req[0], req[1], req[2])
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("want 401, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
	if f.gotTenant != uuid.Nil || f.gotActor != uuid.Nil || f.gotInvite.Email != "" {
		t.Error("usecase must not be called without identity")
	}
}
