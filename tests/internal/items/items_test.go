package items

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/boilerplate-api/go-rest-api-starter/tests/testsuite"
)

func TestIntegrations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Items Test Suite")
}

var _ = BeforeEach(func() {
	_ = testsuite.GetTestApp()
	Expect(testsuite.ResetDB()).To(Succeed())
})

var _ = Describe("Items API", func() {

	Describe("POST /api/v1/items", func() {

		It("creates a new item when authenticated and returns 201 Created", func() {
			app := testsuite.GetTestApp()

			requestBody := map[string]any{
				"name": "Test Item",
			}

			payload, err := json.Marshal(requestBody)
			Expect(err).ToNot(HaveOccurred())

			req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			for _, ck := range testsuite.GetAuthCookies() {
				req.AddCookie(ck)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusCreated),
				"POST /api/v1/items should return 201 Created on success")

			var responseBody map[string]any
			Expect(json.NewDecoder(res.Body).Decode(&responseBody)).To(Succeed(),
				"response should be valid JSON")

			Expect(responseBody).To(HaveKey("id"))
			Expect(responseBody).To(HaveKeyWithValue("name", "Test Item"))
		})

		It("creates a new item with arrivalDate and returns 201 Created", func() {
			app := testsuite.GetTestApp()

			requestBody := map[string]any{
				"name":        "Test Item with Date",
				"arrivalDate": "20240215",
			}

			payload, err := json.Marshal(requestBody)
			Expect(err).ToNot(HaveOccurred())

			req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			for _, ck := range testsuite.GetAuthCookies() {
				req.AddCookie(ck)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusCreated),
				"POST /api/v1/items should return 201 Created on success")

			var responseBody map[string]any
			Expect(json.NewDecoder(res.Body).Decode(&responseBody)).To(Succeed(),
				"response should be valid JSON")

			Expect(responseBody).To(HaveKey("id"))
			Expect(responseBody).To(HaveKeyWithValue("name", "Test Item with Date"))
			Expect(responseBody).To(HaveKey("arrivalDate"))
		})

		It("returns 401 Unauthorized when not authenticated", func() {
			app := testsuite.GetTestApp()

			requestBody := map[string]any{
				"name": "Test Item",
			}

			payload, err := json.Marshal(requestBody)
			Expect(err).ToNot(HaveOccurred())

			req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized),
				"POST /api/v1/items should return 401 Unauthorized without auth")
		})

		It("returns 400 Bad Request when name is missing", func() {
			app := testsuite.GetTestApp()

			requestBody := map[string]any{}

			payload, err := json.Marshal(requestBody)
			Expect(err).ToNot(HaveOccurred())

			req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			for _, ck := range testsuite.GetAuthCookies() {
				req.AddCookie(ck)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusBadRequest),
				"POST /api/v1/items should return 400 Bad Request when validation fails")
		})
	})

	Describe("GET /api/v1/items", func() {

		It("returns list of items when authenticated", func() {
			app := testsuite.GetTestApp()

			// First, create an item
			createReqBody := map[string]any{"name": "Test Item"}
			createPayload, _ := json.Marshal(createReqBody)
			createReq := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(createPayload))
			createReq.Header.Set("Content-Type", "application/json")
			for _, ck := range testsuite.GetAuthCookies() {
				createReq.AddCookie(ck)
			}
			createRec := httptest.NewRecorder()
			app.ServeHTTP(createRec, createReq)

			// Now list items
			req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
			for _, ck := range testsuite.GetAuthCookies() {
				req.AddCookie(ck)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusOK),
				"GET /api/v1/items should return 200 OK")

			var responseBody []map[string]any
			Expect(json.NewDecoder(res.Body).Decode(&responseBody)).To(Succeed(),
				"response should be valid JSON array")

			Expect(responseBody).To(HaveLen(1),
				"should return one item")
			Expect(responseBody[0]).To(HaveKeyWithValue("name", "Test Item"))
		})

		It("returns 401 Unauthorized when not authenticated", func() {
			app := testsuite.GetTestApp()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized),
				"GET /api/v1/items should return 401 Unauthorized without auth")
		})
	})

	Describe("GET /api/v1/items/:id", func() {

		It("returns item by ID when authenticated", func() {
			app := testsuite.GetTestApp()

			// First, create an item
			createReqBody := map[string]any{"name": "Test Item"}
			createPayload, _ := json.Marshal(createReqBody)
			createReq := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(createPayload))
			createReq.Header.Set("Content-Type", "application/json")
			for _, ck := range testsuite.GetAuthCookies() {
				createReq.AddCookie(ck)
			}
			createRec := httptest.NewRecorder()
			app.ServeHTTP(createRec, createReq)

			var createResp map[string]any
			json.NewDecoder(createRec.Result().Body).Decode(&createResp)
			itemID := createResp["id"].(string)

			// Now get the item
			req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+itemID, nil)
			for _, ck := range testsuite.GetAuthCookies() {
				req.AddCookie(ck)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusOK),
				"GET /api/v1/items/:id should return 200 OK")

			var responseBody map[string]any
			Expect(json.NewDecoder(res.Body).Decode(&responseBody)).To(Succeed(),
				"response should be valid JSON")

			Expect(responseBody).To(HaveKeyWithValue("id", itemID))
			Expect(responseBody).To(HaveKeyWithValue("name", "Test Item"))
		})

		It("returns 404 Not Found when item does not exist", func() {
			app := testsuite.GetTestApp()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/items/00000000-0000-0000-0000-000000000000", nil)
			for _, ck := range testsuite.GetAuthCookies() {
				req.AddCookie(ck)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusNotFound),
				"GET /api/v1/items/:id should return 404 Not Found for non-existent item")
		})

		It("returns 401 Unauthorized when not authenticated", func() {
			app := testsuite.GetTestApp()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/items/00000000-0000-0000-0000-000000000000", nil)

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized),
				"GET /api/v1/items/:id should return 401 Unauthorized without auth")
		})
	})

	Describe("PATCH /api/v1/items/:id", func() {

		It("updates item with arrivalDate when authenticated and returns 200 OK", func() {
			app := testsuite.GetTestApp()

			// First, create an item
			createReqBody := map[string]any{"name": "Test Item"}
			createPayload, _ := json.Marshal(createReqBody)
			createReq := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(createPayload))
			createReq.Header.Set("Content-Type", "application/json")
			for _, ck := range testsuite.GetAuthCookies() {
				createReq.AddCookie(ck)
			}
			createRec := httptest.NewRecorder()
			app.ServeHTTP(createRec, createReq)

			var createResp map[string]any
			json.NewDecoder(createRec.Result().Body).Decode(&createResp)
			itemID := createResp["id"].(string)

			// Now update the item with arrivalDate
			updateReqBody := map[string]any{
				"name":        "Updated Item",
				"arrivalDate": "20240220",
			}
			updatePayload, _ := json.Marshal(updateReqBody)
			updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/items/"+itemID, bytes.NewReader(updatePayload))
			updateReq.Header.Set("Content-Type", "application/json")
			for _, ck := range testsuite.GetAuthCookies() {
				updateReq.AddCookie(ck)
			}

			updateRec := httptest.NewRecorder()
			app.ServeHTTP(updateRec, updateReq)

			updateRes := updateRec.Result()
			defer updateRes.Body.Close()

			Expect(updateRes.StatusCode).To(Equal(http.StatusOK),
				"PATCH /api/v1/items/:id should return 200 OK")

			var updateResp map[string]any
			Expect(json.NewDecoder(updateRes.Body).Decode(&updateResp)).To(Succeed(),
				"response should be valid JSON")

			Expect(updateResp).To(HaveKeyWithValue("id", itemID))
			Expect(updateResp).To(HaveKeyWithValue("name", "Updated Item"))
			Expect(updateResp).To(HaveKey("arrivalDate"))
		})
	})

	Describe("DELETE /api/v1/items/:id", func() {

		It("deletes item by ID when authenticated and returns 204 No Content", func() {
			app := testsuite.GetTestApp()

			// First, create an item
			createReqBody := map[string]any{"name": "Test Item"}
			createPayload, _ := json.Marshal(createReqBody)
			createReq := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(createPayload))
			createReq.Header.Set("Content-Type", "application/json")
			for _, ck := range testsuite.GetAuthCookies() {
				createReq.AddCookie(ck)
			}
			createRec := httptest.NewRecorder()
			app.ServeHTTP(createRec, createReq)

			var createResp map[string]any
			json.NewDecoder(createRec.Result().Body).Decode(&createResp)
			itemID := createResp["id"].(string)

			// Now delete the item
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/"+itemID, nil)
			for _, ck := range testsuite.GetAuthCookies() {
				req.AddCookie(ck)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusNoContent),
				"DELETE /api/v1/items/:id should return 204 No Content")

			// Verify item is deleted
			getReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+itemID, nil)
			for _, ck := range testsuite.GetAuthCookies() {
				getReq.AddCookie(ck)
			}
			getRec := httptest.NewRecorder()
			app.ServeHTTP(getRec, getReq)

			Expect(getRec.Result().StatusCode).To(Equal(http.StatusNotFound),
				"item should be deleted")
		})

		It("returns 204 No Content when deleting non-existent item (idempotent)", func() {
			app := testsuite.GetTestApp()

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/00000000-0000-0000-0000-000000000000", nil)
			for _, ck := range testsuite.GetAuthCookies() {
				req.AddCookie(ck)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusNoContent),
				"DELETE /api/v1/items/:id should return 204 No Content for non-existent item (idempotent)")
		})

		It("returns 401 Unauthorized when not authenticated", func() {
			app := testsuite.GetTestApp()

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/items/00000000-0000-0000-0000-000000000000", nil)

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized),
				"DELETE /api/v1/items/:id should return 401 Unauthorized without auth")
		})
	})
})
