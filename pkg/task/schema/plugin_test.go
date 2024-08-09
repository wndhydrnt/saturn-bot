package schema_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func TestPlugin_PathAbs(t *testing.T) {
	relative := &schema.Plugin{Path: "../unittest"}
	result := relative.PathAbs("/tmp/tasks/test/task.yaml")
	require.Equal(t, "/tmp/tasks/unittest", result)

	absolute := &schema.Plugin{Path: "/tmp/plugins/unittest"}
	result = absolute.PathAbs("/tmp/tasks/test/task.yaml")
	require.Equal(t, "/tmp/plugins/unittest", result)
}
