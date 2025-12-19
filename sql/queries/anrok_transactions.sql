-- name: InsertAnrokTransactions :exec

INSERT INTO anrok_transactions (
    transaction_id,
    customer_id,
    customer_name,
    overall_vat_id_status,
    valid_vat_ids,
    other_vat_ids,
    invoice_date,
    tax_date,
    transaction_currency,
    sales_amount,
    exempt_reason,
    tax_amount,
    invoice_amount,
    void,
    customer_address_line_1,
    customer_address_city,
    customer_address_region,
    customer_address_postal_code,
    customer_address_country,
    customer_country_code,
    jurisdictions,
    jurisdiction_ids,
    return_ids
)
values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
)
ON CONFLICT ON CONSTRAINT anrok_transactions_all_fields_unique
DO NOTHING;

-- name: UpdateAnrokTransactions :one

UPDATE anrok_transactions
    SET

    transaction_id = $2,
    customer_id = $3,
    customer_name = $4,
    overall_vat_id_status = $5,
    valid_vat_ids = $6,
    other_vat_ids = $7,
    invoice_date = $8,
    tax_date = $9,
    transaction_currency = $10,
    sales_amount = $11,
    exempt_reason = $12,
    tax_amount = $13,
    invoice_amount = $14,
    void = $15,
    customer_address_line_1 = $16,
    customer_address_city = $17,
    customer_address_region = $18,
    customer_address_postal_code = $19,
    customer_address_country = $20,
    customer_country_code = $21,
    jurisdictions = $22,
    jurisdiction_ids = $23,
    return_ids = $24

WHERE id = $1
RETURNING id;


-- name: RemoveAnrokTransactions :exec
DELETE FROM anrok_transactions
WHERE
    ($1 IS NULL OR transaction_id = $1)
    AND
    ($2 IS NULL OR id = $2);

-- name: ResetAnrokTransactions :exec
DELETE FROM anrok_transactions;
