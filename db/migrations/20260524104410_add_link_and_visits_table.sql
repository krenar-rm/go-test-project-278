-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS links (
	id BIGSERIAL PRIMARY KEY,
	original_url TEXT,
	short_name TEXT UNIQUE,
	short_url TEXT,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS link_visits (
	id BIGSERIAL PRIMARY KEY,
	link_id BIGINT NOT NULL,
	ip VARCHAR(45),
	user_agent VARCHAR(255),
	referer VARCHAR(500),
	status INT,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (link_id) REFERENCES links(id)
	ON DELETE CASCADE
	ON UPDATE CASCADE
);	
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS link_visits;
-- +goose StatementEnd
