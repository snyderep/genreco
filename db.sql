CREATE TABLE user_product_views (
    account_id  INTEGER NOT NULL,
    monetate_id TEXT    NOT NULL,
    pid         TEXT    NOT NULL,
    count       INTEGER NOT NULL,
    PRIMARY KEY (account_id, monetate_id, pid)
);
CREATE INDEX user_product_views_ak1 ON user_product_views (account_id, pid);

CREATE TABLE user_product_purchases (
    account_id  INTEGER NOT NULL,
    monetate_id TEXT    NOT NULL,
    pid         TEXT    NOT NULL,
    count       INTEGER NOT NULL,
    PRIMARY KEY (account_id, monetate_id, pid)
);
CREATE INDEX user_product_purchases_ak1 ON user_product_purchases (account_id, pid);

CREATE TABLE product_conversion (
    account_id      INTEGER NOT NULL,
    pid             TEXT    NOT NULL,
    conversion_rate FLOAT   NOT NULL,
    PRIMARY KEY (account_id, pid)
);

CREATE TABLE product (
    account_id  INTEGER        NOT NULL,
    pid         TEXT           NOT NULL,
    name        TEXT           NULL,
    product_url TEXT           NULL,
    image_url   TEXT           NULL,
    unit_cost   NUMERIC(12, 2) NULL,
    unit_price  NUMERIC(12, 2) NULL,
    margin      NUMERIC(12, 2) NULL,
    margin_rate FLOAT          NULL,
    PRIMARY KEY (account_id, pid)
);
