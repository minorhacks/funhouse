package github

import (
	"encoding/json"
  "io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
  "github.com/bazelbuild/rules_go/go/tools/bazel"
)

func mustReadFile(t *testing.T, filename string) []byte {
  t.Helper()
  f, err := bazel.Runfile(filename)
  if err != nil {
    t.Fatalf("can't get runfile %q: %v", filename, err)
  }
  contents, err := ioutil.ReadFile(f)
  if err != nil {
    t.Fatal(err)
  }
  return contents
}

func TestUnmarshal(t *testing.T) {
  example := mustReadFile(t, "github/testdata/push_response.json")

	want := PushPayload{
		Ref:    "refs/heads/master",
		Before: "0802d5e6cee084a8f867c5406e46a3fca556bf4e",
		After:  "89b269b3c313d05c182e5ff829727f2b5132c2e5",
		Repository: &Repository{
			FullName: "minorhacks/advent_2020",
		},
	}

	got := PushPayload{}

	if err := json.Unmarshal([]byte(example), &got); err != nil {
		t.Fatalf("json.Unmarshal got error %v; want no error", err)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Error(diff)
	}
}
