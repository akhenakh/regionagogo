package mobile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	db := NewGeoDB()
	defer db.Close()
	err := db.OpenDB("../region.db")
	require.NoError(t, err)

	f := db.QueryHandler(47.01492366313195, -70.842592064976714)
	require.NotNil(t, f)
}
