-- name: InsertSfdcOppLineItems :exec
INSERT INTO sfdc_opp_line_items (
    opportunity_id,
    opportunity_product_casesafe_id,
    opportunity_name,
    account_name,
    close_date,
    booked_date,
    fiscal_period,
    payment_schedule,
    payment_due,
    contract_start_date,
    contract_end_date,
    term_in_months_deprecated,
    product_name,
    deployment_type,
    amount,
    quantity,
    list_price,
    sales_price,
    total_price,
    start_date,
    end_date,
    term_in_months,
    product_code,
    total_amount_due_customer,
    total_amount_due_partner,
    active_product
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
)
ON CONFLICT ON CONSTRAINT sfdc_opp_line_items_all_fields_unique
DO NOTHING;

-- name: UpdateSfdcOppLineItemById :one
UPDATE sfdc_opp_line_items
    SET

    opportunity_id = $2,
    opportunity_product_casesafe_id = $3,
    opportunity_name = $4,
    account_name = $5,
    close_date = $6,
    booked_date = $7,
    fiscal_period = $8,
    payment_schedule = $9,
    payment_due = $10,
    contract_start_date = $11,
    contract_end_date = $12,
    term_in_months_deprecated = $13,
    product_name = $14,
    deployment_type = $15,
    amount = $16,
    quantity = $17,
    list_price = $18,
    sales_price = $19,
    total_price = $20,
    start_date = $21,
    end_date = $22,
    term_in_months = $23,
    product_code = $24,
    total_amount_due_customer = $25,
    total_amount_due_partner = $26,
    active_product = $27
WHERE id = $1
RETURNING id;

-- name: RemoveOpp :exec
DELETE FROM sfdc_opp_line_items
WHERE opportunity_id = $1;

-- name: ResetSfdcOppLineItems :exec
DELETE FROM sfdc_opp_line_items;
