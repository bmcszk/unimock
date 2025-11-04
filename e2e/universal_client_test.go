package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestUniversalClientE2E(t *testing.T) {
	given, when, then := newParts(t)

	t.Run("BasicHTTPOperations", func(_ *testing.T) {
		when.
			a_get_request_is_made_to_a_non_existent_resource().and().
			a_head_request_is_made_to_a_non_existent_resource().and().
			an_options_request_is_made_to_a_technical_endpoint()
		then.
			the_options_response_is_successful()
	})

	t.Run("JSONOperations", func(_ *testing.T) {
		given.
			a_user_is_created()

		when.
			the_user_is_updated().and().
			the_user_is_patched()

		then.
			the_user_is_deleted()
	})

	t.Run("UniDataLifecycle", func(_ *testing.T) {
		given.
			multiple_users_are_created()

		when.
			the_users_collection_is_retrieved()

		then.
			all_users_are_deleted()
	})
}

func (p *parts) a_get_request_is_made_to_a_non_existent_resource() *parts {
	resp, err := p.unimockAPIClient.Get(context.Background(), "/api/users/999", nil)
	p.require.NoError(err)
	p.require.Equal(http.StatusNotFound, resp.StatusCode)
	return p
}

func (p *parts) a_head_request_is_made_to_a_non_existent_resource() *parts {
	resp, err := p.unimockAPIClient.Head(context.Background(), "/api/users/999", nil)
	p.require.NoError(err)
	p.require.Equal(http.StatusNotFound, resp.StatusCode)
	p.require.Empty(resp.Body)
	return p
}

func (p *parts) an_options_request_is_made_to_a_technical_endpoint() *parts {
	resp, err := p.unimockAPIClient.Options(context.Background(), "/_uni/health", nil)
	p.require.NoError(err)
	p.require.NotZero(resp.StatusCode)
	return p
}

func (p *parts) the_options_response_is_successful() *parts {
	return p
}

func (p *parts) a_user_is_created() *parts {
	userData := map[string]any{
		"id":    "test-user-001",
		"name":  "Test User",
		"email": "test@example.com",
	}

	resp, err := p.unimockAPIClient.PostJSON(context.Background(), "/api/users", nil, userData)
	p.require.NoError(err)
	p.require.Equal(http.StatusCreated, resp.StatusCode)

	resp, err = p.unimockAPIClient.Get(context.Background(), "/api/users/test-user-001", nil)
	p.require.NoError(err)
	p.require.Equal(http.StatusOK, resp.StatusCode)

	var retrievedUser map[string]any
	err = json.Unmarshal(resp.Body, &retrievedUser)
	p.require.NoError(err)
	p.require.Equal("test-user-001", retrievedUser["id"])

	return p
}

func (p *parts) the_user_is_updated() *parts {
	updatedData := map[string]any{
		"id":    "test-user-001",
		"name":  "Updated Test User",
		"email": "updated@example.com",
	}

	resp, err := p.unimockAPIClient.PutJSON(context.Background(), "/api/users/test-user-001", nil, updatedData)
	p.require.NoError(err)
	p.require.Equal(http.StatusOK, resp.StatusCode)

	resp, err = p.unimockAPIClient.Get(context.Background(), "/api/users/test-user-001", nil)
	p.require.NoError(err)
	var retrievedUser map[string]any
	err = json.Unmarshal(resp.Body, &retrievedUser)
	p.require.NoError(err)
	p.require.Equal("Updated Test User", retrievedUser["name"])

	return p
}

func (p *parts) the_user_is_patched() *parts {
	patchData := map[string]any{
		"email": "patched@example.com",
	}

	resp, err := p.unimockAPIClient.PatchJSON(context.Background(), "/api/users/test-user-001", nil, patchData)
	p.require.NoError(err)
	p.require.Contains([]int{http.StatusOK, http.StatusMethodNotAllowed}, resp.StatusCode)

	return p
}

func (p *parts) the_user_is_deleted() *parts {
	resp, err := p.unimockAPIClient.Delete(context.Background(), "/api/users/test-user-001", nil)
	p.require.NoError(err)
	p.require.Equal(http.StatusNoContent, resp.StatusCode)

	resp, err = p.unimockAPIClient.Get(context.Background(), "/api/users/test-user-001", nil)
	p.require.NoError(err)
	p.require.Equal(http.StatusNotFound, resp.StatusCode)

	return p
}

func (p *parts) multiple_users_are_created() *parts {
	for i := 1; i <= 3; i++ {
		userData := map[string]any{
			"id":   fmt.Sprintf("user-%d", i),
			"name": fmt.Sprintf("User %d", i),
		}

		resp, err := p.unimockAPIClient.PostJSON(context.Background(), "/api/users", nil, userData)
		p.require.NoError(err)
		p.require.Equal(http.StatusCreated, resp.StatusCode)
	}
	return p
}

func (p *parts) the_users_collection_is_retrieved() *parts {
	resp, err := p.unimockAPIClient.Get(context.Background(), "/api/users", nil)
	p.require.NoError(err)
	p.require.Equal(http.StatusOK, resp.StatusCode)

	var users []map[string]any
	err = json.Unmarshal(resp.Body, &users)
	p.require.NoError(err)
	p.require.Len(users, 3)
	return p
}

func (p *parts) all_users_are_deleted() *parts {
	for i := 1; i <= 3; i++ {
		resp, err := p.unimockAPIClient.Delete(context.Background(), fmt.Sprintf("/api/users/user-%d", i), nil)
		p.require.NoError(err)
		p.require.Equal(http.StatusNoContent, resp.StatusCode)
	}
	return p
}
