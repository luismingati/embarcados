-- name: CreateVolume :one
INSERT INTO volumes (
  value, created_at
) VALUES (
  $1, $2
)
RETURNING *;

-- name: CreateFlowRate :one
INSERT INTO flow_rates (
  value, created_at
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetVolumes :many
SELECT * FROM volumes
ORDER BY created_at
LIMIT $1;

-- name: GetVolumesByPeriod :many
SELECT
  date_trunc($1, created_at)::TIMESTAMP AS period,
  SUM(value)::float AS total_value
FROM volumes
WHERE created_at <= NOW() AT TIME ZONE 'America/Recife'
GROUP BY period
ORDER BY period DESC
LIMIT $2;

-- name: GetTotalVolumeByPeriod :one
SELECT
  SUM(value)::float AS total_value
FROM volumes
WHERE created_at >= date_trunc($1, NOW() AT TIME ZONE 'America/Recife');