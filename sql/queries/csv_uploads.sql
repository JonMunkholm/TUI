-- name: InsertCsvUpload :exec
INSERT INTO csv_uploads (name, action, uploaded_at)
VALUES ($1, $2, NOW())
ON CONFLICT (name) DO NOTHING;

-- name: GetCsvUploadsByDateRange :many
SELECT *
FROM csv_uploads
WHERE uploaded_at BETWEEN $1 AND $2
ORDER BY uploaded_at ASC;

-- name: GetCsvUpload :many
SELECT *
FROM csv_uploads
WHERE name = $1;

-- name: ResetCsvUpload :exec
DELETE FROM csv_uploads;
