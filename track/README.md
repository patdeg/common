# Track Package - BigQuery Analytics

This package provides helpers for tracking visits, events, and marketing touchpoints in BigQuery.

## TouchPoints and BigQuery JSON Columns

### Overview

The `touchpoints` table uses BigQuery's native **JSON type** for the `Payload` column. This enables powerful SQL queries using dot-notation to access nested fields:

```sql
-- Query UTM parameters directly
SELECT
    Payload.utm_source,
    Payload.utm_campaign,
    Payload.utm_medium,
    COUNT(*) as hits
FROM `demeterics.touchpoints.touchpoints`
WHERE DATE(Time) = CURRENT_DATE()
GROUP BY 1, 2, 3

-- Filter by nested fields
SELECT * FROM `demeterics.touchpoints.touchpoints`
WHERE Payload.utm_medium = "cpc"
  AND Payload.utm_campaign LIKE "%spring%"
```

### Critical Implementation Details

#### BigQuery JSON Type Requirements

BigQuery's streaming insert API has specific requirements for JSON columns:

| Approach | Works? | Example |
|----------|--------|---------|
| Parsed Go map | **YES** | `"Payload": map[string]interface{}{"utm_source": "google"}` |
| JSON string | **NO** | `"Payload": "{\"utm_source\": \"google\"}"` |

The `touchPointInsertRequest()` function handles this automatically by parsing `PayloadJSON` (a string) into a `map[string]interface{}` before insertion.

#### How It Works

1. **Application code** creates a `TouchPointEvent` with `PayloadJSON` as a JSON string:
   ```go
   event := &track.TouchPointEvent{
       Category:    "landing",
       Action:      "view",
       PayloadJSON: `{"utm_source": "google", "utm_campaign": "spring2025"}`,
   }
   ```

2. **`touchPointInsertRequest()`** parses the JSON string into a Go map:
   ```go
   var payloadMap map[string]interface{}
   json.Unmarshal([]byte(tp.PayloadJSON), &payloadMap)
   // payloadMap = map[string]interface{}{"utm_source": "google", "utm_campaign": "spring2025"}
   ```

3. **BigQuery streaming API** receives the map and serializes it as JSON internally.

4. **SQL queries** can now use dot-notation: `SELECT Payload.utm_source`

### Table Schema

The `touchpoints` table is created with:

```sql
CREATE TABLE touchpoints (
    Time TIMESTAMP,           -- Partitioned by day
    Category STRING,          -- Event category (e.g., 'landing', 'campaign')
    Action STRING,            -- Event action (e.g., 'view', 'cta_click')
    Label STRING,             -- Optional event label
    Referer STRING,           -- HTTP Referer header
    Path STRING,              -- Request path
    Host STRING,              -- HTTP host header
    RemoteAddr STRING,        -- Client IP address
    UserAgent STRING,         -- User-Agent header
    Payload JSON              -- Queryable JSON payload
)
PARTITION BY DATE(Time)
```

### Common Payload Fields

The `Payload` JSON column typically contains:

| Field | Description | Example |
|-------|-------------|---------|
| `utm_source` | Traffic source | `"google"`, `"newsletter"` |
| `utm_medium` | Marketing medium | `"cpc"`, `"email"`, `"organic"` |
| `utm_campaign` | Campaign name | `"spring2025"`, `"black_friday"` |
| `utm_term` | Paid search keyword | `"ai analytics"` |
| `utm_content` | Ad content variant | `"banner_a"`, `"text_link"` |
| `page_title` | Page title | `"Pricing - Demeterics"` |
| `button_text` | CTA button clicked | `"Get Started"` |

### Error Handling

If `PayloadJSON` is empty or contains invalid JSON:
- An empty map `{}` is inserted (row is not rejected)
- A warning is logged for debugging
- The other fields (Category, Action, etc.) are still recorded

### Recreating the Table

If you need to recreate the table with the correct schema:

```bash
# 1. Delete existing table (only if empty or data can be lost)
bq rm -f demeterics:touchpoints.touchpoints

# 2. Create with correct JSON schema
bq mk --table \
  --time_partitioning_field=Time \
  --time_partitioning_type=DAY \
  --description="Marketing touch point events with queryable JSON payload" \
  demeterics:touchpoints.touchpoints \
  'Time:TIMESTAMP,Category:STRING,Action:STRING,Label:STRING,Referer:STRING,Path:STRING,Host:STRING,RemoteAddr:STRING,UserAgent:STRING,Payload:JSON'
```

Or let the code auto-create it via `createTouchpointsTableInBigQuery()`.

### Testing Queries

After inserting data, verify the JSON column works:

```sql
-- Check recent touchpoints with payload fields
SELECT
    Time,
    Category,
    Action,
    Payload,
    JSON_VALUE(Payload, '$.utm_source') as utm_source_alt,  -- Alternative syntax
    Payload.utm_source as utm_source_dot                     -- Dot notation
FROM `demeterics.touchpoints.touchpoints`
ORDER BY Time DESC
LIMIT 10
```

## Version History

- **v1.21.0**: Fixed BigQuery JSON column handling - Payload is now parsed from JSON string to map before streaming insert. Added comprehensive documentation.
- **v1.20.0**: Changed Payload to STRING type (incorrect approach).
- **v1.19.0**: Initial touchpoints implementation with JSON type but string value (broken).
