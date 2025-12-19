-- +goose Up
CREATE TABLE ns_invoice_sales_tax_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    type TEXT,
    date DATE,
    date_due DATE,
    document_number TEXT,
    name TEXT,
    memo TEXT,
    item TEXT,
    qty NUMERIC,
    contract_quantity NUMERIC,
    unit_price NUMERIC,
    amount NUMERIC,
    start_date_line DATE,
    end_date_line_level DATE,
    account TEXT,
    salesforce_opportunity_id TEXT,
    salesforce_pricebook_id TEXT,
    item_internal_id TEXT,
    entity_internal_id TEXT,
    shipping_address_city TEXT,
    shipping_address_state TEXT,
    shipping_address_country TEXT,

    CONSTRAINT ns_invoice_sales_tax_items_all_fields_unique UNIQUE (
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
);

-- +goose Down
DROP TABLE IF EXISTS ns_invoice_sales_tax_items;
