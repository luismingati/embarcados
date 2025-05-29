CREATE TABLE volumes (
    id BIGSERIAL PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE flow_rates (
    id BIGSERIAL PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

---- create above / drop below ----

drop table volumes;
drop table flow_rates;