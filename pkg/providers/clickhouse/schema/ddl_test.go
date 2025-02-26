package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsDistributedDDL(t *testing.T) {
	ddl := "CREATE TABLE logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	require.False(t, IsDistributedDDL(ddl))

	ddl = "CREATE TABLE logs.test7 ON CLUSTER (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	require.True(t, IsDistributedDDL(ddl))

	ddl = "CREATE TABLE logs.test7 on  cluster (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	require.True(t, IsDistributedDDL(ddl))
}

func TestReplaceCluster(t *testing.T) {
	ddl := "CREATE TABLE logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	changedDDL := ReplaceCluster(ddl, "{cluster}")
	require.Equal(t, changedDDL, ddl)

	ddl = "CREATE TABLE logs.test7 ON CLUSTER `abcdef` (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	changedDDL = ReplaceCluster(ddl, "{cluster}")
	require.Equal(t,
		"CREATE TABLE logs.test7  ON CLUSTER `{cluster}` (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192",
		changedDDL,
	)

	ddl = "CREATE TABLE logs.test7 on  cluster abcdef (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	changedDDL = ReplaceCluster(ddl, "{cluster}")
	require.Equal(t,
		"CREATE TABLE logs.test7  ON CLUSTER `{cluster}` (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192",
		changedDDL,
	)
}

func TestSetReplicatedEngine(t *testing.T) {
	ddl := "CREATE TABLE logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	_, err := SetReplicatedEngine(ddl, "MergeTree", "logs", "test7")
	require.Error(t, err)

	changedDDL, err := SetReplicatedEngine(ddl, "ReplicatedMergeTree", "logs", "test7")
	require.NoError(t, err)
	require.Equal(t, ddl, changedDDL)

	ddl = "CREATE TABLE default.attributes (`id` UInt32, `event_date` Date, `orders_count` UInt32, `rating` UInt8) ENGINE = MergeTree ORDER BY event_date SETTINGS index_granularity = 8192"
	changedDDL, err = SetReplicatedEngine(ddl, "MergeTree", "default", "attributes")
	require.NoError(t, err)
	require.Equal(t,
		"CREATE TABLE default.attributes (`id` UInt32, `event_date` Date, `orders_count` UInt32, `rating` UInt8) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/default.attributes_cdc', '{replica}') ORDER BY event_date SETTINGS index_granularity = 8192",
		changedDDL,
	)

	ddl = "CREATE TABLE default.attributes (`id` UInt32, `event_date` Date, `orders_count` UInt32, `rating` UInt8) ENGINE = MergeTree() ORDER BY event_date SETTINGS index_granularity = 8192"
	changedDDL, err = SetReplicatedEngine(ddl, "MergeTree", "default", "attributes")
	require.NoError(t, err)
	require.Equal(t,
		"CREATE TABLE default.attributes (`id` UInt32, `event_date` Date, `orders_count` UInt32, `rating` UInt8) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/default.attributes_cdc', '{replica}') ORDER BY event_date SETTINGS index_granularity = 8192",
		changedDDL,
	)

	ddl = "CREATE TABLE default.search_conversions ( `search_uuid` String, `search_id` Int32, `uuid` String, `client_id` UInt32, `search_date` Date, `updated_at` UInt32, `search_at` UInt32, `pageview_at` Nullable(UInt32), INDEX hp pageview_at TYPE minmax GRANULARITY 1) ENGINE = ReplacingMergeTree PARTITION BY search_date ORDER BY (uuid, search_id) SETTINGS index_granularity = 8192"
	changedDDL, err = SetReplicatedEngine(ddl, "ReplacingMergeTree", "default", "search_conversions")
	require.NoError(t, err)
	require.Equal(t,
		"CREATE TABLE default.search_conversions ( `search_uuid` String, `search_id` Int32, `uuid` String, `client_id` UInt32, `search_date` Date, `updated_at` UInt32, `search_at` UInt32, `pageview_at` Nullable(UInt32), INDEX hp pageview_at TYPE minmax GRANULARITY 1) ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/default.search_conversions_cdc', '{replica}') PARTITION BY search_date ORDER BY (uuid, search_id) SETTINGS index_granularity = 8192",
		changedDDL,
	)
}

