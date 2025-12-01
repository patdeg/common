# Touchpoints Payload: Dual-Column Pattern

This document explains the dual-column pattern used for storing payload data in the BigQuery `touchpoints` table.

## Overview

The touchpoints table uses two columns for payload data:

| Column | Type | Purpose |
|--------|------|---------|
| `PayloadString` | STRING | Reliable ingestion via streaming insert |
| `Payload` | JSON | Efficient queries with dot-notation |

## Why Two Columns?

BigQuery's streaming insert API (v2) has compatibility issues with native JSON column types. When passing Go maps or JSON strings to a JSON column, the API returns errors like:

```
"This field: payload is not a record"
```

Using a STRING column for ingestion is 100% reliable, while the JSON column provides better query ergonomics.

## How It Works

### 1. Ingestion (Automatic)

The `StoreTouchPointInBigQuery()` function in `track/bigquery_store.go` automatically:
- Validates the payload JSON
- Stores it in the `PayloadString` column as a raw string
- Never touches the `Payload` column during insert

### 2. Conversion (Manual)

Run this SQL query to convert `PayloadString` to `Payload`:

```sql
UPDATE `demeterics.touchpoints.touchpoints`
SET Payload = SAFE.PARSE_JSON(PayloadString)
WHERE Payload IS NULL
  AND PayloadString IS NOT NULL
  AND PayloadString != '{}'
  AND _PARTITIONTIME >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
```

Adjust the time interval based on how much historical data to convert.

### 3. Querying

**Before conversion** (using PayloadString):
```sql
SELECT
  Time,
  Category,
  Action,
  JSON_VALUE(PayloadString, '$.utm_source') as utm_source,
  JSON_VALUE(PayloadString, '$.utm_campaign') as utm_campaign
FROM `demeterics.touchpoints.touchpoints`
WHERE _PARTITIONTIME >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 DAY)
```

**After conversion** (using Payload):
```sql
SELECT
  Time,
  Category,
  Action,
  Payload.utm_source,
  Payload.utm_campaign
FROM `demeterics.touchpoints.touchpoints`
WHERE _PARTITIONTIME >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 DAY)
  AND Payload IS NOT NULL
```

## Table Schema

The relevant columns in the touchpoints table:

```sql
CREATE TABLE `demeterics.touchpoints.touchpoints` (
  Time TIMESTAMP,
  Category STRING,
  Action STRING,
  Label STRING,
  Referer STRING,
  Path STRING,
  Host STRING,
  RemoteAddr STRING,
  UserAgent STRING,
  PayloadString STRING,  -- Raw JSON string for ingestion
  Payload JSON           -- Native JSON for queries (populated manually)
)
PARTITION BY DATE(Time)
CLUSTER BY Host, Category;
```

## Best Practices

1. **Run conversion periodically**: Convert `PayloadString` to `Payload` when you need to query the data.

2. **Don't modify ingestion code**: The streaming insert must use `PayloadString` (STRING) for reliability.

3. **Use SAFE.PARSE_JSON**: This prevents conversion errors from stopping the entire UPDATE.

4. **Partition filtering**: Always include `_PARTITIONTIME` filter to avoid scanning the entire table.

## Troubleshooting

### Conversion fails for some rows

If `SAFE.PARSE_JSON` returns NULL for valid-looking JSON:
1. Check for invalid UTF-8 characters
2. Check for unescaped control characters
3. View problematic rows:
   ```sql
   SELECT PayloadString
   FROM `demeterics.touchpoints.touchpoints`
   WHERE SAFE.PARSE_JSON(PayloadString) IS NULL
     AND PayloadString IS NOT NULL
     AND PayloadString != '{}'
   LIMIT 10
   ```

### Empty payloads

Empty payloads are stored as `'{}'`. The conversion query excludes these to keep `Payload` NULL for empty data.

## Related Files

- `track/bigquery_store.go`: `touchPointInsertRequest()` function builds the insert request
- `track/touchpoint.go`: `TouchPointEvent` struct definition
- `gcp/bigquery.go`: Low-level BigQuery operations
