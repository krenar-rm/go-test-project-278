-- +goose Up
CREATE TABLE IF NOT EXISTS link_visits (
	id BIGSERIAL PRIMARY KEY,
	link_id INT,
	ip VARCHAR(45),
	user_agent VARCHAR(255),
	referer TEXT,
	status INT,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS link_visits;