func TestSetIfNotExists(t *testing.T) {
	ddl := "CREATE TABLE logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	changedDDL := SetIfNotExists(ddl)
	require.Equal(t,
		"CREATE TABLE IF NOT EXISTS logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192",
		changedDDL,
	)

	ddl = "CREATE TABLE IF NOT EXISTS logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	changedDDL = SetIfNotExists(ddl)
	require.Equal(t, changedDDL, ddl)

	ddl = "CREATE MATERIALIZED VIEW logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	changedDDL = SetIfNotExists(ddl)
	require.Equal(t,
		"CREATE MATERIALIZED VIEW IF NOT EXISTS logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192",
		changedDDL,
	)
}

func TestMakeDistributedDDL(t *testing.T) {
	ddl := "CREATE TABLE logs.test7 (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	changedDDL := MakeDistributedDDL(ddl, "{cluster}")
	require.Equal(t,
		"CREATE TABLE logs.test7  ON CLUSTER `{cluster}` (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192",
		changedDDL,
	)

	ddl = "CREATE TABLE logs.test7 ON CLUSTER abcd (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192"
	changedDDL = MakeDistributedDDL(ddl, "{cluster}")
	require.Equal(t,
		"CREATE TABLE logs.test7  ON CLUSTER `{cluster}` (`id` String, `counter` Int32, `vals` Array(String) CODEC(LZ4HC(9))) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/logs.test7_cdc', '{replica}') ORDER BY id SETTINGS index_granularity = 8192",
		changedDDL,
	)
}

func TestComplexDDL(t *testing.T) {
	ddl := "CREATE TABLE IF NOT EXISTS research.all_serp_competitor UUID 'fe7491ee-102b-40b1-be74-91ee102b90b1' ON CLUSTER chclpinrb4126vagc689\n(\n    `timestamp` DateTime DEFAULT now(),\n    `keyword` String,\n    `country` String,\n    `lang` String,\n    `organicRank` UInt8,\n    `search_volume` UInt64,\n    `url` String,\n    `subDomain` String,\n    `domain` String,\n    `path` String,\n    `etv` UInt64,\n    `clickstream_etv` Float64,\n    `related` Array(String),\n    `estimated_traffic` UInt64 DEFAULT CAST(multiIf(organicRank = 1, search_volume * 0.3197, organicRank = 2, search_volume * 0.1551, organicRank = 3, search_volume * 0.0932, organicRank = 4, search_volume * 0.0593, organicRank = 5, search_volume * 0.041, organicRank = 6, search_volume * 0.029, organicRank = 7, search_volume * 0.0213, organicRank = 8, search_volume * 0.0163, organicRank = 9, search_volume * 0.0131, organicRank = 10, search_volume * 0.0108, organicRank = 11, search_volume * 0.01, organicRank = 12, search_volume * 0.0112, organicRank = 13, search_volume * 0.0124, organicRank = 14, search_volume * 0.0114, organicRank = 15, search_volume * 0.0103, organicRank = 16, search_volume * 0.0099, organicRank = 17, search_volume * 0.0094, organicRank = 18, search_volume * 0.0081, organicRank = 19, search_volume * 0.0076, organicRank = 20, search_volume * 0.0067, 0.), 'Int64'),\n    INDEX timestamp_minmax timestamp TYPE minmax GRANULARITY 8192,\n    INDEX domain_bf domain TYPE bloom_filter GRANULARITY 65536,\n    PROJECTION order_by_keyword__domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY keyword\n    ),\n    PROJECTION order_by_organic_rank__keyword_domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY organicRank\n    )\n)\nENGINE = MergeTree\nORDER BY (country, domain)\nSETTINGS index_granularity = 8192"
	require.Equal(t,
		"CREATE TABLE IF NOT EXISTS research.all_serp_competitor UUID 'fe7491ee-102b-40b1-be74-91ee102b90b1'  ON CLUSTER `{cluster}`\n(\n    `timestamp` DateTime DEFAULT now(),\n    `keyword` String,\n    `country` String,\n    `lang` String,\n    `organicRank` UInt8,\n    `search_volume` UInt64,\n    `url` String,\n    `subDomain` String,\n    `domain` String,\n    `path` String,\n    `etv` UInt64,\n    `clickstream_etv` Float64,\n    `related` Array(String),\n    `estimated_traffic` UInt64 DEFAULT CAST(multiIf(organicRank = 1, search_volume * 0.3197, organicRank = 2, search_volume * 0.1551, organicRank = 3, search_volume * 0.0932, organicRank = 4, search_volume * 0.0593, organicRank = 5, search_volume * 0.041, organicRank = 6, search_volume * 0.029, organicRank = 7, search_volume * 0.0213, organicRank = 8, search_volume * 0.0163, organicRank = 9, search_volume * 0.0131, organicRank = 10, search_volume * 0.0108, organicRank = 11, search_volume * 0.01, organicRank = 12, search_volume * 0.0112, organicRank = 13, search_volume * 0.0124, organicRank = 14, search_volume * 0.0114, organicRank = 15, search_volume * 0.0103, organicRank = 16, search_volume * 0.0099, organicRank = 17, search_volume * 0.0094, organicRank = 18, search_volume * 0.0081, organicRank = 19, search_volume * 0.0076, organicRank = 20, search_volume * 0.0067, 0.), 'Int64'),\n    INDEX timestamp_minmax timestamp TYPE minmax GRANULARITY 8192,\n    INDEX domain_bf domain TYPE bloom_filter GRANULARITY 65536,\n    PROJECTION order_by_keyword__domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY keyword\n    ),\n    PROJECTION order_by_organic_rank__keyword_domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY organicRank\n    )\n)\nENGINE = MergeTree\nORDER BY (country, domain)\nSETTINGS index_granularity = 8192",
		MakeDistributedDDL(ddl, "{cluster}"))
}

