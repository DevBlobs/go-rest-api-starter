package tests

import (
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
	RunSpecs(t, "Test Suite")
}

var _ = BeforeEach(func() {
	_ = testsuite.GetTestApp() // ensure DB/container/env are ready
	Expect(testsuite.ResetDB()).To(Succeed())
})

var _ = Describe("Health endpoint", func() {

	It("returns 200 OK and a descriptive JSON body", func() {
		app := testsuite.GetTestApp()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)

		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		Expect(res.StatusCode).To(Equal(http.StatusOK), "health endpoint must respond with 200 OK")

		var body map[string]any
		Expect(json.NewDecoder(res.Body).Decode(&body)).To(Succeed(), "response should be valid JSON")
		Expect(body).To(HaveKeyWithValue("status", "ok"), "health response should contain status=ok")
	})
})
