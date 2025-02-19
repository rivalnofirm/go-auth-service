-- Create sequences
CREATE SEQUENCE user_auth_id_seq;
CREATE SEQUENCE user_type_id_seq;
CREATE SEQUENCE user_detail_id_seq;
CREATE SEQUENCE user_login_history_id_seq;
-- CREATE SEQUENCE user_refresh_token_id_seq;

-- Table: user_auth
CREATE TABLE user_auth (
                           id INTEGER PRIMARY KEY DEFAULT nextval('user_auth_id_seq'),
                           email VARCHAR(100) NOT NULL UNIQUE,
                           password VARCHAR(100) NOT NULL,
                           created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP,
                           deleted_at TIMESTAMP
);

-- Table: user_type
CREATE TABLE user_type (
                           id INTEGER PRIMARY KEY DEFAULT nextval('user_type_id_seq'),
                           type VARCHAR(25) NOT NULL
);

-- Table: user_detail
CREATE TABLE user_detail (
                             id INTEGER PRIMARY KEY DEFAULT nextval('user_detail_id_seq'),
                             user_id INTEGER NOT NULL UNIQUE,
                             user_type_id INTEGER NOT NULL,
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
                                    id INTEGER PRIMARY KEY DEFAULT nextval('user_login_history_id_seq'),
                                    user_id INTEGER NOT NULL,
                                    login_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                    ip_address VARCHAR(45) NOT NULL,
                                    user_agent TEXT NOT NULL,
                                    logout_time TIMESTAMP,
                                    logout_reason VARCHAR(50),
                                    CONSTRAINT fk_login_history_user FOREIGN KEY(user_id) REFERENCES user_auth(id) ON DELETE CASCADE
);

-- Table: user_refresh_tokens
-- CREATE TABLE user_refresh_tokens (
--                                      id INTEGER PRIMARY KEY DEFAULT nextval('user_refresh_token_id_seq'),
--                                      user_id INTEGER NOT NULL,
--                                      refresh_token_hash VARCHAR(255) NOT NULL,
--                                      expires_at TIMESTAMP NOT NULL,
--                                      created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--                                      revoked BOOLEAN NOT NULL DEFAULT FALSE,
--                                      CONSTRAINT fk_refresh_token_user FOREIGN KEY(user_id) REFERENCES user_auth(id) ON DELETE CASCADE
-- );

-- Indexes
CREATE INDEX idx_user_auth_id ON user_auth(id);
CREATE INDEX idx_user_detail_id ON user_detail(id);
CREATE INDEX idx_user_login_history_user_id ON user_login_history(user_id);
-- CREATE INDEX idx_user_refresh_tokens_user_id ON user_refresh_tokens(user_id);

-- Seed data for user_type
INSERT INTO public.user_type (id, type) VALUES (DEFAULT, 'super admin');
INSERT INTO public.user_type (id, type) VALUES (DEFAULT, 'admin');
INSERT INTO public.user_type (id, type) VALUES (DEFAULT, 'user');
