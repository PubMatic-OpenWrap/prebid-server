package wakanda

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKeysFromHBRequest(t *testing.T) {
	keys := generateKeysFromHBRequest("31445", "225")
	assert.Equal(t, []string{"PUB:31445__PROF:225", "PUB:31445__PROF:0"}, keys)
}

func TestGenerateKeyFromWakandaRequest(t *testing.T) {
	assert.Equal(t, "PUB:31445__PROF:225", generateKeyFromWakandaRequest("31445", "225"))
	assert.Equal(t, "PUB:31445__PROF:0", generateKeyFromWakandaRequest("31445", ""))
	assert.Equal(t, "", generateKeyFromWakandaRequest("", ""))
}

func wakndaGetTester(t *testing.T, handler http.HandlerFunc, call string, output string) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", call, nil)
	if err != nil {
		t.Fatal(err)
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("For input Query: %s, handler returned wrong status code: \nGOT %v \nWANT %v",
			call, status, http.StatusOK)
	}
	if rr.Body.String() != output {
		t.Errorf("For input Query: %s, handler returned unexpected body:\nGOT %v \nWANT %v",
			call, rr.Body.String(), output)
	}
}

func TestHttpHandler(t *testing.T) {
	config := Wakanda{HostName: "", DCName: "DC1"}
	Init(config)
	handler := http.HandlerFunc(Handler(config))
	wakndaGetTester(t, handler, "/wakanda", `{success: "false", statusMsg: "No key was generated for the request.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/", `{success: "false", statusMsg: "No key was generated for the request.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/?pubId=100", `{success: "false", statusMsg: "No key was generated for the request.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/?pubId=100&profId=1", `{success: "false", statusMsg: "No key was generated for the request.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/?pubId=100&profId=1&debugLevel=2", `{success: "false", statusMsg: "No key was generated for the request.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/?pubId=1000&profId=1&debugLevel=2", `{success: "true", statusMsg: "New key generated.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/?pubId=1000&profId=1&debugLevel=2", `{success: "true", statusMsg: "Key already exists.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/?pubId=2000&profId=2&debugLevel=2", `{success: "true", statusMsg: "New key generated.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/?pubId=2000&profId=2&debugLevel=2", `{success: "true", statusMsg: "Key already exists.", host: ""}`)
	wakndaGetTester(t, handler, "/wakanda/?pubId=2000&profId=2&debugLevel=3", `{success: "true", statusMsg: "Key already exists.", host: ""}`)
}
