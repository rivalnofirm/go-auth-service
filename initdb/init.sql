-- Table: user_auth
CREATE TABLE user_auth (
                           id BIGSERIAL PRIMARY KEY,
                           email VARCHAR(100) NOT NULL UNIQUE,
                           password VARCHAR(100) NOT NULL,
                           created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP,
                           deleted_at TIMESTAMP
);

-- Table: user_type
CREATE TABLE user_type (
                           id BIGSERIAL PRIMARY KEY,
                           type VARCHAR(25) NOT NULL
);

-- Table: user_detail
CREATE TABLE user_detail (
                             id BIGSERIAL PRIMARY KEY,
                             user_id BIGINT NOT NULL UNIQUE,
                             user_type_id BIGINT NOT NULL,
                             phone VARCHAR(20) UNIQUE,
                             first_name VARCHAR(50) NOT NULL,
                             last_name VARCHAR(50),
                             picture VARCHAR(250),
                             birth_date DATE,
                             gender VARCHAR(10),
                             verified BOOLEAN NOT NULL DEFAULT FALSE,
                             created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             updated_at TIMESTAMP,
                             deleted_at TIMESTAMP,
                             CONSTRAINT fk_user_detail_user FOREIGN KEY(user_id) REFERENCES user_auth(id) ON DELETE CASCADE,
                             CONSTRAINT fk_user_detail_user_type FOREIGN KEY(user_type_id) REFERENCES user_type(id) ON DELETE SET NULL
);

-- Table: user_login_history
CREATE TABLE user_login_history (
                                    id BIGSERIAL PRIMARY KEY,
                                    user_id BIGINT NOT NULL,
                                    login_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                    ip_address VARCHAR(45) NOT NULL,
                                    user_agent TEXT NOT NULL,
                                    logout_time TIMESTAMP,
                                    logout_reason VARCHAR(50),
                                    CONSTRAINT fk_login_history_user FOREIGN KEY(user_id) REFERENCES user_auth(id) ON DELETE CASCADE
);

-- Table: user_refresh_token
CREATE TABLE user_refresh_token (
                                    id BIGSERIAL PRIMARY KEY,
                                    user_id BIGINT NOT NULL,
                                    refresh_token_hash VARCHAR(255) NOT NULL,
                                    expires_at TIMESTAMP NOT NULL,
                                    user_agent TEXT NOT NULL,
                                    is_active BOOLEAN NOT NULL DEFAULT TRUE,
                                    CONSTRAINT fk_refresh_token_user FOREIGN KEY(user_id) REFERENCES user_auth(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_user_auth_id ON user_auth(id);
CREATE INDEX idx_user_detail_id ON user_detail(id);
CREATE INDEX idx_user_login_history_user_id ON user_login_history(user_id);
CREATE INDEX idx_user_refresh_token_user_id ON user_refresh_token(user_id);

-- Seed data for user_type
INSERT INTO public.user_type (type) VALUES ('super admin');
INSERT INTO public.user_type (type) VALUES ('admin');
INSERT INTO public.user_type (type) VALUES ('user');
