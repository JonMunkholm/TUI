-- +goose Up
CREATE TABLE ns_so_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    salesforce_opportunity_id TEXT,
    salesforce_opportunity_line_id TEXT,
    customer_project TEXT NOT NULL,
    document_number TEXT,
    document_date DATE,
    start_date DATE,
    end_date DATE,
    item_name TEXT,
    item_display_name TEXT,
    line_start_date DATE,
    line_end_date DATE,
    quantity NUMERIC,
    contract_quantity NUMERIC,
    unit_price NUMERIC,
    total_amount_due_partner NUMERIC,
    amount_gross NUMERIC,
    terms_days_till_net_due NUMERIC,

    CONSTRAINT ns_so_line_items_all_fields_unique UNIQUE (
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
);


-- +goose Down
DROP TABLE IF EXISTS ns_so_line_items;
