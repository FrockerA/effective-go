CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
                                             id           UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
                                             service_name VARCHAR(255) NOT NULL,
                                             price        INTEGER      NOT NULL,
                                             user_id      UUID         NOT NULL,
                                             start_date   DATE         NOT NULL,
                                             end_date     DATE,
                                             created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
                                             updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);