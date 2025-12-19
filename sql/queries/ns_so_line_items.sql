-- name: InsertNsSoLineItems :exec

INSERT INTO ns_so_line_items (
    salesforce_opportunity_id,
    salesforce_opportunity_line_id,
    customer_project,
    document_number,
    document_date,
    start_date,
    end_date,
    item_name,
    item_display_name,
    line_start_date,
    line_end_date,
    quantity,
    contract_quantity,
    unit_price,
    total_amount_due_partner,
    amount_gross,
    terms_days_till_net_due
)
VALUES(
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
ON CONFLICT ON CONSTRAINT ns_so_line_items_all_fields_unique
DO NOTHING;

-- name: UpdateNsSoLineItems :one

UPDATE ns_so_line_items
    SET
    salesforce_opportunity_id = $2,
    salesforce_opportunity_line_id = $3,
    customer_project = $4,
    document_number = $5,
    document_date = $6,
    start_date = $7,
    end_date = $8,
    item_name = $9,
    item_display_name = $10,
    line_start_date = $11,
    line_end_date = $12,
    quantity = $13,
    contract_quantity = $14,
    unit_price = $15,
    total_amount_due_partner = $16,
    amount_gross = $17,
    terms_days_till_net_due = $18
WHERE id = $1
RETURNING id;


-- name: RemoveSo :exec
DELETE FROM ns_so_line_items
WHERE
    ($1 IS NULL OR salesforce_opportunity_id = $1)
    AND
    ($2 IS NULL OR document_number = $2);

-- name: ResetNsSoLineItems :exec
DELETE FROM ns_so_line_items;