func TestSpacedDDL(t *testing.T) {
	ddl := "CREATE TABLE IF NOT EXISTS research.all_serp_competitor UUID 'fe7491ee-102b-40b1-be74-91ee102b90b1' ON CLUSTER `chclpinrb4126va gc689`\n(\n    `timestamp` DateTime DEFAULT now(),\n    `keyword` String,\n    `country` String,\n    `lang` String,\n    `organicRank` UInt8,\n    `search_volume` UInt64,\n    `url` String,\n    `subDomain` String,\n    `domain` String,\n    `path` String,\n    `etv` UInt64,\n    `clickstream_etv` Float64,\n    `related` Array(String),\n    `estimated_traffic` UInt64 DEFAULT CAST(multiIf(organicRank = 1, search_volume * 0.3197, organicRank = 2, search_volume * 0.1551, organicRank = 3, search_volume * 0.0932, organicRank = 4, search_volume * 0.0593, organicRank = 5, search_volume * 0.041, organicRank = 6, search_volume * 0.029, organicRank = 7, search_volume * 0.0213, organicRank = 8, search_volume * 0.0163, organicRank = 9, search_volume * 0.0131, organicRank = 10, search_volume * 0.0108, organicRank = 11, search_volume * 0.01, organicRank = 12, search_volume * 0.0112, organicRank = 13, search_volume * 0.0124, organicRank = 14, search_volume * 0.0114, organicRank = 15, search_volume * 0.0103, organicRank = 16, search_volume * 0.0099, organicRank = 17, search_volume * 0.0094, organicRank = 18, search_volume * 0.0081, organicRank = 19, search_volume * 0.0076, organicRank = 20, search_volume * 0.0067, 0.), 'Int64'),\n    INDEX timestamp_minmax timestamp TYPE minmax GRANULARITY 8192,\n    INDEX domain_bf domain TYPE bloom_filter GRANULARITY 65536,\n    PROJECTION order_by_keyword__domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY keyword\n    ),\n    PROJECTION order_by_organic_rank__keyword_domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY organicRank\n    )\n)\nENGINE = MergeTree\nORDER BY (country, domain)\nSETTINGS index_granularity = 8192"
	require.Equal(t,
		"CREATE TABLE IF NOT EXISTS research.all_serp_competitor UUID 'fe7491ee-102b-40b1-be74-91ee102b90b1'  ON CLUSTER `{cluster}`\n(\n    `timestamp` DateTime DEFAULT now(),\n    `keyword` String,\n    `country` String,\n    `lang` String,\n    `organicRank` UInt8,\n    `search_volume` UInt64,\n    `url` String,\n    `subDomain` String,\n    `domain` String,\n    `path` String,\n    `etv` UInt64,\n    `clickstream_etv` Float64,\n    `related` Array(String),\n    `estimated_traffic` UInt64 DEFAULT CAST(multiIf(organicRank = 1, search_volume * 0.3197, organicRank = 2, search_volume * 0.1551, organicRank = 3, search_volume * 0.0932, organicRank = 4, search_volume * 0.0593, organicRank = 5, search_volume * 0.041, organicRank = 6, search_volume * 0.029, organicRank = 7, search_volume * 0.0213, organicRank = 8, search_volume * 0.0163, organicRank = 9, search_volume * 0.0131, organicRank = 10, search_volume * 0.0108, organicRank = 11, search_volume * 0.01, organicRank = 12, search_volume * 0.0112, organicRank = 13, search_volume * 0.0124, organicRank = 14, search_volume * 0.0114, organicRank = 15, search_volume * 0.0103, organicRank = 16, search_volume * 0.0099, organicRank = 17, search_volume * 0.0094, organicRank = 18, search_volume * 0.0081, organicRank = 19, search_volume * 0.0076, organicRank = 20, search_volume * 0.0067, 0.), 'Int64'),\n    INDEX timestamp_minmax timestamp TYPE minmax GRANULARITY 8192,\n    INDEX domain_bf domain TYPE bloom_filter GRANULARITY 65536,\n    PROJECTION order_by_keyword__domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY keyword\n    ),\n    PROJECTION order_by_organic_rank__keyword_domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY organicRank\n    )\n)\nENGINE = MergeTree\nORDER BY (country, domain)\nSETTINGS index_granularity = 8192",
		MakeDistributedDDL(ddl, "{cluster}"))

	ddl = "CREATE TABLE IF NOT EXISTS research.all_serp_competitor UUID 'fe7491ee-102b-40b1- be74-91ee102b90b1' ON CLUSTER `chclpinrb4126va gc689`\n(\n    `timestamp` DateTime DEFAULT now(),\n    `keyword` String,\n    `country` String,\n    `lang` String,\n    `organicRank` UInt8,\n    `search_volume` UInt64,\n    `url` String,\n    `subDomain` String,\n    `domain` String,\n    `path` String,\n    `etv` UInt64,\n    `clickstream_etv` Float64,\n    `related` Array(String),\n    `estimated_traffic` UInt64 DEFAULT CAST(multiIf(organicRank = 1, search_volume * 0.3197, organicRank = 2, search_volume * 0.1551, organicRank = 3, search_volume * 0.0932, organicRank = 4, search_volume * 0.0593, organicRank = 5, search_volume * 0.041, organicRank = 6, search_volume * 0.029, organicRank = 7, search_volume * 0.0213, organicRank = 8, search_volume * 0.0163, organicRank = 9, search_volume * 0.0131, organicRank = 10, search_volume * 0.0108, organicRank = 11, search_volume * 0.01, organicRank = 12, search_volume * 0.0112, organicRank = 13, search_volume * 0.0124, organicRank = 14, search_volume * 0.0114, organicRank = 15, search_volume * 0.0103, organicRank = 16, search_volume * 0.0099, organicRank = 17, search_volume * 0.0094, organicRank = 18, search_volume * 0.0081, organicRank = 19, search_volume * 0.0076, organicRank = 20, search_volume * 0.0067, 0.), 'Int64'),\n    INDEX timestamp_minmax timestamp TYPE minmax GRANULARITY 8192,\n    INDEX domain_bf domain TYPE bloom_filter GRANULARITY 65536,\n    PROJECTION order_by_keyword__domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY keyword\n    ),\n    PROJECTION order_by_organic_rank__keyword_domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY organicRank\n    )\n)\nENGINE = MergeTree\nORDER BY (country, domain)\nSETTINGS index_granularity = 8192"
	require.Equal(t,
		"CREATE TABLE IF NOT EXISTS research.all_serp_competitor UUID 'fe7491ee-102b-40b1- be74-91ee102b90b1'  ON CLUSTER `{cluster}`\n(\n    `timestamp` DateTime DEFAULT now(),\n    `keyword` String,\n    `country` String,\n    `lang` String,\n    `organicRank` UInt8,\n    `search_volume` UInt64,\n    `url` String,\n    `subDomain` String,\n    `domain` String,\n    `path` String,\n    `etv` UInt64,\n    `clickstream_etv` Float64,\n    `related` Array(String),\n    `estimated_traffic` UInt64 DEFAULT CAST(multiIf(organicRank = 1, search_volume * 0.3197, organicRank = 2, search_volume * 0.1551, organicRank = 3, search_volume * 0.0932, organicRank = 4, search_volume * 0.0593, organicRank = 5, search_volume * 0.041, organicRank = 6, search_volume * 0.029, organicRank = 7, search_volume * 0.0213, organicRank = 8, search_volume * 0.0163, organicRank = 9, search_volume * 0.0131, organicRank = 10, search_volume * 0.0108, organicRank = 11, search_volume * 0.01, organicRank = 12, search_volume * 0.0112, organicRank = 13, search_volume * 0.0124, organicRank = 14, search_volume * 0.0114, organicRank = 15, search_volume * 0.0103, organicRank = 16, search_volume * 0.0099, organicRank = 17, search_volume * 0.0094, organicRank = 18, search_volume * 0.0081, organicRank = 19, search_volume * 0.0076, organicRank = 20, search_volume * 0.0067, 0.), 'Int64'),\n    INDEX timestamp_minmax timestamp TYPE minmax GRANULARITY 8192,\n    INDEX domain_bf domain TYPE bloom_filter GRANULARITY 65536,\n    PROJECTION order_by_keyword__domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY keyword\n    ),\n    PROJECTION order_by_organic_rank__keyword_domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY organicRank\n    )\n)\nENGINE = MergeTree\nORDER BY (country, domain)\nSETTINGS index_granularity = 8192",
		MakeDistributedDDL(ddl, "{cluster}"))

	ddl = "CREATE TABLE IF NOT EXISTS research.all_serp_competitor ON CLUSTER `chclpinrb4126va gc689`\n(\n    `timestamp` DateTime DEFAULT now(),\n    `keyword` String,\n    `country` String,\n    `lang` String,\n    `organicRank` UInt8,\n    `search_volume` UInt64,\n    `url` String,\n    `subDomain` String,\n    `domain` String,\n    `path` String,\n    `etv` UInt64,\n    `clickstream_etv` Float64,\n    `related` Array(String),\n    `estimated_traffic` UInt64 DEFAULT CAST(multiIf(organicRank = 1, search_volume * 0.3197, organicRank = 2, search_volume * 0.1551, organicRank = 3, search_volume * 0.0932, organicRank = 4, search_volume * 0.0593, organicRank = 5, search_volume * 0.041, organicRank = 6, search_volume * 0.029, organicRank = 7, search_volume * 0.0213, organicRank = 8, search_volume * 0.0163, organicRank = 9, search_volume * 0.0131, organicRank = 10, search_volume * 0.0108, organicRank = 11, search_volume * 0.01, organicRank = 12, search_volume * 0.0112, organicRank = 13, search_volume * 0.0124, organicRank = 14, search_volume * 0.0114, organicRank = 15, search_volume * 0.0103, organicRank = 16, search_volume * 0.0099, organicRank = 17, search_volume * 0.0094, organicRank = 18, search_volume * 0.0081, organicRank = 19, search_volume * 0.0076, organicRank = 20, search_volume * 0.0067, 0.), 'Int64'),\n    INDEX timestamp_minmax timestamp TYPE minmax GRANULARITY 8192,\n    INDEX domain_bf domain TYPE bloom_filter GRANULARITY 65536,\n    PROJECTION order_by_keyword__domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY keyword\n    ),\n    PROJECTION order_by_organic_rank__keyword_domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY organicRank\n    )\n)\nENGINE = MergeTree\nORDER BY (country, domain)\nSETTINGS index_granularity = 8192"
	require.Equal(t,
		"CREATE TABLE IF NOT EXISTS research.all_serp_competitor  ON CLUSTER `{cluster}`\n(\n    `timestamp` DateTime DEFAULT now(),\n    `keyword` String,\n    `country` String,\n    `lang` String,\n    `organicRank` UInt8,\n    `search_volume` UInt64,\n    `url` String,\n    `subDomain` String,\n    `domain` String,\n    `path` String,\n    `etv` UInt64,\n    `clickstream_etv` Float64,\n    `related` Array(String),\n    `estimated_traffic` UInt64 DEFAULT CAST(multiIf(organicRank = 1, search_volume * 0.3197, organicRank = 2, search_volume * 0.1551, organicRank = 3, search_volume * 0.0932, organicRank = 4, search_volume * 0.0593, organicRank = 5, search_volume * 0.041, organicRank = 6, search_volume * 0.029, organicRank = 7, search_volume * 0.0213, organicRank = 8, search_volume * 0.0163, organicRank = 9, search_volume * 0.0131, organicRank = 10, search_volume * 0.0108, organicRank = 11, search_volume * 0.01, organicRank = 12, search_volume * 0.0112, organicRank = 13, search_volume * 0.0124, organicRank = 14, search_volume * 0.0114, organicRank = 15, search_volume * 0.0103, organicRank = 16, search_volume * 0.0099, organicRank = 17, search_volume * 0.0094, organicRank = 18, search_volume * 0.0081, organicRank = 19, search_volume * 0.0076, organicRank = 20, search_volume * 0.0067, 0.), 'Int64'),\n    INDEX timestamp_minmax timestamp TYPE minmax GRANULARITY 8192,\n    INDEX domain_bf domain TYPE bloom_filter GRANULARITY 65536,\n    PROJECTION order_by_keyword__domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY keyword\n    ),\n    PROJECTION order_by_organic_rank__keyword_domain_country_lang_organic_rank\n    (\n        SELECT\n            keyword,\n            domain,\n            country,\n            lang,\n            organicRank\n        ORDER BY organicRank\n    )\n)\nENGINE = MergeTree\nORDER BY (country, domain)\nSETTINGS index_granularity = 8192",
		MakeDistributedDDL(ddl, "{cluster}"))
}

