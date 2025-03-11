package snapshot

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/transferria/transferria/pkg/abstract"
	"github.com/transferria/transferria/pkg/providers/clickhouse/conn"
	"github.com/transferria/transferria/pkg/providers/clickhouse/model"
	chrecipe "github.com/transferria/transferria/pkg/providers/clickhouse/recipe"
	"github.com/transferria/transferria/tests/helpers"
	proxy "github.com/transferria/transferria/tests/helpers/http_proxy"
)

var (
	databaseName = "some_db"
	TransferType = abstract.TransferTypeSnapshotOnly
	Source       = *chrecipe.MustSource(chrecipe.WithInitFile("dump/src.sql"), chrecipe.WithDatabase(databaseName))
	Target       = *chrecipe.MustTarget(chrecipe.WithInitFile("dump/dst.sql"), chrecipe.WithDatabase(databaseName), chrecipe.WithPrefix("DB0_"))
)

func init() {
	_ = os.Setenv("YC", "1")                                               // to not go to vanga
	helpers.InitSrcDst(helpers.TransferID, &Source, &Target, TransferType) // to WithDefaults() & FillDependentFields(): IsHomo, helpers.TransferID, IsUpdateable
}

func TestSnapshot(t *testing.T) {
	defer func() {
		require.NoError(t, helpers.CheckConnections(
			helpers.LabeledPort{Label: "CH source", Port: Source.NativePort},
			helpers.LabeledPort{Label: "CH target", Port: Target.NativePort},
		))
	}()

	srcProxy := proxy.NewHTTPProxyWithPortAllocation(fmt.Sprintf("localhost:%d", Source.HTTPPort))
	srcProxy.WithLogger = true
	srcWorker := srcProxy.RunAsync()
	defer srcWorker.Close()
	fmt.Printf("Source.HTTPPort:%d, srcProxy.ListenPort:%d\n", Source.HTTPPort, srcProxy.ListenPort)
	Source.HTTPPort = srcProxy.ListenPort
	Source.BufferSize = 500

	dstProxy := proxy.NewHTTPProxyWithPortAllocation(fmt.Sprintf("localhost:%d", Target.HTTPPort))
	dstProxy.WithLogger = true
	dstWorker := dstProxy.RunAsync()
	defer dstWorker.Close()
	Target.HTTPPort = dstProxy.ListenPort

	t.Run("default, CSV case", func(t *testing.T) {
		transfer := helpers.MakeTransfer("fake", &Source, &Target, abstract.TransferTypeSnapshotOnly)
		helpers.Activate(t, transfer)
		require.NoError(t, helpers.CompareStorages(t, Source, Target, helpers.NewCompareStorageParams()))
		require.True(t, proxy.CheckRequestContains(srcProxy.GetSniffedData(), "FORMAT CSV"))
		require.True(t, proxy.CheckRequestContains(srcProxy.GetSniffedData(), "timeout_before_checking_execution_speed=0"))
		require.True(t, proxy.CheckRequestContains(dstProxy.GetSniffedData(), "FORMAT CSV"))
	})

	t.Run("drop", func(t *testing.T) {
		transfer := helpers.MakeTransfer("fake", &Source, &Target, abstract.TransferTypeSnapshotOnly)
		db, err := conn.ConnectNative("localhost", Target.ToSinkParams(transfer))
		require.NoError(t, err)

		exec := func(query string) {
			_, err := db.Exec(query)
			require.NoError(t, err)
		}

		exec("drop table some_db.some_table")

		srcProxy.ResetSniffedData()
		dstProxy.ResetSniffedData()
	})

	t.Run("JSONCompactEachRow case", func(t *testing.T) {
		Source.IOHomoFormat = model.ClickhouseIOFormatJSONCompact
		transfer := helpers.MakeTransfer("fake", &Source, &Target, abstract.TransferTypeSnapshotOnly)
		helpers.Activate(t, transfer)
		require.NoError(t, helpers.CompareStorages(t, Source, Target, helpers.NewCompareStorageParams()))
		require.True(t, proxy.CheckRequestContains(srcProxy.GetSniffedData(), "FORMAT JSONCompactEachRow"))
		require.True(t, proxy.CheckRequestContains(srcProxy.GetSniffedData(), "timeout_before_checking_execution_speed=0"))
		require.True(t, proxy.CheckRequestContains(dstProxy.GetSniffedData(), "FORMAT JSONCompactEachRow"))
	})
}
