-- name: InsertNsInvoiceSalesTaxItems :exec

INSERT INTO ns_invoice_sales_tax_items (
    type,
    date,
    date_due,
    document_number,
    name,
    memo,
    item,
    qty,
    contract_quantity,
    unit_price,
    amount,
    start_date_line,
    end_date_line_level,
    account,
    salesforce_opportunity_id,
    salesforce_pricebook_id,
    item_internal_id,
    entity_internal_id,
    shipping_address_city,
    shipping_address_state,
    shipping_address_country
)
values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
)
ON CONFLICT ON CONSTRAINT ns_invoice_sales_tax_items_all_fields_unique
DO NOTHING;


-- name: UpdateNsInvoiceSalesTaxItems :one

UPDATE ns_invoice_sales_tax_items
    SET
    type = $2,
    date = $3,
    date_due = $4,
    document_number = $5,
    name = $6,
    memo = $7,
    item = $8,
    qty = $9,
    contract_quantity = $10,
    unit_price = $11,
    amount = $12,
    start_date_line = $13,
    end_date_line_level = $14,
    account = $15,
    salesforce_opportunity_id = $16,
    salesforce_pricebook_id = $17,
    item_internal_id = $18,
    entity_internal_id = $19,
    shipping_address_city = $20,
    shipping_address_state = $21,
    shipping_address_country = $22

WHERE id = $1
RETURNING id;

-- name: RemoveNsInvoiceSalesTaxItems :exec
DELETE FROM ns_invoice_sales_tax_items
WHERE
    ($1 IS NULL OR salesforce_opportunity_id = $1)
    AND
    ($2 IS NULL OR document_number = $2);

-- name: ResetNsInvoiceSalesTaxItems :exec
DELETE FROM ns_invoice_sales_tax_items;
