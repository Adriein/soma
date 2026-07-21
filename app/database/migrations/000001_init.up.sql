/*
================================================================================
TABLES
================================================================================
*/

CREATE TABLE IF NOT EXISTS so_users (
    sou_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sou_telegram_chat_id INT NOT NULL UNIQUE,
    sou_name VARCHAR,
    sou_token VARCHAR,
    sou_token_secret VARCHAR,
    sou_token_verifier INT,
    sou_date_add TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    sou_date_upd TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS so_assessment (
    soa_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    soa_text VARCHAR NOT NULL,
    soa_date_add TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    soa_date_upd TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
)