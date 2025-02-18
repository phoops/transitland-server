package resolvers

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestMain(m *testing.M) {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		fmt.Println("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		return
	}
	model.DB = model.MustOpenDB(g)
	os.Exit(m.Run())
}

// Test helpers

func newTestClient() *client.Client {
	srv, _ := NewServer(config.Config{})
	return client.New(srv)
}

func toJson(m map[string]interface{}) string {
	rr, _ := json.Marshal(&m)
	return string(rr)
}

type hw = map[string]interface{}

type testcase struct {
	name         string
	query        string
	vars         hw
	expect       string
	selector     string
	expectSelect []string
}

func testquery(t *testing.T, c *client.Client, tc testcase) {
	var resp map[string]interface{}
	opts := []client.Option{}
	for k, v := range tc.vars {
		opts = append(opts, client.Var(k, v))
	}
	c.MustPost(tc.query, &resp, opts...)
	jj := toJson(resp)
	if tc.expect != "" {
		if !assert.JSONEq(t, tc.expect, jj) {
			fmt.Printf("got %s -- expect %s\n", jj, tc.expect)
		}
	}
	if tc.selector != "" {
		a := []string{}
		for _, v := range gjson.Get(jj, tc.selector).Array() {
			a = append(a, v.String())
		}
		if len(a) == 0 && tc.expectSelect == nil {
			t.Errorf("selector '%s' returned zero elements", tc.selector)
		} else {
			if !assert.ElementsMatch(t, a, tc.expectSelect) {
				fmt.Printf("got %#v -- expect %#v\n\n", a, tc.expectSelect)
			}
		}
	}
}
