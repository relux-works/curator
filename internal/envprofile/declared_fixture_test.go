package envprofile

import (
	"github.com/relux-works/curator/internal/testgitdispatch"
	"testing"
)

type declaredFixture = testgitdispatch.Config

func newDeclaredFixture(t *testing.T) string { return testgitdispatch.New(t) }
func (g *gitIdentities) configureDispatch(change func(*declaredFixture)) {
	testgitdispatch.Configure(g.t, g.dispatch, change)
}
