package category

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"squirrel/internal/policy"
	"keeper/pkg/auth"

	"github.com/go-chi/chi/v5"
)

func strPtr(s string) *string { return &s }

var testPolicyStore = policy.NewStoreFromPolicies(policy.Compile([]policy.Row{
	{Role: "sysadmin", IsSudo: true},
	{Role: "admin", Resource: strPtr("category"), Action: strPtr("create")},
	{Role: "admin", Resource: strPtr("category"), Action: strPtr("update")},
	{Role: "admin", Resource: strPtr("category"), Action: strPtr("delete")},
}))

type fakeCategoryService struct{ created, updated, deleted bool }

func (f *fakeCategoryService) Create(ctx context.Context, appID, userID, divisionID int, req CreateCategoryRequest) (Category, error) {
	f.created = true
	return Category{ID: 1}, nil
}
func (f *fakeCategoryService) List(ctx context.Context, appID int, divisionID *int, name string, limit, offset int) ([]Category, error) {
	return nil, nil
}
func (f *fakeCategoryService) GetByID(ctx context.Context, appID, id int) (Category, error) {
	return Category{ID: id}, nil
}
func (f *fakeCategoryService) Update(ctx context.Context, appID, id int, req UpdateCategoryRequest) (Category, error) {
	f.updated = true
	return Category{ID: id}, nil
}
func (f *fakeCategoryService) Delete(ctx context.Context, appID, id int) error {
	f.deleted = true
	return nil
}

func jsonRequestWithClaims(method, target string, body any, claims *auth.UserClaims) *http.Request {
	b, _ := json.Marshal(body)
	r := httptest.NewRequest(method, target, bytes.NewReader(b))
	ctx := context.WithValue(r.Context(), auth.UserClaimsKey, claims)
	return r.WithContext(ctx)
}

func withIDParam(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestHandler_Create_DeniedWithoutPermission(t *testing.T) {
	svc := &fakeCategoryService{}
	h := NewHandler(svc, testPolicyStore)

	req := jsonRequestWithClaims(http.MethodPost, "/categories", CreateCategoryRequest{Name: "Food"}, &auth.UserClaims{AppID: 1, Roles: []auth.RoleAssignment{{Name: "member"}}})
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
	if svc.created {
		t.Fatal("service.Create called despite denied permission")
	}
}

func TestHandler_Create_AllowedForGrantedRole(t *testing.T) {
	svc := &fakeCategoryService{}
	h := NewHandler(svc, testPolicyStore)

	req := jsonRequestWithClaims(http.MethodPost, "/categories", CreateCategoryRequest{Name: "Food"}, &auth.UserClaims{AppID: 1, Roles: []auth.RoleAssignment{{Name: "admin"}}})
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rr.Code)
	}
	if !svc.created {
		t.Fatal("service.Create not called despite granted permission")
	}
}

func TestHandler_Update_DeniedWithoutPermission(t *testing.T) {
	svc := &fakeCategoryService{}
	h := NewHandler(svc, testPolicyStore)

	req := withIDParam(jsonRequestWithClaims(http.MethodPut, "/categories/1", UpdateCategoryRequest{Name: "Food"}, &auth.UserClaims{AppID: 1, Roles: []auth.RoleAssignment{{Name: "member"}}}), "1")
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
	if svc.updated {
		t.Fatal("service.Update called despite denied permission")
	}
}

func TestHandler_Delete_DeniedWithoutPermission(t *testing.T) {
	svc := &fakeCategoryService{}
	h := NewHandler(svc, testPolicyStore)

	req := withIDParam(jsonRequestWithClaims(http.MethodDelete, "/categories/1", nil, &auth.UserClaims{AppID: 1, Roles: []auth.RoleAssignment{{Name: "member"}}}), "1")
	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
	if svc.deleted {
		t.Fatal("service.Delete called despite denied permission")
	}
}
