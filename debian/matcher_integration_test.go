package debian

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quay/zlog"

	"github.com/quay/claircore"
	"github.com/quay/claircore/internal/matcher"
	"github.com/quay/claircore/internal/updater"
	vulnstore "github.com/quay/claircore/internal/vulnstore/postgres"
	"github.com/quay/claircore/libvuln/driver"
	"github.com/quay/claircore/test/integration"
)

// TestMatcherIntegration confirms packages are matched
// with vulnerabilities correctly. the returned
// store from postgres.NewTestStore must have Ubuntu
// CVE data
func TestMatcherIntegration(t *testing.T) {
	integration.Skip(t)
	ctx := zlog.Test(context.Background(), t)
	pool := vulnstore.TestDB(ctx, t)
	store := vulnstore.NewVulnStore(pool)

	m := &Matcher{}
	// seed the test vulnstore with CVE data
	ch := make(chan driver.Updater)
	go func() {
		ch <- NewUpdater(Buster)
		close(ch)
	}()
	exec := updater.Online{Pool: pool}
	// force update
	tctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := exec.Run(tctx, ch); err != nil {
		t.Error(err)
	}
	path := filepath.Join("testdata", "indexreport-buster-jackson-databind.json")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var ir claircore.IndexReport
	err = json.NewDecoder(f).Decode(&ir)
	if err != nil {
		t.Fatalf("failed to decode IndexReport: %v", err)
	}
	vr, err := matcher.Match(ctx, &ir, []driver.Matcher{m}, store)
	if err != nil {
		t.Fatalf("expected nil error but got %v", err)
	}
	_, err = json.Marshal(&vr)
	if err != nil {
		t.Fatalf("failed to marshal VR: %v", err)
	}
}
