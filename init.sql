CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subsidy_tasks (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    farmer_id               VARCHAR(100)   NOT NULL,
    region_code             VARCHAR(50)    NOT NULL,
    region_priority         INT            NOT NULL CHECK (region_priority BETWEEN 1 AND 5),
    amount_requested        DECIMAL(15, 2) NOT NULL,
    amount_norm             DECIMAL(15, 2) NOT NULL,
    application_date        TIMESTAMP      NOT NULL,
    crop_type               VARCHAR(100)   NOT NULL,
    farm_size_ha            DECIMAL(10, 2) NOT NULL,
    previous_subsidies_count INT           NOT NULL DEFAULT 0,
    created_at              TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subsidy_results (
    id           UUID           PRIMARY KEY REFERENCES subsidy_tasks (id),
    score        DECIMAL(5, 2)  NOT NULL,
    shap_values  JSONB          NOT NULL DEFAULT '{}',
    flags        TEXT[]         NOT NULL DEFAULT '{}',
    processed_at TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_results_score ON subsidy_results (score DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_farmer  ON subsidy_tasks  (farmer_id);
CREATE INDEX IF NOT EXISTS idx_tasks_region  ON subsidy_tasks  (region_code);
