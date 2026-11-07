package http_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appACL "go-ddd/internal/application/acl"
	domainACL "go-ddd/internal/domain/acl"
	security "go-ddd/internal/infrastructure/security"
	httpapi "go-ddd/internal/presentation/http"
)

type nonAdminACL struct{ domainACL.Repository }

func (nonAdminACL) IsAuthAdmin(context.Context, int64) (bool, error) { return false, nil }
func (nonAdminACL) ListAdminDomains(context.Context, int64) ([]domainACL.Domain, error) {
	return nil, nil
}

func TestAllACLRoutesRejectAnonymousAndNonAdmin(t *testing.T) {
	jwt := security.NewJWTService("test-secret", time.Hour)
	token, err := jwt.GenerateToken(9, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	h := httpapi.NewRouter(nil, jwt, httpapi.NewACLHandler(appACL.NewService(nonAdminACL{})))
	for _, route := range []struct{ method, path string }{
		{"GET", "domains"}, {"POST", "domains"}, {"GET", "domains/1"}, {"PUT", "domains/1"}, {"DELETE", "domains/1"},
		{"GET", "roles"}, {"POST", "roles"}, {"GET", "roles/1"}, {"PUT", "roles/1"}, {"DELETE", "roles/1"},
		{"GET", "permissions"}, {"POST", "permissions"}, {"GET", "permissions/1"}, {"PUT", "permissions/1"}, {"DELETE", "permissions/1"},
	} {
		for _, tc := range []struct {
			name, token string
			want        int
		}{{"anonymous", "", 401}, {"non-admin", token, 403}} {
			r := httptest.NewRequest(route.method, "/api/v1/acl/"+route.path, nil)
			if tc.token != "" {
				r.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Errorf("%s %s %s: got %d want %d", route.method, route.path, tc.name, w.Code, tc.want)
			}
		}
	}
}

type scopedACL struct{ domainACL.Repository }

func (scopedACL) IsAuthAdmin(_ context.Context, id int64) (bool, error) { return id == 1, nil }
func (scopedACL) IsDomainAdmin(_ context.Context, id, domainID int64) (bool, error) {
	return id == 2 && domainID == 10, nil
}
func (scopedACL) ListAdminDomains(_ context.Context, id int64) ([]domainACL.Domain, error) {
	if id == 2 {
		return []domainACL.Domain{{ID: 10, Name: "One", Slug: "one"}}, nil
	}
	return nil, nil
}
func (scopedACL) GetDomain(_ context.Context, id int64) (*domainACL.Domain, error) {
	return &domainACL.Domain{ID: id, Name: "Domain", Slug: "domain"}, nil
}
func (scopedACL) ListRoles(_ context.Context, domainID int64) ([]domainACL.Role, error) {
	return []domainACL.Role{{ID: 100, DomainID: domainID, Code: "member"}}, nil
}
func (scopedACL) GetRole(_ context.Context, id int64) (*domainACL.Role, error) {
	if id == 101 {
		return &domainACL.Role{ID: id, DomainID: 20, Code: "member"}, nil
	}
	if id == 102 {
		return &domainACL.Role{ID: id, DomainID: 10, Code: "admin"}, nil
	}
	return &domainACL.Role{ID: id, DomainID: 10, Code: "member"}, nil
}
func (scopedACL) CreateRole(_ context.Context, role *domainACL.Role) error { role.ID = 103; return nil }

func TestDomainAdministratorScope(t *testing.T) {
	jwt := security.NewJWTService("test-secret", time.Hour)
	admin, _ := jwt.GenerateToken(2, "domain@example.com")
	super, _ := jwt.GenerateToken(1, "super@example.com")
	h := httpapi.NewRouter(nil, jwt, httpapi.NewACLHandler(appACL.NewService(scopedACL{})))
	cases := []struct {
		name, method, path, body, token string
		want                            int
	}{
		{"list own domains", "GET", "/api/v1/acl/domains", "", admin, 200},
		{"own domain", "GET", "/api/v1/acl/domains/10", "", admin, 200},
		{"other domain", "GET", "/api/v1/acl/domains/20", "", admin, 403},
		{"own roles", "GET", "/api/v1/acl/roles?domain_id=10", "", admin, 200},
		{"other roles", "GET", "/api/v1/acl/roles?domain_id=20", "", admin, 403},
		{"own role", "GET", "/api/v1/acl/roles/100", "", admin, 200},
		{"other role", "GET", "/api/v1/acl/roles/101", "", admin, 403},
		{"create own role", "POST", "/api/v1/acl/roles", `{"domain_id":10,"name":"Viewer","code":"viewer"}`, admin, 201},
		{"create cross-domain role", "POST", "/api/v1/acl/roles", `{"domain_id":20,"name":"Viewer","code":"viewer"}`, admin, 403},
		{"move role", "PUT", "/api/v1/acl/roles/100", `{"domain_id":20,"name":"Viewer","code":"viewer"}`, admin, 403},
		{"protect admin role", "DELETE", "/api/v1/acl/roles/102", "", admin, 403},
		{"protect admin role from super", "DELETE", "/api/v1/acl/roles/102", "", super, 403},
		{"global domain creation", "POST", "/api/v1/acl/domains", "{}", admin, 403},
		{"global permission creation", "POST", "/api/v1/acl/permissions", "{}", admin, 403},
		{"super other domain", "GET", "/api/v1/acl/domains/20", "", super, 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			r.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("got %d, want %d: %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}