func TestBigNestedParentheses(t *testing.T) {
	ddl := "CREATE TABLE sentry.errors_local\n(\n    `project_id` UInt64,\n    `timestamp` DateTime,\n    `event_id` UUID CODEC(NONE),\n    `platform` LowCardinality(String),\n    `environment` LowCardinality(Nullable(String)),\n    `release` LowCardinality(Nullable(String)),\n    `dist` LowCardinality(Nullable(String)),\n    `ip_address_v4` Nullable(IPv4),\n    `ip_address_v6` Nullable(IPv6),\n    `user` String DEFAULT '',\n    `user_hash` UInt64 MATERIALIZED cityHash64(user),\n    `user_id` Nullable(String),\n    `user_name` Nullable(String),\n    `user_email` Nullable(String),\n    `sdk_name` LowCardinality(Nullable(String)),\n    `sdk_version` LowCardinality(Nullable(String)),\n    `http_method` LowCardinality(Nullable(String)),\n    `http_referer` Nullable(String),\n    `tags.key` Array(String),\n    `tags.value` Array(String),\n    `_tags_hash_map` Array(UInt64) MATERIALIZED arrayMap((k,v) - cityHash64(concat(replaceRegexpAll(k,'(=)','1'),'=',v)),tags.key,tags.value),\n    `contexts.key` Array(String),\n    `contexts.value` Array(String),\n    `transaction_name` LowCardinality(String) DEFAULT '',\n    `transaction_hash` UInt64 MATERIALIZED cityHash64(transaction_name),\n    `span_id` Nullable(UInt64),\n    `trace_id` Nullable(UUID),\n    `partition` UInt16,\n    `offset` UInt64 CODEC(DoubleDelta, LZ4),\n    `message_timestamp` DateTime,\n    `retention_days` UInt16,\n    `deleted` UInt8,\n    `group_id` UInt64,\n    `primary_hash` UUID,\n    `hierarchical_hashes` Array(UUID),\n    `received` DateTime,\n    `message` String,\n    `title` String,\n    `culprit` String,\n    `level` LowCardinality(Nullable(String)),\n    `location` Nullable(String),\n    `version` LowCardinality(Nullable(String)),\n    `type` LowCardinality(String),\n    `exception_stacks.type` Array(Nullable(String)),\n    `exception_stacks.value` Array(Nullable(String)),\n    `exception_stacks.mechanism_type` Array(Nullable(String)),\n    `exception_stacks.mechanism_handled` Array(Nullable(UInt8)),\n    `exception_frames.abs_path` Array(Nullable(String)),\n    `exception_frames.colno` Array(Nullable(UInt32)),\n    `exception_frames.filename` Array(Nullable(String)),\n    `exception_frames.function` Array(Nullable(String)),\n    `exception_frames.lineno` Array(Nullable(UInt32)),\n    `exception_frames.in_app` Array(Nullable(UInt8)),\n    `exception_frames.package` Array(Nullable(String)),\n    `exception_frames.module` Array(Nullable(String)),\n    `exception_frames.stack_level` Array(Nullable(UInt16)),\n    `sdk_integrations` Array(String),\n    `modules.name` Array(String),\n    `modules.version` Array(String),\n    `exception_main_thread` Nullable(UInt8),\n    `trace_sampled` Nullable(UInt8),\n    `num_processing_errors` Nullable(UInt64),\n    `replay_id` Nullable(UUID),\n    INDEX bf_tags_hash_map _tags_hash_map TYPE bloom_filter GRANULARITY 1,\n    INDEX minmax_group_id group_id TYPE minmax GRANULARITY 1,\n    INDEX bf_release release TYPE bloom_filter GRANULARITY 1\n)\nENGINE = ReplacingMergeTree(deleted)\nPARTITION BY (retention_days, toMonday(timestamp))\nORDER BY (project_id, toStartOfDay(timestamp), primary_hash, cityHash64(event_id))\nSAMPLE BY cityHash64(event_id)\nTTL timestamp + toIntervalDay(retention_days)\nSETTINGS index_granularity = 8192,\n min_bytes_for_wide_part = 1,\n enable_vertical_merge_algorithm = 1,\n min_rows_for_wide_part = 0,\n ttl_only_drop_parts = 1;"
	require.Equal(t,
		"CREATE TABLE sentry.errors_local\n ON CLUSTER `{cluster}` (\n    `project_id` UInt64,\n    `timestamp` DateTime,\n    `event_id` UUID CODEC(NONE),\n    `platform` LowCardinality(String),\n    `environment` LowCardinality(Nullable(String)),\n    `release` LowCardinality(Nullable(String)),\n    `dist` LowCardinality(Nullable(String)),\n    `ip_address_v4` Nullable(IPv4),\n    `ip_address_v6` Nullable(IPv6),\n    `user` String DEFAULT '',\n    `user_hash` UInt64 MATERIALIZED cityHash64(user),\n    `user_id` Nullable(String),\n    `user_name` Nullable(String),\n    `user_email` Nullable(String),\n    `sdk_name` LowCardinality(Nullable(String)),\n    `sdk_version` LowCardinality(Nullable(String)),\n    `http_method` LowCardinality(Nullable(String)),\n    `http_referer` Nullable(String),\n    `tags.key` Array(String),\n    `tags.value` Array(String),\n    `_tags_hash_map` Array(UInt64) MATERIALIZED arrayMap((k,v) - cityHash64(concat(replaceRegexpAll(k,'(=)','1'),'=',v)),tags.key,tags.value),\n    `contexts.key` Array(String),\n    `contexts.value` Array(String),\n    `transaction_name` LowCardinality(String) DEFAULT '',\n    `transaction_hash` UInt64 MATERIALIZED cityHash64(transaction_name),\n    `span_id` Nullable(UInt64),\n    `trace_id` Nullable(UUID),\n    `partition` UInt16,\n    `offset` UInt64 CODEC(DoubleDelta, LZ4),\n    `message_timestamp` DateTime,\n    `retention_days` UInt16,\n    `deleted` UInt8,\n    `group_id` UInt64,\n    `primary_hash` UUID,\n    `hierarchical_hashes` Array(UUID),\n    `received` DateTime,\n    `message` String,\n    `title` String,\n    `culprit` String,\n    `level` LowCardinality(Nullable(String)),\n    `location` Nullable(String),\n    `version` LowCardinality(Nullable(String)),\n    `type` LowCardinality(String),\n    `exception_stacks.type` Array(Nullable(String)),\n    `exception_stacks.value` Array(Nullable(String)),\n    `exception_stacks.mechanism_type` Array(Nullable(String)),\n    `exception_stacks.mechanism_handled` Array(Nullable(UInt8)),\n    `exception_frames.abs_path` Array(Nullable(String)),\n    `exception_frames.colno` Array(Nullable(UInt32)),\n    `exception_frames.filename` Array(Nullable(String)),\n    `exception_frames.function` Array(Nullable(String)),\n    `exception_frames.lineno` Array(Nullable(UInt32)),\n    `exception_frames.in_app` Array(Nullable(UInt8)),\n    `exception_frames.package` Array(Nullable(String)),\n    `exception_frames.module` Array(Nullable(String)),\n    `exception_frames.stack_level` Array(Nullable(UInt16)),\n    `sdk_integrations` Array(String),\n    `modules.name` Array(String),\n    `modules.version` Array(String),\n    `exception_main_thread` Nullable(UInt8),\n    `trace_sampled` Nullable(UInt8),\n    `num_processing_errors` Nullable(UInt64),\n    `replay_id` Nullable(UUID),\n    INDEX bf_tags_hash_map _tags_hash_map TYPE bloom_filter GRANULARITY 1,\n    INDEX minmax_group_id group_id TYPE minmax GRANULARITY 1,\n    INDEX bf_release release TYPE bloom_filter GRANULARITY 1\n)\nENGINE = ReplacingMergeTree(deleted)\nPARTITION BY (retention_days, toMonday(timestamp))\nORDER BY (project_id, toStartOfDay(timestamp), primary_hash, cityHash64(event_id))\nSAMPLE BY cityHash64(event_id)\nTTL timestamp + toIntervalDay(retention_days)\nSETTINGS index_granularity = 8192,\n min_bytes_for_wide_part = 1,\n enable_vertical_merge_algorithm = 1,\n min_rows_for_wide_part = 0,\n ttl_only_drop_parts = 1;",
		MakeDistributedDDL(ddl, "{cluster}"))
}
